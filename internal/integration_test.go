package internal_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Noziop/s4v3my4ss/internal/backup"
	"github.com/Noziop/s4v3my4ss/internal/restore"
	"github.com/Noziop/s4v3my4ss/pkg/common"
)

// TestBackupAndRestore tests the backup and restore process end-to-end
func TestBackupAndRestore(t *testing.T) {
	// Skip this test when running short tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Initialize application
	err := common.InitApp()
	if err != nil {
		t.Fatalf("Failed to initialize application: %v", err)
	}

	// Create temporary source directory
	sourceDir, err := os.MkdirTemp("", "s4v3my4ss_test_source")
	if err != nil {
		t.Fatalf("Failed to create temp source dir: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	// Create temporary destination directory
	destDir, err := os.MkdirTemp("", "s4v3my4ss_test_dest")
	if err != nil {
		t.Fatalf("Failed to create temp dest dir: %v", err)
	}
	defer os.RemoveAll(destDir)

	// Create temporary restore directory
	restoreDir, err := os.MkdirTemp("", "s4v3my4ss_test_restore")
	if err != nil {
		t.Fatalf("Failed to create temp restore dir: %v", err)
	}
	defer os.RemoveAll(restoreDir)

	// Create some test files in source directory
	testFiles := []string{"file1.txt", "file2.txt", "subdir/file3.txt"}
	testContent := "This is test content"

	// Create subdirectory
	err = os.MkdirAll(filepath.Join(sourceDir, "subdir"), 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create test files
	for _, file := range testFiles {
		filePath := filepath.Join(sourceDir, file)
		err = os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory for %s: %v", file, err)
		}
		err = ioutil.WriteFile(filePath, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Set backup destination in configuration
	common.AppConfig.BackupDestination = destDir
	
	// Create backup configuration
	backupCfg := backup.BackupConfig{
		SourcePath:   sourceDir,
		Name:         "integration-test",
		Compression:  true,
		ExcludeDirs:  []string{},
		ExcludeFiles: []string{},
		Incremental:  false,
	}

	// Perform backup
	err = backup.CreateBackup(backupCfg)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Get backup info
	backups, err := common.ListBackups()
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}

	if len(backups) == 0 {
		t.Fatalf("No backups found after CreateBackup")
	}

	// Find our test backup
	var backupID string
	for _, b := range backups {
		if b.Name == "integration-test" {
			backupID = b.ID
			break
		}
	}

	if backupID == "" {
		t.Fatalf("Could not find integration-test backup in the list")
	}

	// Allow time for backup to complete
	time.Sleep(1 * time.Second)

	// Restore the backup
	err = restore.RestoreBackup(backupID, restoreDir)
	if err != nil {
		t.Fatalf("Failed to restore backup: %v", err)
	}

	// Verify restored files
	for _, file := range testFiles {
		restoredFile := filepath.Join(restoreDir, file)
		if _, err := os.Stat(restoredFile); os.IsNotExist(err) {
			t.Errorf("Restored file %s doesn't exist", file)
			continue
		}

		content, err := ioutil.ReadFile(restoredFile)
		if err != nil {
			t.Errorf("Failed to read restored file %s: %v", file, err)
			continue
		}

		if string(content) != testContent {
			t.Errorf("Restored file %s has wrong content. Expected: %s, Got: %s", 
				file, testContent, string(content))
		}
	}
}