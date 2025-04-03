package repo

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"reflect"
)

var (
	ErrOptimisticLockConflict = errors.New("optimistic lock conflict")
)

type Model struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt int64
	UpdatedAt int64
}

func CheckFound[T any](val T, err error) (T, bool, error) {
	if err == nil {
		return val, true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return val, false, nil
	}
	return val, false, err
}

func CheckUpdate(tx *gorm.DB) error {
	return CheckUpdateWithRowsAffected(tx, 1)
}

func CheckUpdateWithRowsAffected(tx *gorm.DB, rowsAffected int) error {
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected != int64(rowsAffected) {
		return ErrOptimisticLockConflict
	}
	return nil
}

type BaseRepo[T any] struct {
	DB *gorm.DB
}

func (r *BaseRepo[T]) Save(item T) error {
	return r.DB.Create(item).Error
}

func (r *BaseRepo[T]) PrimaryKeySaveTabler(item schema.Tabler) error {
	return r.DB.Table(item.TableName()).Save(item).Error
}

func (r *BaseRepo[T]) SaveAll(items []T) error {
	return r.DB.Create(items).Error
}

func (r *BaseRepo[T]) Update(item T) error {
	tx := r.DB.Updates(item)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected != 1 {
		return ErrOptimisticLockConflict
	}
	return nil
}

func (r *BaseRepo[T]) VersionSave(v schema.Tabler) error {
	version := getVersion(v)
	var tx *gorm.DB

	if version > 0 {
		// For records with version > 0, use optimistic locking
		tx = r.DB.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			Where: clause.Where{
				Exprs: []clause.Expression{
					clause.Expr{SQL: v.TableName() + ".version = excluded.version - 1"},
				},
			},
			UpdateAll: true,
		}).Create(v)
	} else {
		// For records with version = 0, update without version check
		tx = r.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			UpdateAll: true,
		}).Create(v)
	}
	return CheckUpdateWithRowsAffected(tx, 1)
}

func (r *BaseRepo[T]) BatchVersionSave(values []T, tableName string) error {
	// Split records into two groups: save or update by version
	var versionedValues []T
	var unversionedValues []T

	for _, v := range values {
		if getVersion(v) > 0 {
			versionedValues = append(versionedValues, v)
		} else {
			unversionedValues = append(unversionedValues, v)
		}
	}

	var tx *gorm.DB
	var err error

	// Process records with version > 0
	if len(versionedValues) > 0 {
		tx = r.DB.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			Where: clause.Where{
				Exprs: []clause.Expression{
					clause.Expr{SQL: tableName + ".version = excluded.version - 1"},
				},
			},
			UpdateAll: true,
		}).Create(&versionedValues)
		if err = CheckUpdateWithRowsAffected(tx, len(versionedValues)); err != nil {
			return err
		}
	}

	// Process records with version = 0
	if len(unversionedValues) > 0 {
		tx = r.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			UpdateAll: true,
		}).Create(&unversionedValues)
		if err = CheckUpdateWithRowsAffected(tx, len(unversionedValues)); err != nil {
			return err
		}
	}

	return nil
}

// getVersion gets the version field value of a struct
func getVersion(v interface{}) int {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	versionField := val.FieldByName("Version")
	if !versionField.IsValid() {
		return 0
	}
	return int(versionField.Int())
}

func (r *BaseRepo[T]) DeleteByID(id uint64) error {
	var item T
	tx := r.DB.Where("id = ?", id).Delete(&item)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected != 1 {
		return ErrOptimisticLockConflict
	}
	return nil
}
