package restore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Noziop/s4v3my4ss/internal/wrappers"
	"github.com/Noziop/s4v3my4ss/pkg/common"
)

// RestoreBackup restaure une sauvegarde avec l'ID spécifié vers le chemin de destination
func RestoreBackup(backupID string, targetPath string) error {
	// Chercher la sauvegarde correspondante
	backupInfo, err := findBackupByID(backupID)
	if err != nil {
		return fmt.Errorf("impossible de trouver la sauvegarde: %w", err)
	}

	// Créer le répertoire de destination si nécessaire
	if targetPath == "" {
		// Si aucun chemin de destination n'est spécifié, utiliser le chemin d'origine
		targetPath = backupInfo.SourcePath
	}

	// Confirmer le remplacement si le répertoire existe déjà
	if common.DirExists(targetPath) {
		fmt.Printf("⚠️  Attention: Le répertoire '%s' existe déjà.\n", targetPath)
		fmt.Printf("Toutes les données existantes seront écrasées.\n")
		fmt.Print("Voulez-vous continuer? [o/N]: ")
		
		var response string
		fmt.Scanln(&response)
		
		if !strings.EqualFold(response, "o") && !strings.EqualFold(response, "oui") {
			return fmt.Errorf("restauration annulée par l'utilisateur")
		}
	} else {
		// Créer le répertoire s'il n'existe pas
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			return fmt.Errorf("impossible de créer le répertoire de destination: %w", err)
		}
	}

	// Chemin de la sauvegarde
	backupPath := backupInfo.BackupPath

	// Vérifier si la sauvegarde est compressée
	if backupInfo.Compression {
		// Vérifier l'existence du fichier tar.gz
		var compressedPath string
		
		// Vérifier si le chemin se termine déjà par .tar.gz
		if strings.HasSuffix(backupPath, ".tar.gz") {
			compressedPath = backupPath  // Utiliser directement le chemin existant
		} else {
			compressedPath = backupPath + ".tar.gz"  // Sinon ajouter l'extension
		}
		
		if !common.FileExists(compressedPath) {
			return fmt.Errorf("fichier de sauvegarde compressé introuvable: %s", compressedPath)
		}

		fmt.Printf("Décompression de la sauvegarde %s...\n", backupID)
		
		// Créer le wrapper de compression
		compressor, err := wrappers.NewCompressionWrapper()
		if err != nil {
			return fmt.Errorf("impossible d'initialiser le module de décompression: %w", err)
		}
		
		// Décompresser dans un répertoire temporaire
		tempDir := filepath.Join(common.TempDir, "restore_"+backupID)
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return fmt.Errorf("impossible de créer le répertoire temporaire: %w", err)
		}
		
		// Nettoyer le répertoire temporaire à la fin
		defer os.RemoveAll(tempDir)
		
		// Décompresser la sauvegarde
		if err := compressor.Decompress(compressedPath, tempDir); err != nil {
			return fmt.Errorf("erreur lors de la décompression: %w", err)
		}
		
		// Mettre à jour le chemin de la sauvegarde avec le chemin du répertoire décompressé
		// Rechercher le répertoire décompressé (souvent le seul sous-répertoire)
		entries, err := os.ReadDir(tempDir)
		if err != nil || len(entries) == 0 {
			return fmt.Errorf("erreur lors de l'accès au répertoire décompressé: %w", err)
		}
		
		// Utiliser le premier répertoire trouvé
		for _, entry := range entries {
			if entry.IsDir() {
				backupPath = filepath.Join(tempDir, entry.Name())
				break
			}
		}
	} else {
		// Vérifier l'existence du répertoire de sauvegarde
		if !common.DirExists(backupPath) {
			return fmt.Errorf("répertoire de sauvegarde introuvable: %s", backupPath)
		}
	}

	fmt.Printf("Restauration de '%s' (%s) vers '%s'...\n", 
		backupInfo.Name, 
		backupInfo.Time.Format("02/01/2006 15:04:05"), 
		targetPath)

	startTime := time.Now()

	// Restaurer la sauvegarde
	if err := wrappers.RsyncRestore(backupPath, targetPath, nil); err != nil {
		return fmt.Errorf("erreur lors de la restauration avec rsync: %w", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("✅ Restauration terminée avec succès en %v.\n", formatDuration(duration))
	return nil
}

// GetAvailableBackups récupère la liste des sauvegardes disponibles
func GetAvailableBackups() ([]common.BackupInfo, error) {
	return common.ListBackups()
}

// findBackupByID cherche une sauvegarde par son ID
func findBackupByID(id string) (common.BackupInfo, error) {
	backups, err := common.ListBackups()
	if err != nil {
		return common.BackupInfo{}, err
	}

	for _, backup := range backups {
		if backup.ID == id {
			return backup, nil
		}
	}

	return common.BackupInfo{}, fmt.Errorf("sauvegarde non trouvée avec l'ID: %s", id)
}

// formatDuration convertit une durée en chaîne lisible
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	
	hours := d / time.Hour
	d -= hours * time.Hour
	
	minutes := d / time.Minute
	d -= minutes * time.Minute
	
	seconds := d / time.Second
	
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	
	return fmt.Sprintf("%ds", seconds)
}