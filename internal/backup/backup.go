package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	
	"github.com/Noziop/s4v3my4ss/internal/wrappers"
	"github.com/Noziop/s4v3my4ss/pkg/common"
)

// BackupConfig contient les paramètres pour créer une sauvegarde
type BackupConfig struct {
	// Source directory to backup
	SourcePath string
	// Name of the backup
	Name string
	// Whether to compress the backup
	Compression bool
	// List of directories to exclude
	ExcludeDirs []string
	// List of files to exclude
	ExcludeFiles []string
	// Whether to create an incremental backup
	Incremental bool
}

// CreateBackup crée une sauvegarde d'un répertoire selon la configuration
func CreateBackup(config BackupConfig) error {
	// Générer un ID unique pour la sauvegarde
	backupID := generateBackupID(config.Name)
	
	// Déterminer le chemin de destination
	destPath := filepath.Join(common.AppConfig.BackupDestination, backupID)
	
	// Trouver la dernière sauvegarde pour faire une sauvegarde incrémentielle
	if config.Incremental {
		// À implémenter: logique pour sauvegarde incrémentielle
		// Note: cette fonctionnalité sera développée dans une prochaine version
		// La logique utiliserait findLastBackup() pour établir un lien vers la sauvegarde précédente
		_, err := findLastBackup(config.Name)
		if err != nil {
			fmt.Printf("Note: première sauvegarde de '%s', création d'une sauvegarde complète.\n", config.Name)
		}
	}
	
	// Créer le répertoire de destination
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("impossible de créer le répertoire de destination: %w", err)
	}
	
	fmt.Printf("Création d'une sauvegarde %s de %s vers %s...\n", 
		getBackupTypeStr(config.Incremental, config.Compression),
		config.SourcePath, 
		destPath)
	
	// Effectuer la sauvegarde avec rsync
	if err := wrappers.RsyncBackup(config.SourcePath, destPath, config.ExcludeDirs, config.ExcludeFiles, config.Compression, nil); err != nil {
		return fmt.Errorf("erreur lors de la sauvegarde avec rsync: %w", err)
	}
	
	// Calculer la taille de la sauvegarde avant compression
	size, err := getDirSize(destPath)
	if err != nil {
		fmt.Printf("Impossible de calculer la taille de la sauvegarde avant compression: %v\n", err)
		size = 0 // Initialiser pour éviter des erreurs plus tard
	}
	
	finalDestPath := destPath // Chemin final qui sera utilisé pour les métadonnées
	
	// Si la compression est activée, compresser la sauvegarde
	if config.Compression {
		compressedFile := destPath + ".tar.gz"
		if err := compressBackup(destPath, config.Name); err != nil {
			return fmt.Errorf("erreur lors de la compression: %w", err)
		}
		
		// Mettre à jour le chemin final et calculer la taille du fichier compressé
		finalDestPath = compressedFile
		if compressedSize, err := getFileSize(compressedFile); err == nil {
			size = compressedSize // Utiliser la taille du fichier compressé
		} else {
			fmt.Printf("Impossible de calculer la taille du fichier compressé: %v\n", err)
		}
	}
	
	// Créer l'info de sauvegarde
	backupInfo := common.BackupInfo{
		ID:           backupID,
		Name:         config.Name,
		SourcePath:   config.SourcePath,
		BackupPath:   finalDestPath,
		Time:         time.Now(),
		Size:         size,
		IsIncremental: config.Incremental,
		Compression:   config.Compression,
	}
	
	// Sauvegarder les métadonnées
	if err := common.SaveBackupInfo(backupInfo); err != nil {
		return fmt.Errorf("erreur lors de l'enregistrement des métadonnées: %w", err)
	}
	
	fmt.Printf("Sauvegarde terminée avec succès. Taille: %s\n", formatSize(size))
	
	// Nettoyer les anciennes sauvegardes selon la politique de rétention
	go cleanupOldBackups(config.Name)
	
	return nil
}

// getDirSize calcule la taille totale d'un répertoire
func getDirSize(path string) (int64, error) {
	var size int64
	
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	
	return size, err
}

// getFileSize récupère la taille d'un fichier unique
func getFileSize(path string) (int64, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

// generateBackupID génère un ID unique pour la sauvegarde
func generateBackupID(name string) string {
	// Format: name_date_hash
	timestamp := time.Now().Format("20060102_150405")
	
	// Ajouter une valeur aléatoire pour garantir l'unicité
	hash := sha256.Sum256([]byte(name + timestamp + fmt.Sprintf("%d", time.Now().UnixNano())))
	shortHash := hex.EncodeToString(hash[:3]) // Utiliser seulement les 6 premiers caractères
	
	// Nettoyer le nom pour qu'il soit utilisable dans un nom de fichier
	safeName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}
		return '_'
	}, name)
	
	return fmt.Sprintf("%s_%s_%s", safeName, timestamp, shortHash)
}

// findLastBackup trouve la dernière sauvegarde pour un nom donné
func findLastBackup(name string) (string, error) {
	backups, err := common.ListBackups()
	if err != nil {
		return "", err
	}
	
	var lastBackup common.BackupInfo
	var found bool
	
	for _, backup := range backups {
		if backup.Name == name {
			if !found || backup.Time.After(lastBackup.Time) {
				lastBackup = backup
				found = true
			}
		}
	}
	
	if !found {
		return "", fmt.Errorf("aucune sauvegarde précédente trouvée pour: %s", name)
	}
	
	return lastBackup.BackupPath, nil
}

// compressBackup compresse une sauvegarde terminée
func compressBackup(path string, name string) error {
	// Créer le wrapper de compression
	cw, err := wrappers.NewCompressionWrapper()
	if err != nil {
		return fmt.Errorf("impossible d'initialiser la compression: %w", err)
	}
	
	// Chemin du fichier compressé
	compressedFile := path + ".tar.gz"
	
	fmt.Printf("Compression de la sauvegarde vers %s...\n", compressedFile)
	
	// Compresser la sauvegarde
	if err := cw.Compress(path, compressedFile, wrappers.FormatTarGz); err != nil {
		return fmt.Errorf("erreur lors de la compression: %w", err)
	}
	
	// Suppression du répertoire original après compression réussie
	if err := os.RemoveAll(path); err != nil {
		fmt.Printf("Avertissement: impossible de supprimer le répertoire original après compression: %v\n", err)
	}
	
	fmt.Println("Compression terminée avec succès.")
	
	return nil
}

// cleanupOldBackups nettoie les anciennes sauvegardes selon la politique de rétention.
// Cette fonction est exécutée en arrière-plan après chaque sauvegarde.
func cleanupOldBackups(name string) {
	common.LogInfo("Démarrage du nettoyage des anciennes sauvegardes pour '%s'...", name)
	defer common.LogInfo("Nettoyage des anciennes sauvegardes pour '%s' terminé.", name)

	allBackups, err := common.ListBackups()
	if err != nil {
		common.LogError("cleanupOldBackups: impossible de lister les sauvegardes: %v", err)
		return
	}

	var relevantBackups []common.BackupInfo
	for _, b := range allBackups {
		if b.Name == name {
			relevantBackups = append(relevantBackups, b)
		}
	}

	if len(relevantBackups) == 0 {
		common.LogInfo("Aucune sauvegarde trouvée pour '%s'.", name)
		return
	}

	// Trier les sauvegardes par date, les plus anciennes en premier
	// Sort by time, oldest first
	for i := 0; i < len(relevantBackups); i++ {
		for j := i + 1; j < len(relevantBackups); j++ {
			if relevantBackups[i].Time.After(relevantBackups[j].Time) {
				relevantBackups[i], relevantBackups[j] = relevantBackups[j], relevantBackups[i]
			}
		}
	}

	policy := common.AppConfig.RetentionPolicy

	// Nettoyage quotidien
	dailyKept := cleanupByInterval(relevantBackups, policy.KeepDaily, common.Daily)
	// Nettoyage hebdomadaire
	weeklyKept := cleanupByInterval(dailyKept, policy.KeepWeekly, common.Weekly)
	// Nettoyage mensuel
	monthlyKept := cleanupByInterval(weeklyKept, policy.KeepMonthly, common.Monthly)

	// Supprimer les sauvegardes qui ne sont pas conservées par la politique
	toDelete := make(map[string]struct{})
	for _, b := range relevantBackups {
		toDelete[b.ID] = struct{}{}
	}

	for _, b := range dailyKept {
		delete(toDelete, b.ID)
	}
	for _, b := range weeklyKept {
		delete(toDelete, b.ID)
	}
	for _, b := range monthlyKept {
		delete(toDelete, b.ID)
	}

	for id := range toDelete {
		if err := common.DeleteBackup(id); err != nil {
			common.LogError("cleanupOldBackups: impossible de supprimer la sauvegarde %s: %v", id, err)
		} else {
			common.LogSecurity("Sauvegarde %s supprimée par la politique de rétention.", id)
		}
	}
}

// cleanupByInterval filtre les sauvegardes pour ne garder que celles qui respectent la politique de rétention pour un intervalle donné
func cleanupByInterval(backups []common.BackupInfo, keep int, interval common.RetentionInterval) []common.BackupInfo {
	if keep <= 0 {
		return []common.BackupInfo{} // Ne rien garder si la politique est 0 ou moins
	}

	var keptBackups []common.BackupInfo
	lastKeptTime := time.Time{} // Initialiser à l'heure zéro

	for _, b := range backups {
		shouldKeep := false
		switch interval {
		case common.Daily:
			// Garder si c'est la première sauvegarde du jour ou si le jour est différent
			if lastKeptTime.IsZero() || b.Time.YearDay() != lastKeptTime.YearDay() || b.Time.Year() != lastKeptTime.Year() {
				shouldKeep = true
			}
		case common.Weekly:
			// Garder si c'est la première sauvegarde de la semaine ou si la semaine est différente
			year1, week1 := b.Time.ISOWeek()
			year2, week2 := lastKeptTime.ISOWeek()
			if lastKeptTime.IsZero() || week1 != week2 || year1 != year2 {
				shouldKeep = true
			}
		case common.Monthly:
			// Garder si c'est la première sauvegarde du mois ou si le mois est différent
			if lastKeptTime.IsZero() || b.Time.Month() != lastKeptTime.Month() || b.Time.Year() != lastKeptTime.Year() {
				shouldKeep = true
			}
		}

		if shouldKeep {
			keptBackups = append(keptBackups, b)
			lastKeptTime = b.Time
		}
	}

	// Si on a plus de sauvegardes que la politique "keep", on supprime les plus anciennes
	if len(keptBackups) > keep {
		common.LogInfo("Suppression de %d sauvegardes excédentaires pour l'intervalle %s.", len(keptBackups)-keep, interval)
		return keptBackups[len(keptBackups)-keep:]
	}

	return keptBackups
}


// getBackupTypeStr renvoie une chaîne décrivant le type de sauvegarde
func getBackupTypeStr(incremental bool, compression bool) string {
	backupType := "complète"
	if incremental {
		backupType = "incrémentielle"
	}
	
	if compression {
		backupType += " compressée"
	}
	
	return backupType
}

// formatSize convertit une taille en octets en une chaîne lisible
func formatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	
	switch {
	case size >= GB:
		return fmt.Sprintf("%.1f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.1f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.1f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}