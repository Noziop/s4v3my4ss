package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Noziop/s4v3my4ss/internal/restore"
	"github.com/Noziop/s4v3my4ss/internal/ui/display"
	"github.com/Noziop/s4v3my4ss/internal/ui/input"
	"github.com/Noziop/s4v3my4ss/internal/watch"
	"github.com/Noziop/s4v3my4ss/internal/wrappers"
	"github.com/Noziop/s4v3my4ss/pkg/common"
)

// Patterns d'exclusion communs pour le développement
var (
	// Patterns de répertoires à exclure par défaut
	commonExcludeDirs = []string{
		"node_modules", ".git", ".svn", ".hg", "__pycache__", ".venv", "venv", "env",
		"dist", "build", "target", "bin", "obj", ".next", ".nuxt", ".output",
		"vendor", ".vscode", ".idea", ".DS_Store", "coverage", ".pytest_cache",
		".gradle", "tmp", "temp", "logs", "cache",
	}

	// Patterns de fichiers à exclure par défaut
	commonExcludeFiles = []string{
		"*.log", "*.tmp", "*.swp", "*.swo", "*.pyc", "*.pyo", "*.class", "*.o",
		"*.so", "*.exe", "*.dll", "*.db", "*.sqlite", "*.sqlite3", "*.pid",
		"package-lock.json", "yarn.lock", "pnpm-lock.yaml", ".env", ".cache", ".env.*",
	}
)







// ConfigureRemoteBackup configure une sauvegarde vers un serveur rsync distant
func ConfigureRemoteBackup(serverConfig common.RsyncServerConfig) {
	common.LogInfo("Début de la configuration d'une sauvegarde distante vers %s.", serverConfig.Name)
	display.ClearScreen()
	fmt.Printf("%sConfiguration d'une sauvegarde vers %s%s\n\n", display.ColorBold(), serverConfig.Name, display.ColorReset())
	
	name := input.ReadAndValidateInput("Nom de la sauvegarde: ", common.IsValidName, "Nom invalide.") // Utilisation de common.IsValidName
	sourcePath := input.ReadAndValidateInput("Chemin du répertoire à surveiller: ", common.IsValidPath, "Chemin invalide ou non sécurisé.") // Utilisation de common.IsValidPath
	
	// Expandir les chemins relatifs, y compris ~/
	if strings.HasPrefix(sourcePath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			sourcePath = filepath.Join(homeDir, sourcePath[2:])
		} else {
			common.LogError("Impossible d'obtenir le répertoire personnel de l'utilisateur pour la source distante: %v", err)
		}
	}
	
	// Normaliser et vérifier le chemin
	sourcePath = filepath.Clean(sourcePath)
	if !common.DirExists(sourcePath) {
		common.LogError("Répertoire source '%s' n'existe pas pour la sauvegarde distante.", sourcePath)
		fmt.Printf("%sErreur: Le répertoire '%s' n'existe pas.%s\n", display.ColorRed, sourcePath, display.ColorReset)
		return
	}
	
	// Sélectionner le module à utiliser
	module := serverConfig.DefaultModule
	if len(serverConfig.Modules) > 0 {
		common.LogInfo("Modules disponibles pour %s: %v", serverConfig.IP, serverConfig.Modules)
		fmt.Println("\nModules disponibles:")
		for i, mod := range serverConfig.Modules {
			fmt.Printf("%d. %s\n", i+1, mod)
		}
		
		moduleChoice := input.ReadInput("\nChoisissez un module par défaut (0 pour aucun): ")
		
		if moduleIdx, err := strconv.Atoi(moduleChoice); err == nil && moduleIdx > 0 && moduleIdx <= len(serverConfig.Modules) {
			module = serverConfig.Modules[moduleIdx-1]
			common.LogInfo("Module sélectionné pour la sauvegarde distante: %s.", module)
		} else {
			common.LogWarning("Choix de module invalide ou aucun module sélectionné pour la sauvegarde distante: %s", moduleChoice)
		}
	}
	
	// Construction du chemin de destination
	var destination string
	if serverConfig.SSHPort > 0 {
		// Format rsync+ssh avec port SSH personnalisé
		destination = fmt.Sprintf("rsync://%s@%s:%d/", serverConfig.Username, serverConfig.IP, serverConfig.SSHPort)
	} else {
		// Format rsync standard
		destination = fmt.Sprintf("rsync://%s@%s/", serverConfig.Username, serverConfig.IP)
	}
	
	if module != "" {
		destination += module
	}
	
	// Option pour une sauvegarde incrémentale
	incrementalStr := input.ReadInput("Activer les sauvegardes incrémentales? (o/n): ")
	incremental := strings.ToLower(incrementalStr) == "o"
	
	// Compression
	compressStr := input.ReadInput("Activer la compression? (o/n): ")
	compression := strings.ToLower(compressStr) == "o"
	
	// Exclusion de fichiers/dossiers
	excludeDirsStr := input.ReadInput("\nRépertoires à exclure (séparés par des virgules): ")
	var excludeDirs []string
	if excludeDirsStr != "" {
		excludeDirs = strings.Split(excludeDirsStr, ",")
		for i, dir := range excludeDirs {
			excludeDirs[i] = strings.TrimSpace(dir)
			if !common.IsValidExcludePattern(excludeDirs[i]) { // Utilisation de common.IsValidExcludePattern
				common.LogError("Modèle d'exclusion de répertoire invalide pour la sauvegarde distante: %s", excludeDirs[i])
				fmt.Printf("%sModèle d'exclusion de répertoire invalide: %s%s\n", display.ColorRed, excludeDirs[i], display.ColorReset)
				return
			}
		}
	} else {
		excludeDirs = commonExcludeDirs
	}
	
	excludeFilesStr := input.ReadInput("Fichiers à exclure (séparés par des virgules): ")
	var excludeFiles []string
	if excludeFilesStr != "" {
		excludeFiles = strings.Split(excludeFilesStr, ",")
		for i, file := range excludeFiles {
			excludeFiles[i] = strings.TrimSpace(file)
			if !common.IsValidExcludePattern(excludeFiles[i]) { // Utilisation de common.IsValidExcludePattern
				common.LogError("Modèle d'exclusion de fichier invalide pour la sauvegarde distante: %s", excludeFiles[i])
				fmt.Printf("%sModèle d'exclusion de fichier invalide: %s%s\n", display.ColorRed, excludeFiles[i], display.ColorReset)
				return
			}
		}
	} else {
		excludeFiles = commonExcludeFiles
	}
	
	// Intervalle de sauvegarde automatique
	intervalStr := input.ReadInput("Intervalle de sauvegarde en minutes (0 pour désactiver): ")
	interval := 0
	if intervalVal, err := strconv.Atoi(intervalStr); err == nil && intervalVal > 0 {
		interval = intervalVal
	} else {
		common.LogWarning("Intervalle de sauvegarde distante invalide: %s. Utilisation de 0.", intervalStr)
	}
	
	// Créer la configuration de sauvegarde
	backupConfig := common.BackupConfig{
		Name:          name,
		SourcePath:    sourcePath,
		Compression:   compression,
		ExcludeDirs:   excludeDirs,
		ExcludeFiles:  excludeFiles,
		Interval:      interval,
		IsIncremental: incremental,
		RemoteServer:  &serverConfig, // Utiliser l'adresse de serverConfig pour obtenir un pointeur
	}
	
	// Mise à jour de la destination dans la configuration globale
	// Sauvegarder la destination temporairement
	prevDestination := common.AppConfig.BackupDestination
	common.AppConfig.BackupDestination = destination
	
	// Enregistrer la configuration
	if err := common.AddBackupDirectory(backupConfig); err != nil {
		common.LogError("Erreur lors de l'enregistrement de la configuration de sauvegarde distante %s: %v", backupConfig.Name, err)
		fmt.Printf("%sErreur lors de l'enregistrement de la configuration: %v%s\n", display.ColorRed, err, display.ColorReset)
		common.AppConfig.BackupDestination = prevDestination // Restaurer l'ancienne destination
	} else {
		common.LogInfo("Configuration de sauvegarde distante %s enregistrée avec succès.", backupConfig.Name)
		if err := common.SaveConfig(common.AppConfig); err != nil {
			common.LogError("Erreur lors de la mise à jour de la destination de sauvegarde distante: %v", err)
			fmt.Printf("%sErreur lors de la mise à jour de la destination: %v%s\n", display.ColorRed(), err, display.ColorReset())
		} else {
			common.LogInfo("Destination de sauvegarde distante mise à jour avec succès.")
			fmt.Printf("\n%sConfiguration de sauvegarde vers %s enregistrée avec succès!%s\n", 
				display.ColorGreen(), serverConfig.Name, display.ColorReset())
			
			// Proposer de faire une sauvegarde immédiate
			doBackupNow := input.ReadInput("\nVoulez-vous effectuer une sauvegarde immédiate? (o/n): ")
			if strings.ToLower(doBackupNow) == "o" {
				common.LogInfo("Sauvegarde immédiate vers %s demandée.", serverConfig.Name)
				fmt.Printf("\nDémarrage de la sauvegarde vers %s...\n", serverConfig.Name)
				
				// Générer un ID unique pour la sauvegarde
				backupID := common.GenerateBackupID(name)
				
				// Utiliser la fonction RsyncBackup pour effectuer la sauvegarde
				if err := wrappers.RsyncBackup(sourcePath, destination, excludeDirs, excludeFiles, compression, &serverConfig); err != nil {
					common.LogError("Erreur lors de la sauvegarde immédiate vers %s: %v", serverConfig.Name, err)
					fmt.Printf("%sErreur lors de la sauvegarde: %v%s\n", display.ColorRed(), err, display.ColorReset())
					return
				}
				
				// Obtenir la taille du répertoire source comme approximation
				size, err := common.GetDirSize(sourcePath)
				if err != nil {
					common.LogError("Impossible de calculer la taille de la sauvegarde: %v", err)
					fmt.Printf("Impossible de calculer la taille de la sauvegarde: %v\n", err)
					size = 0
				}
				
				// Créer l'info de sauvegarde
				backupInfo := common.BackupInfo{
					ID:           backupID,
					Name:         name,
					SourcePath:   sourcePath,
					BackupPath:   destination,
					Time:         time.Now(),
					Size:         size,
					IsIncremental: incremental,
					Compression:   compression,
					RemoteServer: &serverConfig, // Utiliser l'adresse de serverConfig pour obtenir un pointeur
				}
				
				// Sauvegarder les métadonnées
				if err := common.SaveBackupInfo(backupInfo); err != nil {
					common.LogError("Erreur lors de l'enregistrement des métadonnées pour %s: %v", backupInfo.ID, err)
					fmt.Printf("%sErreur lors de l'enregistrement des métadonnées: %v%s\n", display.ColorRed, err, display.ColorReset)
					return
				}
				
				common.LogInfo("Sauvegarde immédiate vers %s terminée avec succès.", serverConfig.Name)
				fmt.Printf("%sSauvegarde vers %s terminée avec succès!%s\n", display.ColorGreen(), serverConfig.Name, display.ColorReset())
			} else {
				common.LogInfo("Sauvegarde immédiate annulée par l'utilisateur.")
			}
		}
	}
}

// HandleWatchCommand traite la commande 'watch' depuis la ligne de commande
func HandleWatchCommand(args []string) {
	common.LogInfo("Traitement de la commande 'watch' avec les arguments: %v", args)
	if len(args) < 1 {
		common.LogError("Utilisation incorrecte de la commande watch: arguments manquants.")
		fmt.Fprintln(os.Stderr, "Usage: " + common.CommandName + " watch <nom_configuration>")
		os.Exit(1)
	}

	name := args[0]
	if !common.IsValidName(name) { // Utilisation de common.IsValidName
		common.LogError("Nom de configuration invalide fourni pour watch: %s", name)
		fmt.Fprintf(os.Stderr, "Erreur: Nom de configuration invalide: %s\n", name)
		os.Exit(1)
	}

	config, found := common.GetBackupConfig(name)
	if !found {
		common.LogError("Configuration '%s' non trouvée pour la commande watch.", name)
		fmt.Fprintf(os.Stderr, "Erreur: Configuration '%s' non trouvée\n", name)
		os.Exit(1)
	}

	fmt.Printf("Démarrage de la surveillance du répertoire: %s\n", config.SourcePath)
	common.LogInfo("Démarrage de la surveillance du répertoire: %s pour la configuration %s.", config.SourcePath, config.Name)
	if err := watch.StartWatch(config); err != nil {
		common.LogError("Erreur de surveillance pour %s: %v", config.Name, err)
		fmt.Fprintf(os.Stderr, "Erreur de surveillance: %v\n", err)
		os.Exit(1)
	}
	common.LogInfo("Surveillance démarrée avec succès pour %s.", config.Name)
}

// HandleRestoreCommand traite la commande 'restore' depuis la ligne de commande
func HandleRestoreCommand(args []string) {
	common.LogInfo("Traitement de la commande 'restore' avec les arguments: %v", args)
	if len(args) < 1 {
		common.LogInfo("Aucun argument fourni pour restore. Lancement du mode interactif.")
		RestoreBackupInteractive(false)
		return
	}

	backupID := args[0]
	if !common.IsValidName(backupID) { // Utilisation de common.IsValidName
		common.LogError("ID de sauvegarde invalide fourni pour restore: %s", backupID)
		fmt.Fprintf(os.Stderr, "Erreur: ID de sauvegarde invalide: %s\n", backupID)
		os.Exit(1)
	}

	target := ""
	if len(args) > 1 {
		target = args[1]
		if !common.IsValidPath(target) { // Utilisation de common.IsValidPath
			common.LogError("Chemin de destination invalide fourni pour restore: %s", target)
			fmt.Fprintf(os.Stderr, "Erreur: Chemin de destination invalide: %s\n", target)
			os.Exit(1)
		}
	}

	common.LogInfo("Restauration de la sauvegarde '%s' vers '%s'.", backupID, target)
	fmt.Printf("Restauration de la sauvegarde '%s' vers '%s'...\n", backupID, target)

	if err := restore.RestoreBackup(backupID, target); err != nil {
		common.LogError("Erreur de restauration pour %s: %v", backupID, err)
		fmt.Fprintf(os.Stderr, "Erreur de restauration: %v\n", err)
		os.Exit(1)
	}

	common.LogInfo("Restauration terminée avec succès pour %s.", backupID)
	fmt.Println("Restauration terminée avec succès.")
}


