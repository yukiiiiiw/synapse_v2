package repo

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
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
	tx := r.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		Where: clause.Where{
			Exprs: []clause.Expression{
				clause.Expr{SQL: v.TableName() + ".version = excluded.version - 1"},
			},
		},
		UpdateAll: true,
	}).Create(v)
	return CheckUpdateWithRowsAffected(tx, 1)
}

func (r *BaseRepo[T]) BatchVersionSave(values []T, tableName string) error {
	tx := r.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		Where: clause.Where{
			Exprs: []clause.Expression{
				clause.Expr{SQL: tableName + ".version = excluded.version - 1"},
			},
		},
		UpdateAll: true,
	}).Create(&values)
	return CheckUpdateWithRowsAffected(tx, len(values)*2)
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
