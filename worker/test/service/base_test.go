package service

import (
	"context"
	"os"
	"path/filepath"
	"synapse/common"
	commoncfg "synapse/common/config"
	idgenerator "synapse/common/id-generator"
	"synapse/common/log"
	"synapse/worker/config"
	"testing"
)

// SetupTestEnv initializes the test environment
func SetupTestEnv(t *testing.T) {
	// Set environment variable
	os.Setenv("PROFILE", "local")

	// Get project root directory
	projectRoot, err := getProjectRoot()
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	// Change working directory to project root
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Failed to change working directory: %v", err)
	}

	// Read configuration
	_, err = commoncfg.ReadConfig(common.ServiceWorker, &config.Config)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}
	t.Logf("Config loaded successfully ")

	// Initialize database connection
	if err := commoncfg.InitDatasource(context.Background(), log.Log, &config.Config.Datasource); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	t.Logf("Database configuration: %+v", config.Config.Datasource)
	t.Logf("Database connection initialized successfully: %+v", commoncfg.DB)

	// Set the database connection in worker config
	config.DB = commoncfg.DB

	// Initialize Redis client
	config.Redis = commoncfg.InitRedis(&config.Config.Redis)

	t.Logf("Redis client initialized successfully")

	// Initialize Snowflake
	if err := idgenerator.InitMultiSnowflakeInstances(
		context.Background(),
		config.Redis,
		config.SnowflakeNodeIDRedisKeyForNextId,
	); err != nil {
		t.Fatalf("Failed to initialize snowflake: %v", err)
	}
}

// getProjectRoot returns the project root directory
func getProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
