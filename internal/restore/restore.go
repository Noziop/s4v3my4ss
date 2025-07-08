package restore

import (
	"fmt"
	os "os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Noziop/s4v3my4ss/internal/wrappers"
	"github.com/Noziop/s4v3my4ss/pkg/common"
)

// RestoreBackup restaure une sauvegarde avec l'ID spécifié vers le chemin de destination
func RestoreBackup(backupID string, targetPath string) error {
	common.LogInfo("Début de la restauration de la sauvegarde %s vers %s.", backupID, targetPath)
	// Chercher la sauvegarde correspondante
	backupInfo, err := findBackupByID(backupID)
	if err != nil {
		common.LogError("Impossible de trouver la sauvegarde %s: %v", backupID, err)
		return fmt.Errorf("impossible de trouver la sauvegarde: %w", err)
	}

	// Créer le répertoire de destination si nécessaire
	if targetPath == "" {
		// Si aucun chemin de destination n'est spécifié, utiliser le chemin d'origine
		targetPath = backupInfo.SourcePath
	}

	// SECURITY: Valider le chemin de destination avant toute opération de fichier.
	if !common.IsValidPath(targetPath) {
		common.LogError("Chemin de destination invalide ou non sécurisé: %s", targetPath)
		return fmt.Errorf("chemin de destination invalide ou non sécurisé: %s", targetPath)
	}
	// SECURITY: Vérifier si le chemin de destination est autorisé par la configuration de sécurité.
	if !common.AppConfig.Security.IsPathAllowed(targetPath) {
		common.LogSecurity("Tentative de restauration vers un chemin non autorisé: %s", targetPath)
		return fmt.Errorf("le chemin de destination '%s' n'est pas autorisé par la politique de sécurité", targetPath)
	}

	// Confirmer le remplacement si le répertoire existe déjà
	if common.DirExists(targetPath) {
		fmt.Printf("⚠️  Attention: Le répertoire '%s' existe déjà.\n", targetPath)
		fmt.Printf("Toutes les données existantes seront écrasées.\n")
		fmt.Print("Voulez-vous continuer? [o/N]: ")
		
		var response string
		fmt.Scanln(&response)
		
		if !strings.EqualFold(response, "o") && !strings.EqualFold(response, "oui") {
			common.LogInfo("Restauration annulée par l'utilisateur.")
			return fmt.Errorf("restauration annulée par l'utilisateur")
		}
	} else {
		// Créer le répertoire s'il n'existe pas
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			common.LogError("Impossible de créer le répertoire de destination %s: %v", targetPath, err)
			return fmt.Errorf("impossible de créer le répertoire de destination: %w", err)
		}
		common.LogInfo("Répertoire de destination %s créé.", targetPath)
	}

	// Chemin de la sauvegarde
	backupPath := backupInfo.BackupPath

	// SECURITY: Gérer le chiffrement si la sauvegarde est chiffrée
	if backupInfo.Encrypted {
		common.LogSecurity("Détection d'une sauvegarde chiffrée (%s). Tentative de déchiffrement...", backupID)
		// Ici, vous intégreriez la logique de déchiffrement.
		// Par exemple, en utilisant une bibliothèque de chiffrement et la clé AppConfig.EncryptionKey.
		// Pour l'instant, c'est un placeholder.
		common.LogError("Le déchiffrement n'est pas encore implémenté pour la sauvegarde %s.", backupID)
		return fmt.Errorf("le déchiffrement n'est pas encore implémenté")
	}

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
			common.LogError("Fichier de sauvegarde compressé introuvable: %s", compressedPath)
			return fmt.Errorf("fichier de sauvegarde compressé introuvable: %s", compressedPath)
		}

		common.LogInfo("Décompression de la sauvegarde %s...", backupID)
		
		// Créer le wrapper de compression
		compressor, err := wrappers.NewCompressionWrapper()
		if err != nil {
			common.LogError("Impossible d'initialiser le module de décompression: %v", err)
			return fmt.Errorf("impossible d'initialiser le module de décompression: %w", err)
		}
		
		// Décompresser dans un répertoire temporaire
		tempDir := filepath.Join(common.TempDir, "restore_"+backupID)
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			common.LogError("Impossible de créer le répertoire temporaire %s: %v", tempDir, err)
			return fmt.Errorf("impossible de créer le répertoire temporaire: %w", err)
		}
		common.LogInfo("Répertoire temporaire %s créé pour la décompression.", tempDir)
		
		// Nettoyer le répertoire temporaire à la fin
		defer func() {
			err := os.RemoveAll(tempDir)
			if err != nil {
				common.LogError("Erreur lors du nettoyage du répertoire temporaire %s: %v", tempDir, err)
			}
			common.LogInfo("Répertoire temporaire %s nettoyé.", tempDir)
		}()
		
		// Décompresser la sauvegarde
		if err := compressor.Decompress(compressedPath, tempDir); err != nil {
			common.LogError("Erreur lors de la décompression de %s: %v", compressedPath, err)
			return fmt.Errorf("erreur lors de la décompression: %w", err)
		}
		common.LogInfo("Décompression de %s terminée avec succès dans %s.", compressedPath, tempDir)
		
		// Mettre à jour le chemin de la sauvegarde avec le chemin du répertoire décompressé
		// Rechercher le répertoire décompressé (souvent le seul sous-répertoire)
		entries, err := os.ReadDir(tempDir)
		if err != nil || len(entries) == 0 {
			common.LogError("Erreur lors de l'accès au répertoire décompressé %s: %v", tempDir, err)
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
			common.LogError("Répertoire de sauvegarde introuvable: %s", backupPath)
			return fmt.Errorf("répertoire de sauvegarde introuvable: %s", backupPath)
		}
	}

	common.LogInfo("Restauration de '%s' (%s) vers '%s'...", 
		backupInfo.Name, 
		backupInfo.Time.Format("02/01/2006 15:04:05"), 
		targetPath)

	startTime := time.Now()

	// Restaurer la sauvegarde
	if err := wrappers.RsyncRestore(backupPath, targetPath, nil); err != nil {
		common.LogError("Erreur lors de la restauration avec rsync de %s vers %s: %v", backupPath, targetPath, err)
		return fmt.Errorf("erreur lors de la restauration avec rsync: %w", err)
	}

	duration := time.Since(startTime)
	common.LogInfo("Restauration terminée avec succès en %v.", formatDuration(duration))
	return nil
}

// GetAvailableBackups récupère la liste des sauvegardes disponibles
func GetAvailableBackups() ([]common.BackupInfo, error) {
	common.LogInfo("Récupération de la liste des sauvegardes disponibles.")
	backups, err := common.ListBackups()
	if err != nil {
		common.LogError("Impossible de lister les sauvegardes: %v", err)
		return nil, err
	}
	common.LogInfo("%d sauvegardes disponibles trouvées.", len(backups))
	return backups, nil
}

// findBackupByID cherche une sauvegarde par son ID
func findBackupByID(id string) (common.BackupInfo, error) {
	common.LogInfo("Recherche de la sauvegarde avec ID: %s", id)
	backups, err := common.ListBackups()
	if err != nil {
		common.LogError("Impossible de lister les sauvegardes pour la recherche par ID: %v", err)
		return common.BackupInfo{}, err
	}

	for _, backup := range backups {
		if backup.ID == id {
			common.LogInfo("Sauvegarde avec ID %s trouvée.", id)
			return backup, nil
		}
	}

	common.LogWarning("Sauvegarde non trouvée avec l'ID: %s", id)
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