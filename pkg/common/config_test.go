package common

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadConfig(t *testing.T) {
	// Create a temporary config file for testing
	tempDir, err := os.MkdirTemp("", "s4v3my4ss_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set the config file to our temp location
	origConfigFile := ConfigFile
	ConfigFile = filepath.Join(tempDir, "config.json")
	defer func() { ConfigFile = origConfigFile }()
	
	// Create test config
	testConfig := Config{
		BackupDirs: []BackupConfig{
			{
				Name:        "test-backup",
				SourcePath:  "/tmp/source",
				Compression: true,
				ExcludeDirs: []string{"node_modules", ".git"},
				Interval:    60,
			},
		},
		BackupDestination: "/tmp/backups",
		RetentionPolicy: RetentionPolicy{
			KeepDaily:   7,
			KeepWeekly:  4,
			KeepMonthly: 3,
		},
	}

	// Save the config
	if err := SaveConfig(testConfig); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Load the config
	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify the loaded config matches what we saved
	if len(loadedConfig.BackupDirs) != len(testConfig.BackupDirs) {
		t.Errorf("Expected %d backup dirs, got %d", 
			len(testConfig.BackupDirs), len(loadedConfig.BackupDirs))
	}

	if loadedConfig.BackupDirs[0].Name != testConfig.BackupDirs[0].Name {
		t.Errorf("Expected backup name %s, got %s", 
			testConfig.BackupDirs[0].Name, loadedConfig.BackupDirs[0].Name)
	}

	if loadedConfig.BackupDestination != testConfig.BackupDestination {
		t.Errorf("Expected destination %s, got %s", 
			testConfig.BackupDestination, loadedConfig.BackupDestination)
	}
}

func TestGetBackupConfig(t *testing.T) {
	// Setup test config
	AppConfig = Config{
		BackupDirs: []BackupConfig{
			{
				Name:       "test1",
				SourcePath: "/path/to/test1",
			},
			{
				Name:       "test2",
				SourcePath: "/path/to/test2",
			},
		},
	}

	// Test finding existing config
	config, found := GetBackupConfig("test1")
	if !found {
		t.Errorf("GetBackupConfig didn't find existing config 'test1'")
	}
	if config.Name != "test1" || config.SourcePath != "/path/to/test1" {
		t.Errorf("GetBackupConfig returned incorrect config")
	}

	// Test finding non-existent config
	_, found = GetBackupConfig("non-existent")
	if found {
		t.Errorf("GetBackupConfig unexpectedly found non-existent config")
	}
}