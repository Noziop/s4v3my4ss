package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BackupInfo contient les informations sur une sauvegarde
type BackupInfo struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	SourcePath    string             `json:"sourcePath"`
	BackupPath    string             `json:"backupPath"`
	Time          time.Time          `json:"time"`
	Size          int64              `json:"size"`
	IsIncremental bool               `json:"isIncremental"`
	Compression   bool               `json:"compression"`
	Encrypted     bool               `json:"encrypted,omitempty"`
	RemoteServer  *RsyncServerConfig `json:"remoteServer,omitempty"` // Serveur rsync distant si applicable
	DestinationName string `json:"destinationName,omitempty"` // Nom de la destination utilisée
}

// SaveBackupInfo sauvegarde les métadonnées d'une sauvegarde
func SaveBackupInfo(info BackupInfo) error {
	filename := filepath.Join(BackupInfoDir, info.ID+".json")
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		LogError("Impossible de sérialiser les informations de sauvegarde pour '%s': %v", info.ID, err)
		return err
	}

	// SECURITY: Restreindre les permissions du fichier de métadonnées
	if err := os.WriteFile(filename, data, 0600); err != nil {
		LogError("Impossible d'écrire le fichier de métadonnées %s: %v", filename, err)
		return err
	}
	LogInfo("Informations de sauvegarde pour '%s' sauvegardées dans %s.", info.ID, filename)
	return nil
}

// ListBackups liste toutes les sauvegardes disponibles
func ListBackups() ([]BackupInfo, error) {
	var backups []BackupInfo

	files, err := os.ReadDir(BackupInfoDir)
	if err != nil {
		LogError("Impossible de lire le répertoire des métadonnées %s: %v", BackupInfoDir, err)
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			data, err := os.ReadFile(filepath.Join(BackupInfoDir, file.Name()))
			if err != nil {
				LogError("Impossible de lire le fichier de métadonnées %s: %v", file.Name(), err)
				continue
			}

			var info BackupInfo
			if err := json.Unmarshal(data, &info); err != nil {
				LogError("Impossible de désérialiser le fichier de métadonnées %s: %v", file.Name(), err)
				continue
			}

			backups = append(backups, info)
		}
	}
	LogInfo("%d sauvegardes disponibles listées.", len(backups))
	return backups, nil
}

// DeleteBackup supprime une sauvegarde par son ID
func DeleteBackup(id string) error {
	LogSecurity("Tentative de suppression de la sauvegarde avec ID: %s", id)
	// Chercher la sauvegarde
	backups, err := ListBackups()
	if err != nil {
		LogError("Impossible de récupérer la liste des sauvegardes pour suppression: %v", err)
		return fmt.Errorf("impossible de récupérer la liste des sauvegardes: %w", err)
	}

	var backup BackupInfo
	found := false
	for _, b := range backups {
		if b.ID == id {
			backup = b
			found = true
			break
		}
	}

	if !found {
		LogWarning("Sauvegarde avec ID %s non trouvée pour suppression.", id)
		return fmt.Errorf("sauvegarde avec ID %s non trouvée", id)
	}

	// Supprimer le fichier de sauvegarde s'il est local
	if backup.RemoteServer == nil {
		// Si c'est un fichier local
		if FileExists(backup.BackupPath) {
			// Si c'est un fichier (sauvegarde compressée)
			if err := os.Remove(backup.BackupPath); err != nil {
				LogError("Impossible de supprimer le fichier de sauvegarde %s: %v", backup.BackupPath, err)
				return fmt.Errorf("impossible de supprimer le fichier de sauvegarde %s: %w", backup.BackupPath, err)
			}
			LogSecurity("Fichier de sauvegarde local %s supprimé.", backup.BackupPath)
		} else if DirExists(backup.BackupPath) {
			// Si c'est un répertoire (sauvegarde non compressée)
			if err := os.RemoveAll(backup.BackupPath); err != nil {
				LogError("Impossible de supprimer le répertoire de sauvegarde %s: %v", backup.BackupPath, err)
				return fmt.Errorf("impossible de supprimer le répertoire de sauvegarde %s: %w", backup.BackupPath, err)
			}
			LogSecurity("Répertoire de sauvegarde local %s supprimé.", backup.BackupPath)
		}
	} else {
		LogInfo("La sauvegarde '%s' est distante, la suppression du fichier distant doit être gérée manuellement.", backup.ID)
	}

	// Supprimer le fichier de métadonnées
	metaPath := filepath.Join(BackupInfoDir, id+".json")
	if err := os.Remove(metaPath); err != nil {
		LogError("Impossible de supprimer le fichier de métadonnées %s: %v", metaPath, err)
		return fmt.Errorf("impossible de supprimer le fichier de métadonnées %s: %w", metaPath, err)
	}
	LogSecurity("Fichier de métadonnées %s supprimé.", metaPath)

	LogSecurity("Sauvegarde avec ID %s supprimée avec succès.", id)
	return nil
}
