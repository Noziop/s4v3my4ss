package ui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Noziop/s4v3my4ss/internal/restore"
	"github.com/Noziop/s4v3my4ss/internal/watch"
	"github.com/Noziop/s4v3my4ss/internal/wrappers"
	"github.com/Noziop/s4v3my4ss/pkg/common" // Import du package common
)

// Couleurs pour l'interface
const (
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[0;33m"
	colorBlue   = "\033[0;34m"
	colorBold   = "\033[1m"
	colorReset  = "\033[0m"
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

// Constante pour le nom de la commande
const CommandName = "saveme"

var reader *bufio.Reader

// readAndValidateInput lit une entrée utilisateur et la valide en utilisant une fonction de validation.
// La lecture se répète jusqu'à ce que l'entrée soit valide.
func readAndValidateInput(prompt string, validator func(string) bool, errorMessage string) string {
	for {
		input := readInput(prompt)
		if validator(input) {
			return input
		}
		fmt.Printf("%s%s%s\n", colorRed, errorMessage, colorReset)
	}
}

// RunInteractiveMode démarre le mode interactif de l'application
func RunInteractiveMode() {
	common.LogInfo("Démarrage du mode interactif.")
	// SECURITY: Vérifier si l'application est exécutée en tant que root.
	if os.Geteuid() == 0 {
		common.LogSecurity("AVERTISSEMENT: Application exécutée en tant que root.")
		fmt.Printf("%s%sAVERTISSEMENT: Vous exécutez cette application en tant que root. Il est fortement recommandé d'utiliser un utilisateur avec des privilèges moindres pour des raisons de sécurité.%s\n\n", colorBold, colorYellow, colorReset)
	}

	reader = bufio.NewReader(os.Stdin)


	for {
		clearScreen()
		displayHeader()
		displayMainMenu()

		choice := readInput("Votre choix: ")

		switch choice {
		case "1":
			configureBackup()
		case "2":
			watchDirectoryInteractive()
		case "3":
			restoreBackupInteractive()
		case "4":
			manageBackupsInteractive()
		case "5":
			checkDependencies()
		case "6":
			discoverRsyncServers()
		case "7":
			manageConfiguration()
		case "0":
			common.LogInfo("Quitter l'application.")
			fmt.Println("Au revoir !")
			return
		default:
			common.LogWarning("Option de menu non valide: %s", choice)
			fmt.Println("Option non valide. Veuillez réessayer.")
		}

		fmt.Println()
		readInput("Appuyez sur Entrée pour continuer...")
	}
}

// HandleWatchCommand traite la commande 'watch' depuis la ligne de commande
func HandleWatchCommand(args []string) {
	common.LogInfo("Traitement de la commande 'watch' avec les arguments: %v", args)
	if len(args) < 1 {
		common.LogError("Utilisation incorrecte de la commande watch: arguments manquants.")
		fmt.Fprintln(os.Stderr, "Usage: "+CommandName+" watch <nom_configuration>")
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
		restoreBackupInteractive()
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

	common.LogInfo("Restauration de la sauvegarde %s vers %s.", backupID, target)
	if err := restore.RestoreBackup(backupID, target); err != nil {
		common.LogError("Erreur de restauration pour %s: %v", backupID, err)
		fmt.Fprintf(os.Stderr, "Erreur de restauration: %v\n", err)
		os.Exit(1)
	}

	common.LogInfo("Restauration terminée avec succès pour %s.", backupID)
	fmt.Println("Restauration terminée avec succès.")
}

// HandleManageCommand traite la commande 'manage' depuis la ligne de commande
func HandleManageCommand(args []string) {
	common.LogInfo("Traitement de la commande 'manage' avec les arguments: %v", args)
	if len(args) == 0 {
		common.LogInfo("Aucun argument fourni pour manage. Lancement du mode interactif.")
		manageBackupsInteractive()
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "list":
		common.LogInfo("Exécution de la sous-commande manage list.")
		listBackups()
	case "delete":
		common.LogInfo("Exécution de la sous-commande manage delete.")
		if len(args) < 2 {
			common.LogError("Utilisation incorrecte de manage delete: ID de sauvegarde manquant.")
			fmt.Fprintln(os.Stderr, "Usage: "+CommandName+" manage delete <backup_id>")
			os.Exit(1)
		}
		backupID := args[1]
		if !common.IsValidName(backupID) { // Utilisation de common.IsValidName
			common.LogError("ID de sauvegarde invalide fourni pour manage delete: %s", backupID)
			fmt.Fprintf(os.Stderr, "Erreur: ID de sauvegarde invalide: %s\n", backupID)
			os.Exit(1)
		}
		deleteBackup(backupID)
	case "clean":
		common.LogInfo("Exécution de la sous-commande manage clean.")
		cleanOldBackups()
	default:
		common.LogWarning("Sous-commande manage inconnue: %s", subcommand)
		fmt.Fprintln(os.Stderr, "Sous-commande inconnue:", subcommand)
		fmt.Fprintln(os.Stderr, "Sous-commandes disponibles: list, delete, clean")
		os.Exit(1)
	}
}

// configureBackup permet de configurer une nouvelle sauvegarde
func configureBackup() {
	common.LogInfo("Début de la configuration d'une nouvelle sauvegarde.")
	fmt.Printf("%sConfiguration d'une nouvelle sauvegarde%s\n\n", colorBold, colorReset)

	// SECURITY: Utiliser la validation pour les entrées utilisateur.
	name := readAndValidateInput("Nom de la sauvegarde: ", common.IsValidName, "Nom invalide. Utilisez uniquement des lettres, chiffres, - et _.") // Utilisation de common.IsValidName
	sourcePath := readAndValidateInput("Chemin du répertoire à surveiller: ", common.IsValidPath, "Chemin invalide ou non sécurisé.") // Utilisation de common.IsValidPath

	if sourcePath == "" {
		common.LogWarning("Configuration annulée: chemin source vide.")
		fmt.Println("Configuration annulée. Le chemin source ne peut pas être vide.")
		return
	}

	// Expandir les chemins relatifs, y compris ~/
	if strings.HasPrefix(sourcePath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			sourcePath = filepath.Join(homeDir, sourcePath[2:])
		} else {
			common.LogError("Impossible d'obtenir le répertoire personnel de l'utilisateur: %v", err)
		}
	}

	// Normaliser et vérifier le chemin
	sourcePath = filepath.Clean(sourcePath)
	if !common.DirExists(sourcePath) {
		common.LogError("Répertoire source '%s' n'existe pas.", sourcePath)
		fmt.Printf("%sErreur: Le répertoire '%s' n'existe pas.%s\n", colorRed, sourcePath, colorReset)
		return
	}

	// Afficher les destinations de sauvegarde disponibles
	destChoice := ""
	destinations := common.AppConfig.BackupDestinations
	
	if len(destinations) > 0 {
		common.LogInfo("Affichage des destinations de sauvegarde disponibles.")
		fmt.Println("\nDestinations de sauvegarde disponibles:")
		defaultIndex := -1
		
		for i, dest := range destinations {
			
			destTypeStr := ""
			if dest.Type == "rsync" {
				destTypeStr = "(rsync)"
			} else if dest.Type == "local" {
				destTypeStr = "(local)"
			}
			
			defaultStr := ""
			if dest.IsDefault {
				defaultStr = " (Par défaut)"
			}
			
			fmt.Printf("%d. %s %s%s\n   Chemin: %s\n", i+1, dest.Name, destTypeStr, defaultStr, dest.Path)
		}
		
		fmt.Printf("%d. Nouvelle destination\n", len(destinations)+1)
		
		destChoiceStr := readInput("\nEmplacement des sauvegardes (numéro ou vide pour utiliser l'emplacement par défaut): ")
		
		if destChoiceStr == "" && defaultIndex >= 0 {
			destChoice = destinations[defaultIndex].Name
		} else if destChoiceStr != "" {
			destIdx, err := strconv.Atoi(destChoiceStr)
			if err == nil && destIdx > 0 {
				if destIdx <= len(destinations) {
					destChoice = destinations[destIdx-1].Name
				} else if destIdx == len(destinations)+1 {
					// Nouvelle destination
					destChoice = "new"
				}
			} else {
				common.LogWarning("Choix de destination invalide: %s", destChoiceStr)
			}
		}
	}
	
	// Configuration de la destination de sauvegarde
	backupDestinationName := ""
	backupDestination := ""
	
	if destChoice == "new" || (len(destinations) == 0 && destChoice == "") {
		common.LogInfo("Configuration d'une nouvelle destination de sauvegarde.")
		// Demander un nouvel emplacement de sauvegarde
		backupDestination = readAndValidateInput("Nouvel emplacement des sauvegardes: ", common.IsValidPath, "Chemin invalide ou non sécurisé.") // Utilisation de common.IsValidPath
		
		if backupDestination != "" {
			// Expandir les chemins relatifs, y compris ~/
			if strings.HasPrefix(backupDestination, "~/") {
				homeDir, err := os.UserHomeDir()
				if err == nil {
					backupDestination = filepath.Join(homeDir, backupDestination[2:])
				} else {
					common.LogError("Impossible d'obtenir le répertoire personnel de l'utilisateur pour la destination: %v", err)
				}
			}
			
			// Normaliser le chemin
			backupDestination = filepath.Clean(backupDestination)
			
			// Détecter si c'est un chemin rsync
			destType := "local"
			if strings.HasPrefix(backupDestination, "rsync://") {
				destType = "rsync"
			}
			
			// Vérifier/créer le répertoire de destination (seulement pour les destinations locales)
			if destType == "local" && !common.DirExists(backupDestination) {
				fmt.Printf("Le répertoire de destination n'existe pas. Voulez-vous le créer? (o/n): ")
				createDest := readInput("")
				if strings.ToLower(createDest) == "o" {
					if err := os.MkdirAll(backupDestination, 0755); err != nil {
						common.LogError("Erreur lors de la création du répertoire de destination %s: %v", backupDestination, err)
						return
					}
					common.LogInfo("Répertoire de destination %s créé.", backupDestination)
				} else {
					common.LogInfo("Création de répertoire annulée par l'utilisateur.")
					fmt.Println("Configuration annulée.")
					return
				}
			}
			
			// Demander un nom pour la nouvelle destination
			newDestName := readAndValidateInput("Nom de la nouvelle destination: ", common.IsValidName, "Nom invalide.") // Utilisation de common.IsValidName
			if newDestName == "" {
				newDestName = "Destination " + strconv.Itoa(len(destinations)+1)
			}
			
			// Demander si cette destination doit être la destination par défaut
			defaultDestStr := readInput("Définir comme destination par défaut? (o/n): ")
			isDefaultDest := strings.ToLower(defaultDestStr) == "o"
			
			// Créer et ajouter la nouvelle destination
			newDest := common.BackupDestination{
				Name:      newDestName,
				Path:      backupDestination,
				Type:      destType,
				IsDefault: isDefaultDest,
			}
			
			if err := common.AddBackupDestination(newDest); err != nil {
				common.LogError("Erreur lors de l'ajout de la destination %s: %v", newDest.Name, err)
				return
			}
			
			backupDestinationName = newDestName
		} else {
			common.LogWarning("Nouvelle destination de sauvegarde annulée: chemin vide.")
		}
	} else if destChoice != "" {
		common.LogInfo("Utilisation de la destination de sauvegarde existante: %s", destChoice)
		// Utiliser une destination existante
		backupDestinationName = destChoice
	}
	
	// Option pour une sauvegarde incrémentale
	incrementalStr := readInput("Activer les sauvegardes incrémentales? (o/n): ")
	incremental := strings.ToLower(incrementalStr) == "o"
	
	// Compression
	compressStr := readInput("Activer la compression? (o/n): ")
	compression := strings.ToLower(compressStr) == "o"
	
	// Répertoires à exclure
	excludeDirsStr := readInput("Répertoires à exclure (séparés par des virgules): ")
	var excludeDirs []string
	if excludeDirsStr != "" {
		excludeDirs = strings.Split(excludeDirsStr, ",")
		for i, dir := range excludeDirs {
			excludeDirs[i] = strings.TrimSpace(dir)
			if !common.IsValidExcludePattern(excludeDirs[i]) { // Utilisation de common.IsValidExcludePattern
				common.LogError("Modèle d'exclusion de répertoire invalide: %s", excludeDirs[i])
				fmt.Printf("%sModèle d'exclusion de répertoire invalide: %s%s\n", colorRed, excludeDirs[i], colorReset)
				return
			}
		}
	} else {
		excludeDirs = commonExcludeDirs
	}
	
	// Fichiers à exclure
	excludeFilesStr := readInput("Fichiers à exclure (séparés par des virgules): ")
	var excludeFiles []string
	if excludeFilesStr != "" {
		excludeFiles = strings.Split(excludeFilesStr, ",")
		for i, file := range excludeFiles {
			excludeFiles[i] = strings.TrimSpace(file)
			if !common.IsValidExcludePattern(excludeFiles[i]) { // Utilisation de common.IsValidExcludePattern
				common.LogError("Modèle d'exclusion de fichier invalide: %s", excludeFiles[i])
				fmt.Printf("%sModèle d'exclusion de fichier invalide: %s%s\n", colorRed, excludeFiles[i], colorReset)
				return
			}
		}
	} else {
		excludeFiles = commonExcludeFiles
	}
	
	// Intervalle (0 = pas de surveillance automatique)
	intervalStr := readInput("Intervalle de sauvegarde en minutes (0 pour désactiver): ")
	interval := 0
	if i, err := strconv.Atoi(intervalStr); err == nil && i >= 0 {
		interval = i
	} else {
		common.LogWarning("Intervalle de sauvegarde invalide: %s. Utilisation de 0.", intervalStr)
	}
	
	// Créer la configuration
	config := common.BackupConfig{
		Name:          name,
		SourcePath:    sourcePath,
		Compression:   compression,
		ExcludeDirs:   excludeDirs,
		ExcludeFiles:  excludeFiles,
		Interval:      interval,
		IsIncremental: incremental,
		DestinationName: backupDestinationName,
	}
	
	// Ajouter à la configuration
	if err := common.AddBackupDirectory(config); err != nil {
		common.LogError("Erreur lors de l'ajout de la configuration de sauvegarde %s: %v", config.Name, err)
		fmt.Printf("%sErreur lors de l'ajout de la configuration: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	common.LogInfo("Configuration de sauvegarde %s ajoutée avec succès.", config.Name)
	fmt.Printf("%sConfiguration ajoutée avec succès.%s\n", colorGreen, colorReset)
	
	// Proposer de démarrer la surveillance
	startWatchStr := readInput("Démarrer la surveillance maintenant? (o/n): ")
	if strings.ToLower(startWatchStr) == "o" {
		common.LogInfo("Démarrage de la surveillance pour %s.", config.Name)
		fmt.Printf("Démarrage de la surveillance du répertoire: %s\n", config.SourcePath)
		fmt.Println("Mode surveillance continue. La surveillance s'exécute en arrière-plan.")
		fmt.Println("Les sauvegardes continueront même si vous revenez au menu principal.")
		
		// Lancer la surveillance en arrière-plan pour ne pas bloquer l'interface
		go func() {
			if err := watch.StartWatch(config); err != nil {
				common.LogError("Erreur lors de la surveillance en arrière-plan pour %s: %v", config.Name, err)
				fmt.Printf("%sErreur lors de la surveillance: %v%s\n", colorRed, err, colorReset)
			}
		}()
	}
}

// watchDirectoryInteractive permet de démarrer la surveillance d'un répertoire
func watchDirectoryInteractive() {
	common.LogInfo("Début de la surveillance interactive.")
	fmt.Printf("%sSurveillance d'un répertoire%s\n\n", colorBold, colorReset)
	
	// Afficher les configurations disponibles
	configs := common.AppConfig.BackupDirs
	if len(configs) == 0 {
		common.LogWarning("Aucune configuration de sauvegarde disponible pour la surveillance.")
		fmt.Printf("%sAucune configuration de sauvegarde n'est disponible.%s\n", colorYellow, colorReset)
		fmt.Println("Veuillez d'abord créer une configuration.")
		return
	}
	
	fmt.Println("Configurations disponibles:")
	for i, cfg := range configs {
		fmt.Printf("%d. %s (%s)\n", i+1, cfg.Name, cfg.SourcePath)
	}
	
	// Demander quelle configuration utiliser
	idxStr := readInput("Sélectionnez une configuration (numéro): ")
	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 1 || idx > len(configs) {
		common.LogError("Choix de configuration invalide pour la surveillance: %s", idxStr)
		fmt.Printf("%sChoix invalide.%s\n", colorRed, colorReset)
		return
	}
	
	config := configs[idx-1]
	common.LogInfo("Démarrage de la surveillance pour la configuration: %s.", config.Name)
	fmt.Printf("Démarrage de la surveillance du répertoire: %s\n", config.SourcePath)
	
	// Option pour la durée de surveillance
	durationStr := readInput("Durée de surveillance en minutes (0 pour mode continu, Ctrl+C pour arrêter): ")
	duration := 0
	if dur, err := strconv.Atoi(durationStr); err == nil && dur > 0 {
		duration = dur
	} else {
		common.LogInfo("Durée de surveillance invalide ou continue sélectionnée: %s", durationStr)
	}
	
	if duration > 0 {
		common.LogInfo("Surveillance pendant %d minutes pour %s.", duration, config.Name)
		fmt.Printf("Surveillance pendant %d minutes...\n", duration)
		
		// Créer un canal pour communiquer avec la goroutine de surveillance
		done := make(chan bool)
		
		// Lancer la surveillance dans une goroutine
		go func() {
			if err := watch.StartWatchWithCallback(config, done); err != nil {
				common.LogError("Erreur lors de la surveillance avec callback pour %s: %v", config.Name, err)
				fmt.Printf("%sErreur lors de la surveillance: %v%s\n", colorRed, err, colorReset)
			}
		}()
		
		// Attendre la durée spécifiée
		time.Sleep(time.Duration(duration) * time.Minute)
		
		// Signaler l'arrêt
		done <- true
		
		common.LogInfo("Surveillance terminée après %d minutes pour %s.", duration, config.Name)
		fmt.Printf("\n%sSurveillance terminée après %d minutes.%s\n", colorGreen, duration, colorReset)
	} else {
		common.LogInfo("Surveillance en mode continu pour %s.", config.Name)
		// Mode continu avec prompt pour retourner au menu
		fmt.Println("Mode surveillance continue. Appuyez sur Ctrl+C pour arrêter.")
		fmt.Println("Les sauvegardes continueront même si vous quittez ce prompt.")
		
		// Lancer la surveillance en arrière-plan
		go watch.StartWatch(config)
		
		// Attendre que l'utilisateur appuie sur Entrée pour revenir au menu
		readInput("\nAppuyez sur Entrée pour revenir au menu principal (la surveillance continuera en arrière-plan)...")
	}
}

// restoreBackupInteractive permet de restaurer une sauvegarde
func restoreBackupInteractive() {
	common.LogInfo("Début de la restauration interactive.")
	fmt.Printf("%sRestauration d'une sauvegarde%s\n\n", colorBold, colorReset)

	backups, err := common.ListBackups()
	if err != nil {
		common.LogError("Erreur lors de la récupération des sauvegardes pour restauration: %v", err)
		fmt.Printf("%sErreur lors de la récupération des sauvegardes: %v%s\n", colorRed, err, colorReset)
		return
	}

	if len(backups) == 0 {
		common.LogWarning("Aucune sauvegarde disponible pour la restauration.")
		fmt.Printf("%sAucune sauvegarde disponible.%s\n", colorYellow, colorReset)
		return
	}

	fmt.Println("Sauvegardes disponibles:")
	for i, b := range backups {
		timeStr := b.Time.Format("02/01/2006 15:04:05")
		fmt.Printf("%d. %s (%s) - %s\n", i+1, b.Name, b.SourcePath, timeStr)
	}

	idxStr := readInput("Sélectionnez une sauvegarde (numéro): ")
	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 1 || idx > len(backups) {
		common.LogError("Choix de sauvegarde invalide pour la restauration: %s", idxStr)
		fmt.Printf("%sChoix invalide.%s\n", colorRed, colorReset)
		return
	}

	backup := backups[idx-1]

	targetPath := readAndValidateInput("Chemin de destination (vide pour restaurer à l'emplacement d'origine): ", common.IsValidPath, "Chemin invalide ou non sécurisé.") // Utilisation de common.IsValidPath
	if targetPath == "" {
		targetPath = backup.SourcePath
	}

	// SECURITY: Vérifier si le chemin de destination est autorisé
	if !common.AppConfig.Security.IsPathAllowed(targetPath) {
		common.LogSecurity("Tentative de restauration vers un chemin non autorisé: %s", targetPath)
		fmt.Printf("%sErreur: Le chemin de destination '%s' n'est pas autorisé dans la configuration de sécurité.%s\n", colorRed, targetPath, colorReset)
		return
	}

	// Vérifier la destination
	if common.DirExists(targetPath) {
		common.LogWarning("Répertoire de destination '%s' existe déjà. Demande de confirmation pour écrasement.", targetPath)
		overwriteStr := readInput(fmt.Sprintf("Le répertoire '%s' existe déjà. Écraser? (o/n): ", targetPath))
		if strings.ToLower(overwriteStr) != "o" {
			common.LogInfo("Restauration annulée par l'utilisateur.")
			fmt.Println("Restauration annulée.")
			return
		}
	}

	common.LogInfo("Restauration de la sauvegarde '%s' vers '%s'.", backup.Name, targetPath)
	fmt.Printf("Restauration de la sauvegarde '%s' vers '%s'...\n", backup.Name, targetPath)

	if err := restore.RestoreBackup(backup.ID, targetPath); err != nil {
		common.LogError("Erreur lors de la restauration de %s: %v", backup.Name, err)
		fmt.Printf("%sErreur lors de la restauration: %v%s\n", colorRed, err, colorReset)
		return
	}

	common.LogInfo("Restauration terminée avec succès pour %s.", backup.Name)
	fmt.Printf("%sRestauration terminée avec succès.%s\n", colorGreen, colorReset)
}

// manageBackupsInteractive permet de gérer les sauvegardes
func manageBackupsInteractive() {
	common.LogInfo("Début de la gestion interactive des sauvegardes.")
	for {
		clearScreen()
		fmt.Printf("%sGestion des sauvegardes%s\n\n", colorBold, colorReset)
		
		fmt.Printf("  %s1.%s Lister les sauvegardes\n", colorGreen, colorReset)
		fmt.Printf("  %s2.%s Supprimer une sauvegarde\n", colorGreen, colorReset)
		fmt.Printf("  %s3.%s Nettoyer les anciennes sauvegardes\n", colorGreen, colorReset)
		fmt.Printf("  %s0.%s Retour au menu principal\n", colorGreen, colorReset)
		
		choice := readInput("Votre choix: ")
		
		switch choice {
		case "1":
			listBackups()
		case "2":
			deleteBackupInteractive()
		case "3":
			cleanOldBackups()
		case "0":
			common.LogInfo("Retour au menu principal depuis la gestion des sauvegardes.")
			return
		default:
			common.LogWarning("Option de gestion des sauvegardes non valide: %s", choice)
			fmt.Println("Option non valide. Veuillez réessayer.")
		}
		
		readInput("Appuyez sur Entrée pour continuer...")
	}
}

// listBackups affiche la liste des sauvegardes
func listBackups() {
	common.LogInfo("Liste des sauvegardes demandée.")
	backups, err := common.ListBackups()
	if err != nil {
		common.LogError("Erreur lors de la récupération de la liste des sauvegardes: %v", err)
		fmt.Printf("%sErreur lors de la récupération des sauvegardes: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	if len(backups) == 0 {
		common.LogInfo("Aucune sauvegarde disponible à lister.")
		fmt.Printf("%sAucune sauvegarde disponible.%s\n", colorYellow, colorReset)
		return
	}
	
	fmt.Printf("%-20s %-30s %-20s %-10s %-8s\n", "NOM", "CHEMIN SOURCE", "DATE", "TAILLE", "TYPE")
	fmt.Println(strings.Repeat("-", 100))
	
	for _, b := range backups {
		timeStr := b.Time.Format("02/01/2006 15:04")
		sizeStr := formatSize(b.Size)
		typeStr := "Normal"
		if b.IsIncremental {
			typeStr = "Incr."
		}
		if b.Compression {
			typeStr += " (C)"
		}
		
		fmt.Printf("%-20s %-30s %-20s %-10s %-8s\n", 
			truncateString(b.Name, 20),
			truncateString(b.SourcePath, 30),
			timeStr,
			sizeStr,
			typeStr)
	}
	common.LogInfo("Liste des %d sauvegardes affichée.", len(backups))
}

// deleteBackupInteractive permet de supprimer une sauvegarde
func deleteBackupInteractive() {
	common.LogInfo("Début de la suppression interactive de sauvegarde.")
	backups, err := common.ListBackups()
	if err != nil {
		common.LogError("Erreur lors de la récupération des sauvegardes pour suppression: %v", err)
		fmt.Printf("%sErreur lors de la récupération des sauvegardes: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	if len(backups) == 0 {
		common.LogWarning("Aucune sauvegarde disponible à supprimer.")
		fmt.Printf("%sAucune sauvegarde disponible.%s\n", colorYellow, colorReset)
		return
	}
	
	fmt.Println("Sauvegardes disponibles:")
	for i, b := range backups {
		timeStr := b.Time.Format("02/01/2006 15:04:05")
		fmt.Printf("%d. %s (%s) - %s\n", i+1, b.Name, b.SourcePath, timeStr)
	}
	
	idxStr := readInput("Sélectionnez une sauvegarde à supprimer (numéro): ")
	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 1 || idx > len(backups) {
		common.LogError("Choix de sauvegarde invalide pour suppression: %s", idxStr)
		fmt.Printf("%sChoix invalide.%s\n", colorRed, colorReset)
		return
	}
	
	backup := backups[idx-1]
	
	confirmStr := readInput(fmt.Sprintf("Êtes-vous sûr de vouloir supprimer '%s'? (o/n): ", backup.Name))
	if strings.ToLower(confirmStr) != "o" {
		common.LogInfo("Suppression annulée par l'utilisateur pour la sauvegarde: %s.", backup.Name)
		fmt.Println("Suppression annulée.")
		return
	}
	
	deleteBackup(backup.ID)
	common.LogInfo("Demande de suppression de la sauvegarde %s.", backup.ID)
}

// deleteBackup supprime une sauvegarde
func deleteBackup(id string) {
	common.LogInfo("Tentative de suppression de la sauvegarde avec ID: %s", id)
	fmt.Printf("Suppression de la sauvegarde %s...\n", id)
	
	err := common.DeleteBackup(id)
	if err != nil {
		common.LogError("Erreur lors de la suppression de la sauvegarde %s: %v", id, err)
		fmt.Printf("%sErreur lors de la suppression: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	common.LogInfo("Sauvegarde %s supprimée avec succès.", id)
	fmt.Printf("%sSuppression terminée avec succès.%s\n", colorGreen, colorReset)
}

// cleanOldBackups nettoie les anciennes sauvegardes selon la politique de rétention
func cleanOldBackups() {
	common.LogInfo("Début du nettoyage des anciennes sauvegardes.")
	fmt.Println("Nettoyage des anciennes sauvegardes...")
	// TODO: Implémenter la logique de nettoyage des anciennes sauvegardes ici.
	// Cette fonctionnalité est un placeholder et doit être développée ultérieurement.
	common.LogInfo("Nettoyage des anciennes sauvegardes terminé.")
	fmt.Printf("%sNettoyage terminé.%s\n", colorGreen, colorReset)
}

// checkDependencies vérifie et installe les dépendances nécessaires
func checkDependencies() {
	common.LogInfo("Vérification des dépendances.")
	fmt.Printf("%sVérification des dépendances%s\n\n", colorBold, colorReset)
	
	deps := []struct {
		command string
		pkg     string
		desc    string
	}{
		{"rsync", "rsync", "Outil de synchronisation de fichiers"},
		{"inotifywait", "inotify-tools", "Surveillance des fichiers et répertoires"},
		{"jq", "jq", "Traitement JSON"},
		{"tar", "tar", "Compression et archivage"},
	}
	
	for _, dep := range deps {
		fmt.Printf("Vérification de %s... ", dep.command)
		
		if common.IsCommandAvailable(dep.command) {
			common.LogInfo("Dépendance %s: OK.", dep.command)
			fmt.Printf("%sOK%s\n", colorGreen, colorReset)
			continue
		}
		
		common.LogWarning("Dépendance %s: Non trouvée.", dep.command)
		fmt.Printf("%sNon trouvé%s\n", colorYellow, colorReset)
		installStr := readInput(fmt.Sprintf("Installer %s (%s)? (o/n): ", dep.pkg, dep.desc))
		
		if strings.ToLower(installStr) != "o" {
			common.LogInfo("Installation de %s ignorée par l'utilisateur.", dep.pkg)
			fmt.Println("Installation ignorée.")
			continue
		}
		
		common.LogInfo("Installation de %s...", dep.pkg)
		fmt.Printf("Installation de %s...\n", dep.pkg)
		if err := common.EnsureDependency(dep.command, dep.pkg); err != nil {
			common.LogError("Erreur lors de l'installation de %s: %v", dep.pkg, err)
			fmt.Printf("%sErreur lors de l'installation: %v%s\n", colorRed, err, colorReset)
		} else {
			common.LogInfo("%s installé avec succès.", dep.pkg)
			fmt.Printf("%s%s installé avec succès.%s\n", colorGreen, dep.pkg, colorReset)
		}
	}
	common.LogInfo("Vérification des dépendances terminée.")
}

// Fonctions utilitaires
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// displayHeader affiche l'en-tête de l'application
func displayHeader() {
	// Couleurs du rainbow flag pour la pride (6 couleurs exactement)
	prideColors := []string{
		"\033[38;5;196m", // Rouge (en haut)
		"\033[38;5;208m", // Orange
		"\033[38;5;226m", // Jaune
		"\033[38;5;46m",  // Vert
		"\033[38;5;27m",  // Bleu
		"\033[38;5;129m", // Violet (en bas)
	}

	// Essayer de lire le fichier banner.txt
	execPath, err := os.Executable()
	var bannerPath string
	if err == nil {
		// Chercher le banner par rapport à l'emplacement de l'exécutable
		bannerPath = filepath.Join(filepath.Dir(execPath), "..", "banner.txt")
	} else {
		common.LogError("Impossible d'obtenir le chemin de l'exécutable: %v", err)
	}
	
	// Si on ne trouve pas le banner à partir de l'exécutable, essayer dans le répertoire actuel
	if _, err := os.Stat(bannerPath); os.IsNotExist(err) {
		common.LogWarning("Fichier banner.txt non trouvé à %s. Essai dans le répertoire courant.", bannerPath)
		// Essayer dans le répertoire projet
		bannerPath = filepath.Join(".", "banner.txt")
		// Si toujours pas trouvé, utiliser un chemin absolu pour le développement
		if _, err := os.Stat(bannerPath); os.IsNotExist(err) {
			common.LogWarning("Fichier banner.txt non trouvé dans le répertoire courant. Utilisation du chemin de développement.")
			bannerPath = "/home/noziop/projects/s4v3my4ss/Projet Go/banner.txt"
		}
	}

	// Lire le fichier banner s'il existe
	if bannerContent, err := os.ReadFile(bannerPath); err == nil {
		common.LogInfo("Fichier banner.txt lu avec succès depuis %s.", bannerPath)
		// Convertir le contenu en string et le diviser en lignes
		bannerLines := strings.Split(string(bannerContent), "\n")
		
		// Affichage avec couleurs du drapeau pride (1 couleur pour 2 lignes)
		fmt.Print(colorBold) // Texte en gras
		for i, line := range bannerLines {
			if i < len(bannerLines) {
				// Déterminer l'index de couleur (une couleur pour deux lignes)
				colorIndex := i / 2
				if colorIndex >= len(prideColors) {
					colorIndex = len(prideColors) - 1
				}
				
				// Afficher la ligne avec la couleur correspondante
				fmt.Print(prideColors[colorIndex], line, "\n")
			}
		}
		fmt.Print(colorReset) // Réinitialiser la couleur
	} else {
		common.LogError("Impossible de lire le fichier banner.txt: %v. Utilisation de la bannière de secours.", err)
		// Bannière de secours si le fichier n'est pas trouvé
		fmt.Print(colorBold)
		fmt.Println("  ___ _ _  _               __  __  ___ ")
		fmt.Println(" / __| | \\| |  ___  /\\ /\\ /__\\/__\\|_  )")
		fmt.Println(" \\__ \\ | .` | / -_)/ _  //_\\ / \\/ / / / ")
		fmt.Println(" |___/_|_|\\|_| \\___|\\__,_/\\__/\\__//___| ")
		fmt.Print(colorReset)
	}

	fmt.Printf("%sSystème de Sauvegarde et Restauration Automatique%s\n\n", colorBold, colorReset)
}

func displayMainMenu() {
	fmt.Printf("%sMENU PRINCIPAL%s\n", colorBold, colorReset)
	fmt.Printf("  %s1.%s Configurer une nouvelle sauvegarde\n", colorGreen, colorReset)
	fmt.Printf("  %s2.%s Démarrer la surveillance d'un répertoire\n", colorGreen, colorReset)
	fmt.Printf("  %s3.%s Restaurer une sauvegarde\n", colorGreen, colorReset)
	fmt.Printf("  %s4.%s Gérer les sauvegardes existantes\n", colorGreen, colorReset)
	fmt.Printf("  %s5.%s Vérifier/installer les dépendances\n", colorGreen, colorReset)
	fmt.Printf("  %s6.%s Rechercher des serveurs rsync sur le réseau\n", colorGreen, colorReset)
	fmt.Printf("  %s7.%s Gérer la configuration\n", colorGreen, colorReset)
	fmt.Printf("  %s0.%s Quitter\n", colorGreen, colorReset)
	fmt.Println()
}

func readInput(prompt string) string {
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}

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

// showMainMenu affiche le menu principal
func ShowMainMenu() {
	common.LogInfo("Affichage du menu principal.")
	for {
		clearScreen()
		displayHeader()
		
		fmt.Printf("\n%sMenu Principal%s\n\n", colorBold, colorReset)
		fmt.Println("1. Configurer une nouvelle sauvegarde")
		fmt.Println("2. Lancer une sauvegarde manuelle")
		fmt.Println("3. Restaurer une sauvegarde")
		fmt.Println("4. Surveiller un répertoire")
		fmt.Println("5. Gérer les sauvegardes")
		fmt.Println("6. Chercher des serveurs rsync sur le réseau")
		fmt.Println("7. Gérer la configuration") // Ajout de l'option manquante
		fmt.Println("0. Quitter")
		
		fmt.Print("\nVotre choix: ")
		choice := readInput("")
		
		switch choice {
		case "1":
			configureBackup()
		case "2":
			// manualBackup - à implémenter plus tard
			common.LogInfo("Fonctionnalité de sauvegarde manuelle non implémentée.")
			fmt.Println("Fonction non implémentée")
		case "3":
			restoreBackupInteractive()
		case "4":
			watchDirectoryInteractive()
		case "5":
			manageBackupsInteractive()
		case "6":
			discoverRsyncServers()
		case "7": // Ajout du case pour l'option manquante
			manageConfiguration()
		case "0":
			common.LogInfo("Quitter l'application depuis le menu principal.")
			fmt.Println("Au revoir !")
			return // Suppression de readInput ici
		default:
			common.LogWarning("Option de menu principal non valide: %s", choice)
			fmt.Printf("\n%sOption invalide. Appuyez sur Entrée pour continuer...%s", colorRed, colorReset)
			readInput("")
		}
	}
}

// discoverRsyncServers recherche et configure les serveurs rsync sur le réseau
func discoverRsyncServers() {
	common.LogInfo("Début de la découverte des serveurs rsync.")
	clearScreen()
	fmt.Printf("%sRecherche de serveurs rsync sur le réseau%s\n\n", colorBold, colorReset)
	
	// Demander le sous-réseau à scanner
	fmt.Println("Entrez le sous-réseau à scanner (format CIDR, par exemple 192.168.0.0/24)")
	fmt.Println("ou laissez vide pour utiliser 192.168.0.0/24:")
	
	subnetCIDR := readAndValidateInput("Sous-réseau: ", common.IsValidSubnet, "Format CIDR invalide (ex: 192.168.0.0/24).") // Utilisation de common.IsValidSubnet
	if subnetCIDR == "" {
		subnetCIDR = "192.168.0.0/24"
		common.LogInfo("Sous-réseau par défaut utilisé pour la découverte: %s", subnetCIDR)
	}
	
	// Afficher un message d'attente
	fmt.Printf("\nRecherche de serveurs rsync sur %s...\n", subnetCIDR)
	fmt.Println("Cela peut prendre jusqu'à 30 secondes. Veuillez patienter...")
	
	// Créer l'objet de découverte
	discovery := wrappers.NewRsyncDiscovery()
	
	// Scanner le réseau (timeout de 30 secondes)
	servers := discovery.ScanNetwork(subnetCIDR, 30)
	
	if len(servers) == 0 {
		common.LogInfo("Aucun serveur rsync trouvé sur le réseau %s.", subnetCIDR)
		fmt.Printf("\n%sAucun serveur rsync trouvé sur le réseau.%s\n", colorYellow, colorReset)
		fmt.Println("Vérifiez que:")
		fmt.Println("1. Le service rsync est activé sur votre NAS ou serveur")
		fmt.Println("2. Le port rsync (873) est ouvert dans le pare-feu")
		fmt.Println("3. Le sous-réseau spécifié est correct")
		return
	}
	
	common.LogInfo("%d serveurs rsync trouvés sur le réseau %s.", len(servers), subnetCIDR)
	// Afficher les serveurs trouvés
	fmt.Printf("\n%s%d serveurs rsync trouvés:%s\n\n", colorGreen, len(servers), colorReset)
	
	for i, server := range servers {
		serverName := server.IP
		if server.Hostname != "" {
			serverName = server.Hostname
		}
		
		fmt.Printf("%d. %s (%s)\n", i+1, serverName, server.IP)
		
		if server.SSHPort > 0 {
			fmt.Printf("   Port SSH détecté: %d (connexion chiffrée disponible)\n", server.SSHPort)
		}
		
		if len(server.Modules) > 0 {
			fmt.Println("   Modules disponibles:")
			for _, module := range server.Modules {
				fmt.Printf("   - %s: %s\n", module.Name, module.Description)
			}
		}
		
		fmt.Println()
	}
	
	// Demander à l'utilisateur de choisir un serveur
	choiceStr := readInput("\nChoisissez un serveur à configurer (0 pour annuler): ")
	choice, err := strconv.Atoi(choiceStr)
	
	if err != nil || choice < 1 || choice > len(servers) {
		if choiceStr != "0" {
			common.LogWarning("Choix de serveur invalide pour la configuration: %s", choiceStr)
			fmt.Printf("\n%sChoix invalide.%s\n", colorRed, colorReset)
		}
		return
	}
	
	// Configurer le serveur sélectionné
	selectedServer := servers[choice-1]
	common.LogInfo("Serveur %s sélectionné pour la configuration.", selectedServer.IP)
	configureRsyncServer(selectedServer)
}

// configureRsyncServer configure un serveur rsync sélectionné
func configureRsyncServer(server wrappers.RsyncServer) {
	common.LogInfo("Début de la configuration du serveur rsync: %s.", server.IP)
	clearScreen()
	serverName := server.IP
	if server.Hostname != "" {
		serverName = server.Hostname
	}
	
	fmt.Printf("%sConfiguration du serveur rsync: %s (%s)%s\n\n", 
		colorBold, serverName, server.IP, colorReset)
	
	// Demander un nom pour le serveur
	defaultName := serverName
	if server.Hostname != "" {
		// Extraire le nom d'hôte sans le domaine si possible
		parts := strings.Split(server.Hostname, ".")
		defaultName = parts[0]
	}
	
	name := readAndValidateInput(fmt.Sprintf("Nom du serveur (défaut: %s): ", defaultName), common.IsValidName, "Nom invalide.") // Utilisation de common.IsValidName
	if name == "" {
		name = defaultName
	}
	
	// Demander le nom d'utilisateur
	username := readAndValidateInput("Nom d'utilisateur pour la connexion: ", common.IsValidName, "Nom d'utilisateur invalide.") // Utilisation de common.IsValidName
	
	// Si des modules sont disponibles, demander lequel utiliser
	defaultModule := ""
	if len(server.Modules) > 0 {
		common.LogInfo("Modules disponibles pour %s: %v", server.IP, server.Modules)
		fmt.Println("\nModules disponibles:")
		for i, module := range server.Modules {
			fmt.Printf("%d. %s: %s\n", i+1, module.Name, module.Description)
		}
		
		moduleChoice := readInput("\nChoisissez un module par défaut (0 pour aucun): ")
		
		if moduleIdx, err := strconv.Atoi(moduleChoice); err == nil && moduleIdx > 0 && moduleIdx <= len(server.Modules) {
			defaultModule = server.Modules[moduleIdx-1].Name
			common.LogInfo("Module par défaut sélectionné: %s", defaultModule)
		} else {
			common.LogWarning("Choix de module invalide ou aucun module sélectionné: %s", moduleChoice)
		}
	}
	
	// Créer la configuration du serveur
	rsyncModuleNames := make([]string, 0, len(server.Modules))
	for _, module := range server.Modules {
		rsyncModuleNames = append(rsyncModuleNames, module.Name)
	}
	
	serverConfig := common.RsyncServerConfig{
		Name:          name,
		IP:            server.IP,
		Port:          server.Port,
		SSHPort:       server.SSHPort,
		Username:      username,
		Modules:       rsyncModuleNames,
		DefaultModule: defaultModule,
	}
	
	// Enregistrer le serveur dans la configuration
	if err := common.AddRsyncServer(serverConfig); err != nil {
		common.LogError("Erreur lors de la sauvegarde de la configuration du serveur rsync %s: %v", serverConfig.Name, err)
		fmt.Printf("\n%sErreur lors de la sauvegarde de la configuration: %v%s\n", colorRed, err, colorReset)
	} else {
		common.LogInfo("Serveur rsync %s configuré avec succès.", serverConfig.Name)
		fmt.Printf("\n%sServeur rsync %s configuré avec succès!%s\n", colorGreen, name, colorReset)
	}
	
	// Proposer de configurer une sauvegarde vers ce serveur
	configBackup := readInput("\nVoulez-vous configurer une sauvegarde vers ce serveur maintenant? (o/n): ")
	
	if strings.ToLower(configBackup) == "o" {
		common.LogInfo("Configuration d'une sauvegarde distante vers %s demandée.", serverConfig.Name)
		configureRemoteBackup(serverConfig)
	} else {
		common.LogInfo("Configuration d'une sauvegarde distante vers %s annulée.", serverConfig.Name)
	}
}

// configureRemoteBackup configure une sauvegarde vers un serveur rsync distant
func configureRemoteBackup(serverConfig common.RsyncServerConfig) {
	common.LogInfo("Début de la configuration d'une sauvegarde distante vers %s.", serverConfig.Name)
	clearScreen()
	fmt.Printf("%sConfiguration d'une sauvegarde vers %s%s\n\n", colorBold, serverConfig.Name, colorReset)
	
	name := readAndValidateInput("Nom de la sauvegarde: ", common.IsValidName, "Nom invalide.") // Utilisation de common.IsValidName
	sourcePath := readAndValidateInput("Chemin du répertoire à surveiller: ", common.IsValidPath, "Chemin invalide ou non sécurisé.") // Utilisation de common.IsValidPath
	
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
		fmt.Printf("%sErreur: Le répertoire '%s' n'existe pas.%s\n", colorRed, sourcePath, colorReset)
		return
	}
	
	// Sélectionner le module à utiliser
	module := serverConfig.DefaultModule
	if len(serverConfig.Modules) > 0 {
		common.LogInfo("Modules disponibles pour la sauvegarde distante: %v", serverConfig.Modules)
		fmt.Println("\nModules disponibles:")
		for i, mod := range serverConfig.Modules {
			fmt.Printf("%d. %s\n", i+1, mod)
		}
		
		moduleChoice := readInput("\nChoisissez un module par défaut (0 pour aucun): ")
		
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
	incrementalStr := readInput("Activer les sauvegardes incrémentales? (o/n): ")
	incremental := strings.ToLower(incrementalStr) == "o"
	
	// Compression
	compressStr := readInput("Activer la compression? (o/n): ")
	compression := strings.ToLower(compressStr) == "o"
	
	// Exclusion de fichiers/dossiers
	excludeDirsStr := readInput("\nRépertoires à exclure (séparés par des virgules): ")
	var excludeDirs []string
	if excludeDirsStr != "" {
		excludeDirs = strings.Split(excludeDirsStr, ",")
		for i, dir := range excludeDirs {
			excludeDirs[i] = strings.TrimSpace(dir)
			if !common.IsValidExcludePattern(excludeDirs[i]) { // Utilisation de common.IsValidExcludePattern
				common.LogError("Modèle d'exclusion de répertoire invalide pour la sauvegarde distante: %s", excludeDirs[i])
				fmt.Printf("%sModèle d'exclusion de répertoire invalide: %s%s\n", colorRed, excludeDirs[i], colorReset)
				return
			}
		}
	} else {
		excludeDirs = commonExcludeDirs
	}
	
	excludeFilesStr := readInput("Fichiers à exclure (séparés par des virgules): ")
	var excludeFiles []string
	if excludeFilesStr != "" {
		excludeFiles = strings.Split(excludeFilesStr, ",")
		for i, file := range excludeFiles {
			excludeFiles[i] = strings.TrimSpace(file)
			if !common.IsValidExcludePattern(excludeFiles[i]) { // Utilisation de common.IsValidExcludePattern
				common.LogError("Modèle d'exclusion de fichier invalide pour la sauvegarde distante: %s", excludeFiles[i])
				fmt.Printf("%sModèle d'exclusion de fichier invalide: %s%s\n", colorRed, excludeFiles[i], colorReset)
				return
			}
		}
	} else {
		excludeFiles = commonExcludeFiles
	}
	
	// Intervalle de sauvegarde automatique
	intervalStr := readInput("Intervalle de sauvegarde en minutes (0 pour désactiver): ")
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
		fmt.Printf("%sErreur lors de l'enregistrement de la configuration: %v%s\n", colorRed, err, colorReset)
		common.AppConfig.BackupDestination = prevDestination // Restaurer l'ancienne destination
	} else {
		common.LogInfo("Configuration de sauvegarde distante %s enregistrée avec succès.", backupConfig.Name)
		if err := common.SaveConfig(common.AppConfig); err != nil {
			common.LogError("Erreur lors de la mise à jour de la destination de sauvegarde distante: %v", err)
			fmt.Printf("%sErreur lors de la mise à jour de la destination: %v%s\n", colorRed, err, colorReset)
		} else {
			common.LogInfo("Destination de sauvegarde distante mise à jour avec succès.")
			fmt.Printf("\n%sConfiguration de sauvegarde vers %s enregistrée avec succès!%s\n", 
				colorGreen, serverConfig.Name, colorReset)
			
			// Proposer de faire une sauvegarde immédiate
			doBackupNow := readInput("\nVoulez-vous effectuer une sauvegarde immédiate? (o/n): ")
			if strings.ToLower(doBackupNow) == "o" {
				common.LogInfo("Sauvegarde immédiate vers %s demandée.", serverConfig.Name)
				fmt.Printf("\nDémarrage de la sauvegarde vers %s...\n", serverConfig.Name)
				
				// Générer un ID unique pour la sauvegarde
				backupID := generateBackupID(name)
				
				// Utiliser la fonction RsyncBackup pour effectuer la sauvegarde
				if err := wrappers.RsyncBackup(sourcePath, destination, excludeDirs, excludeFiles, compression, &serverConfig); err != nil {
					common.LogError("Erreur lors de la sauvegarde immédiate vers %s: %v", serverConfig.Name, err)
					fmt.Printf("%sErreur lors de la sauvegarde: %v%s\n", colorRed, err, colorReset)
					return
				}
				
				// Obtenir la taille du répertoire source comme approximation
				size, err := getDirSize(sourcePath)
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
					fmt.Printf("%sErreur lors de l'enregistrement des métadonnées: %v%s\n", colorRed, err, colorReset)
					return
				}
				
				common.LogInfo("Sauvegarde immédiate vers %s terminée avec succès.", serverConfig.Name)
				fmt.Printf("%sSauvegarde vers %s terminée avec succès!%s\n", colorGreen, serverConfig.Name, colorReset)
			} else {
				common.LogInfo("Sauvegarde immédiate annulée par l'utilisateur.")
			}
		}
	}
}

// generateBackupID génère un ID unique pour une sauvegarde
func generateBackupID(name string) string {
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s-%s", name, timestamp)
}

// getDirSize calcule la taille totale d'un répertoire en octets
func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			common.LogError("Erreur lors du calcul de la taille du répertoire %s: %v", path, err)
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// manageConfiguration permet de gérer la configuration de l'application
func manageConfiguration() {
	common.LogInfo("Début de la gestion de la configuration.")
	for {
		clearScreen()
		fmt.Printf("%sGestion de la Configuration%s\n\n", colorBold, colorReset)
		
		fmt.Printf("  %s1.%s Afficher la configuration complète\n", colorGreen, colorReset)
		fmt.Printf("  %s2.%s Modifier les répertoires sauvegardés\n", colorGreen, colorReset)
		fmt.Printf("  %s3.%s Modifier/supprimer les serveurs rsync\n", colorGreen, colorReset)
		fmt.Printf("  %s4.%s Modifier la politique de rétention\n", colorGreen, colorReset)
		fmt.Printf("  %s5.%s Gérer les destinations de sauvegarde\n", colorGreen, colorReset)
		fmt.Printf("  %s6.%s Modifier la destination principale\n", colorGreen, colorReset)
		fmt.Printf("  %s0.%s Retour au menu principal\n", colorGreen, colorReset)
		
		choice := readInput("Votre choix: ")
		
		switch choice {
		case "1":
			displayFullConfig()
		case "2":
			manageBackupDirectories()
		case "3":
			manageRsyncServers()
		case "4":
			manageRetentionPolicy()
		case "5":
			manageBackupDestinations()
		case "6":
			changeBackupDestination()
		case "0":
			common.LogInfo("Retour au menu principal depuis la gestion de la configuration.")
			return // Suppression de readInput ici
		default:
			common.LogWarning("Option de gestion de configuration non valide: %s", choice)
			fmt.Println("Option non valide. Veuillez réessayer.")
		}
		
		readInput("Appuyez sur Entrée pour continuer...")
	}
}

// displayFullConfig affiche la configuration complète de l'application
func displayFullConfig() {
	common.LogInfo("Affichage de la configuration complète.")
	clearScreen()
	fmt.Printf("%sConfiguration Complète de l'Application%s\n\n", colorBold, colorReset)
	
	// Afficher les informations générales
	fmt.Printf("%sInformations générales:%s\n", colorBold, colorReset)
	fmt.Printf("Destination des sauvegardes: %s\n", common.AppConfig.BackupDestination)
	fmt.Printf("Dernière mise à jour: %s\n\n", common.AppConfig.LastUpdate.Format("02/01/2006 15:04:05"))
	
	// Afficher la politique de rétention
	fmt.Printf("%sPolitique de rétention:%s\n", colorBold, colorReset)
	fmt.Printf("Conservation quotidienne: %d jours\n", common.AppConfig.RetentionPolicy.KeepDaily)
	fmt.Printf("Conservation hebdomadaire: %d semaines\n", common.AppConfig.RetentionPolicy.KeepWeekly)
	fmt.Printf("Conservation mensuelle: %d mois\n\n", common.AppConfig.RetentionPolicy.KeepMonthly)
	
	// Afficher les répertoires sauvegardés
	fmt.Printf("%sRépertoires sauvegardés:%s\n", colorBold, colorReset)
	if len(common.AppConfig.BackupDirs) == 0 {
		common.LogInfo("Aucun répertoire de sauvegarde configuré à afficher.")
		fmt.Printf("%sAucun répertoire configuré.%s\n\n", colorYellow, colorReset)
	} else {
		common.LogInfo("Affichage des %d répertoires de sauvegarde configurés.", len(common.AppConfig.BackupDirs))
		for i, dir := range common.AppConfig.BackupDirs {
			fmt.Printf("%d. %s (%s)\n", i+1, dir.Name, dir.SourcePath)
			incr := "Non"
			if dir.IsIncremental {
				incr = "Oui"
			}
			comp := "Non"
			if dir.Compression {
				comp = "Oui"
			}
			fmt.Printf("   Compression: %s, Incrémental: %s, Intervalle: %d min\n", 
				comp, incr, dir.Interval)
			
			// Afficher les exclusions si présentes
			if len(dir.ExcludeDirs) > 0 || len(dir.ExcludeFiles) > 0 {
				fmt.Println("   Exclusions:")
				
				if len(dir.ExcludeDirs) > 0 {
					fmt.Printf("   - Répertoires: %s\n", strings.Join(dir.ExcludeDirs, ", "))
				}
				
				if len(dir.ExcludeFiles) > 0 {
					fmt.Printf("   - Fichiers: %s\n", strings.Join(dir.ExcludeFiles, ", "))
				}
			}
			
			// Afficher le serveur distant si configuré
			if dir.RemoteServer != nil {
				fmt.Printf("   Serveur distant: %s (%s)\n", 
					dir.RemoteServer.Name, dir.RemoteServer.IP)
				if dir.RemoteServer.DefaultModule != "" {
					fmt.Printf("   Module: %s\n", dir.RemoteServer.DefaultModule)
				}
			}
			
			fmt.Println()
		}
	}
	
	// Afficher les serveurs rsync
	fmt.Printf("%sServeurs rsync configurés:%s\n", colorBold, colorReset)
	if len(common.AppConfig.RsyncServers) == 0 {
		common.LogInfo("Aucun serveur rsync configuré à afficher.")
		fmt.Printf("%sAucun serveur rsync configuré.%s\n", colorYellow, colorReset)
	} else {
		common.LogInfo("Affichage des %d serveurs rsync configurés.", len(common.AppConfig.RsyncServers))
		for i, server := range common.AppConfig.RsyncServers {
			fmt.Printf("%d. %s (%s)\n", i+1, server.Name, server.IP)
			fmt.Printf("   Port: %d, Port SSH: %d, Utilisateur: %s\n", 
				server.Port, server.SSHPort, server.Username)
			
			if len(server.Modules) > 0 {
				fmt.Printf("   Modules disponibles: %s\n", strings.Join(server.Modules, ", "))
				if server.DefaultModule != "" {
					fmt.Printf("   Module par défaut: %s\n", server.DefaultModule)
				}
			}
			
			fmt.Println()
		}
	}
	common.LogInfo("Configuration complète affichée.")
}

// manageBackupDirectories permet de gérer les répertoires à sauvegarder
func manageBackupDirectories() {
	common.LogInfo("Début de la gestion des répertoires de sauvegarde.")
	for {
		clearScreen()
		fmt.Printf("%sGestion des répertoires sauvegardés%s\n\n", colorBold, colorReset)
		
		// Afficher les répertoires sauvegardés
		if len(common.AppConfig.BackupDirs) == 0 {
			common.LogInfo("Aucun répertoire de sauvegarde configuré à gérer.")
			fmt.Printf("%sAucun répertoire configuré.%s\n\n", colorYellow, colorReset)
		} else {
			common.LogInfo("Affichage des %d répertoires de sauvegarde configurés pour gestion.", len(common.AppConfig.BackupDirs))
			fmt.Printf("%sRépertoires configurés:%s\n", colorBold, colorReset)
			for i, dir := range common.AppConfig.BackupDirs {
				fmt.Printf("%d. %s (%s)\n", i+1, dir.Name, dir.SourcePath)
				incr := "Non"
				if dir.IsIncremental {
					incr = "Oui"
				}
				comp := "Non"
				if dir.Compression {
					comp = "Oui"
				}
				fmt.Printf("   Compression: %s, Incrémental: %s, Intervalle: %d min\n", 
					comp, incr, dir.Interval)
			}
			fmt.Println()
		}
		
		fmt.Printf("  %s1.%s Ajouter un répertoire\n", colorGreen, colorReset)
		fmt.Printf("  %s2.%s Modifier un répertoire\n", colorGreen, colorReset)
		fmt.Printf("  %s3.%s Supprimer un répertoire\n", colorGreen, colorReset)
		fmt.Printf("  %s0.%s Retour\n\n", colorGreen, colorReset)
		
		choice := readInput("Votre choix: ")
		
		switch choice {
		case "1":
			configureBackup() // Utiliser la fonction existante
			common.LogInfo("Ajout d'un répertoire de sauvegarde demandé.")
			return
		case "2":
			editBackupDirectory()
		case "3":
			deleteBackupDirectory()
		case "0":
			common.LogInfo("Retour au menu de gestion de la configuration.")
			return // Suppression de readInput ici
		default:
			common.LogWarning("Option de gestion des répertoires de sauvegarde non valide: %s", choice)
			fmt.Println("Option non valide.")
		}
		
		readInput("Appuyez sur Entrée pour continuer...")
	}
}

// editBackupDirectory permet de modifier un répertoire de sauvegarde existant
func editBackupDirectory() {
	common.LogInfo("Début de la modification d'un répertoire de sauvegarde.")
	if len(common.AppConfig.BackupDirs) == 0 {
		common.LogWarning("Aucun répertoire de sauvegarde à modifier.")
		fmt.Printf("%sAucun répertoire à modifier.%s\n", colorYellow, colorReset)
		return
	}
	
	idxStr := readInput("Numéro du répertoire à modifier: ")
	idx, err := strconv.Atoi(idxStr)
	
	if err != nil || idx < 1 || idx > len(common.AppConfig.BackupDirs) {
		common.LogError("Numéro de répertoire invalide pour modification: %s", idxStr)
		fmt.Printf("%sNuméro invalide.%s\n", colorRed, colorReset)
		return
	}
	
	// Récupérer la configuration à modifier
	dir := common.AppConfig.BackupDirs[idx-1]
	common.LogInfo("Modification du répertoire de sauvegarde: %s.", dir.Name)
	fmt.Printf("%sModification de la configuration '%s'%s\n\n", colorBold, dir.Name, colorReset)
	
	// Permettre de modifier chaque propriété
	fmt.Printf("Nom actuel: %s\n", dir.Name)
	name := readAndValidateInput("Nouveau nom (vide pour garder l'actuel): ", common.IsValidName, "Nom invalide.") // Utilisation de common.IsValidName
	if name == "" {
		name = dir.Name
	}
	
	fmt.Printf("Chemin actuel: %s\n", dir.SourcePath)
	sourcePath := readAndValidateInput("Nouveau chemin (vide pour garder l'actuel): ", common.IsValidPath, "Chemin invalide.") // Utilisation de common.IsValidPath
	if sourcePath == "" {
		sourcePath = dir.SourcePath
	} else {
		// Expandir les chemins relatifs, y compris ~/
		if strings.HasPrefix(sourcePath, "~/") {
			homeDir, err := os.UserHomeDir()
			if err == nil {
				sourcePath = filepath.Join(homeDir, sourcePath[2:])
			} else {
				common.LogError("Impossible d'obtenir le répertoire personnel de l'utilisateur pour la source: %v", err)
			}
		}
		
		// Normaliser et vérifier le chemin
		sourcePath = filepath.Clean(sourcePath)
		if !common.DirExists(sourcePath) {
			common.LogWarning("Répertoire source '%s' n'existe pas lors de la modification.", sourcePath)
			fmt.Printf("%sAttention: Le répertoire '%s' n'existe pas.%s\n", colorYellow, sourcePath, colorReset)
			confirm := readInput("Continuer quand même? (o/n): ")
			if strings.ToLower(confirm) != "o" {
				common.LogInfo("Modification du répertoire annulée par l'utilisateur.")
				return
			}
		}
	}
	
	// Option pour une sauvegarde incrémentale
	incStr := "n"
	if dir.IsIncremental {
		incStr = "o"
	}
	fmt.Printf("Sauvegarde incrémentale actuelle: %s\n", incStr)
	incrementalStr := readInput("Activer les sauvegardes incrémentales? (o/n, vide pour garder l'actuel): ")
	incremental := dir.IsIncremental
	if incrementalStr != "" {
		incremental = strings.ToLower(incrementalStr) == "o"
	}
	
	// Compression
	compStr := "n"
	if dir.Compression {
		compStr = "o"
	}
	fmt.Printf("Compression actuelle: %s\n", compStr)
	compressStr := readInput("Activer la compression? (o/n, vide pour garder l'actuel): ")
	compression := dir.Compression
	if compressStr != "" {
		compression = strings.ToLower(compressStr) == "o"
	}
	
	// Répertoires à exclure
	fmt.Printf("Répertoires exclus actuels: %s\n", strings.Join(dir.ExcludeDirs, ", "))
	excludeDirsStr := readInput("Nouveaux répertoires à exclure (séparés par des virgules, vide pour garder les actuels): ")
	excludeDirs := dir.ExcludeDirs
	if excludeDirsStr != "" {
		excludeDirs = strings.Split(excludeDirsStr, ",")
		for i, dir := range excludeDirs {
			excludeDirs[i] = strings.TrimSpace(dir)
			if !common.IsValidExcludePattern(excludeDirs[i]) { // Utilisation de common.IsValidExcludePattern
				common.LogError("Modèle d'exclusion de répertoire invalide lors de la modification: %s", excludeDirs[i])
				fmt.Printf("%sModèle d'exclusion de répertoire invalide: %s%s\n", colorRed, excludeDirs[i], colorReset)
				return
			}
		}
	}
	
	// Fichiers à exclure
	fmt.Printf("Fichiers exclus actuels: %s\n", strings.Join(dir.ExcludeFiles, ", "))
	excludeFilesStr := readInput("Nouveaux fichiers à exclure (séparés par des virgules, vide pour garder les actuels): ")
	excludeFiles := dir.ExcludeFiles
	if excludeFilesStr != "" {
		excludeFiles = strings.Split(excludeFilesStr, ",")
		for i, file := range excludeFiles {
			excludeFiles[i] = strings.TrimSpace(file)
			if !common.IsValidExcludePattern(excludeFiles[i]) { // Utilisation de common.IsValidExcludePattern
				common.LogError("Modèle d'exclusion de fichier invalide lors de la modification: %s", excludeFiles[i])
				fmt.Printf("%sModèle d'exclusion de fichier invalide: %s%s\n", colorRed, excludeFiles[i], colorReset)
				return
			}
		}
	}
	
	// Intervalle
	fmt.Printf("Intervalle actuel: %d minutes\n", dir.Interval)
	intervalStr := readInput("Nouvel intervalle en minutes (vide pour garder l'actuel): ")
	interval := dir.Interval
	if intervalStr != "" {
		if i, err := strconv.Atoi(intervalStr); err == nil && i >= 0 {
			interval = i
		} else {
			common.LogWarning("Intervalle invalide lors de la modification: %s. Garde l'actuel.", intervalStr)
		}
	}
	
	// Créer la configuration modifiée
	updatedConfig := common.BackupConfig{
		Name:          name,
		SourcePath:    sourcePath,
		Compression:   compression,
		IsIncremental: incremental,
		ExcludeDirs:   excludeDirs,
		ExcludeFiles:  excludeFiles,
		Interval:      interval,
		RemoteServer:  dir.RemoteServer, // Conserver le serveur distant s'il existe
	}
	
	// Mettre à jour la configuration
	common.AppConfig.BackupDirs[idx-1] = updatedConfig
	
	if err := common.SaveConfig(common.AppConfig); err != nil {
		common.LogError("Erreur lors de la mise à jour de la configuration du répertoire %s: %v", name, err)
		fmt.Printf("%sErreur lors de la mise à jour de la configuration: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	common.LogInfo("Configuration du répertoire %s modifiée avec succès.", name)
	fmt.Printf("%sConfiguration '%s' modifiée avec succès.%s\n", colorGreen, name, colorReset)
}

// deleteBackupDirectory permet de supprimer un répertoire de sauvegarde
func deleteBackupDirectory() {
	common.LogInfo("Début de la suppression d'un répertoire de sauvegarde.")
	if len(common.AppConfig.BackupDirs) == 0 {
		common.LogWarning("Aucun répertoire de sauvegarde à supprimer.")
		fmt.Printf("%sAucun répertoire à supprimer.%s\n", colorYellow, colorReset)
		return
	}
	
	idxStr := readInput("Numéro du répertoire à supprimer: ")
	idx, err := strconv.Atoi(idxStr)
	
	if err != nil || idx < 1 || idx > len(common.AppConfig.BackupDirs) {
		common.LogError("Numéro de répertoire invalide pour suppression: %s", idxStr)
		fmt.Printf("%sNuméro invalide.%s\n", colorRed, colorReset)
		return
	}
	
	// Récupérer le nom pour confirmation
	name := common.AppConfig.BackupDirs[idx-1].Name
	
	confirm := readInput(fmt.Sprintf("Êtes-vous sûr de vouloir supprimer la configuration '%s'? (o/n): ", name))
	if strings.ToLower(confirm) != "o" {
		common.LogInfo("Suppression du répertoire annulée par l'utilisateur pour %s.", name)
		fmt.Println("Suppression annulée.")
		return
	}
	
	// Supprimer l'élément
	common.AppConfig.BackupDirs = append(
		common.AppConfig.BackupDirs[:idx-1], 
		common.AppConfig.BackupDirs[idx:]...,
	)
	
	if err := common.SaveConfig(common.AppConfig); err != nil {
		common.LogError("Erreur lors de la suppression de la configuration du répertoire %s: %v", name, err)
		fmt.Printf("%sErreur lors de la mise à jour de la configuration: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	common.LogInfo("Configuration du répertoire %s supprimée avec succès.", name)
	fmt.Printf("%sConfiguration '%s' supprimée avec succès.%s\n", colorGreen, name, colorReset)
}

// manageRsyncServers permet de gérer les serveurs rsync configurés
func manageRsyncServers() {
	common.LogInfo("Début de la gestion des serveurs rsync.")
	for {
		clearScreen()
		fmt.Printf("%sGestion des serveurs rsync%s\n\n", colorBold, colorReset)
		
		// Afficher les serveurs configurés
		if len(common.AppConfig.RsyncServers) == 0 {
			common.LogWarning("Aucun serveur rsync configuré à gérer.")
			fmt.Printf("%sAucun serveur rsync configuré.%s\n\n", colorYellow, colorReset)
		} else {
			common.LogInfo("Affichage des %d serveurs rsync configurés pour gestion.", len(common.AppConfig.RsyncServers))
			fmt.Printf("%sServeurs configurés:%s\n", colorBold, colorReset)
			for i, server := range common.AppConfig.RsyncServers {
				fmt.Printf("%d. %s (%s)\n", i+1, server.Name, server.IP)
				fmt.Printf("   Port: %d, Port SSH: %d, Utilisateur: %s\n", 
					server.Port, server.SSHPort, server.Username)
				
				if len(server.Modules) > 0 {
					fmt.Printf("   Modules: %s\n", strings.Join(server.Modules, ", "))
				}
				fmt.Println()
			}
		}
		
		fmt.Printf("  %s1.%s Rechercher et ajouter un serveur\n", colorGreen, colorReset)
		fmt.Printf("  %s2.%s Modifier un serveur\n", colorGreen, colorReset)
		fmt.Printf("  %s3.%s Supprimer un serveur\n", colorGreen, colorReset)
		fmt.Printf("  %s0.%s Retour\n\n", colorGreen, colorReset)
		
		choice := readInput("Votre choix: ")
		
		switch choice {
		case "1":
			discoverRsyncServers()
			common.LogInfo("Recherche et ajout de serveur rsync demandé.")
			return
		case "2":
			editRsyncServer()
		case "3":
			deleteRsyncServer()
		case "0":
			common.LogInfo("Retour au menu de gestion de la configuration.")
			return // Suppression de readInput ici
		default:
			common.LogWarning("Option de gestion des serveurs rsync non valide: %s", choice)
			fmt.Println("Option non valide. Veuillez réessayer.")
		}
		
		readInput("Appuyez sur Entrée pour continuer...")
	}
}

// editRsyncServer permet de modifier un serveur rsync existant
func editRsyncServer() {
	common.LogInfo("Début de la modification d'un serveur rsync.")
	if len(common.AppConfig.RsyncServers) == 0 {
		common.LogWarning("Aucun serveur rsync à modifier.")
		fmt.Printf("%sAucun serveur à modifier.%s\n", colorYellow, colorReset)
		return
	}
	
	idxStr := readInput("Numéro du serveur à modifier: ")
	idx, err := strconv.Atoi(idxStr)
	
	if err != nil || idx < 1 || idx > len(common.AppConfig.RsyncServers) {
		common.LogError("Numéro de serveur rsync invalide pour modification: %s", idxStr)
		fmt.Printf("%sNuméro invalide.%s\n", colorRed, colorReset)
		return
	}
	
	// Récupérer la configuration à modifier
	server := common.AppConfig.RsyncServers[idx-1]
	common.LogInfo("Modification du serveur rsync: %s.", server.Name)
	fmt.Printf("%sModification du serveur '%s'%s\n\n", colorBold, server.Name, colorReset)
	
	// Permettre de modifier chaque propriété
	fmt.Printf("Nom actuel: %s\n", server.Name)
	name := readAndValidateInput("Nouveau nom (vide pour garder l'actuel): ", common.IsValidName, "Nom invalide.") // Utilisation de common.IsValidName
	if name == "" {
		name = server.Name
	}
	
	fmt.Printf("Adresse IP actuelle: %s\n", server.IP)
	ip := readInput("Nouvelle adresse IP (vide pour garder l'actuelle): ")
	if ip == "" {
		ip = server.IP
	}
	
	fmt.Printf("Nom d'utilisateur actuel: %s\n", server.Username)
	username := readAndValidateInput("Nouveau nom d'utilisateur (vide pour garder l'actuel): ", common.IsValidName, "Nom d'utilisateur invalide.") // Utilisation de common.IsValidName
	if username == "" {
		username = server.Username
	}
	
	fmt.Printf("Port SSH actuel: %d\n", server.SSHPort)
	sshPortStr := readInput("Nouveau port SSH (vide pour garder l'actuel): ")
	sshPort := server.SSHPort
	if sshPortStr != "" {
		if port, err := strconv.Atoi(sshPortStr); err == nil && port > 0 {
			sshPort = port
		} else {
			common.LogWarning("Port SSH invalide: %s. Garde l'actuel.", sshPortStr)
		}
	}
	
	// Si des modules sont disponibles, permettre de modifier le module par défaut
	defaultModule := server.DefaultModule
	if len(server.Modules) > 0 {
		fmt.Printf("Modules disponibles: %s\n", strings.Join(server.Modules, ", "))
		fmt.Printf("Module par défaut actuel: %s\n", defaultModule)
		
		fmt.Println("\nVoulez-vous changer le module par défaut?")
		moduleChoice := readInput("\nChoisissez un module (vide pour garder l'actuel): ")
		
		if moduleChoice != "" {
			if moduleIdx, err := strconv.Atoi(moduleChoice); err == nil {
				if moduleIdx == 0 {
					defaultModule = ""
					common.LogInfo("Module par défaut supprimé pour %s.", name)
				} else if moduleIdx > 0 && moduleIdx <= len(server.Modules) {
					defaultModule = server.Modules[moduleIdx-1]
					common.LogInfo("Module par défaut mis à jour pour %s: %s.", name, defaultModule)
				} else {
					common.LogWarning("Choix de module invalide: %s. Garde l'actuel.", moduleChoice)
				}
			} else {
				common.LogWarning("Choix de module invalide: %s. Garde l'actuel.", moduleChoice)
			}
		}
	}
	
	// Créer la configuration modifiée
	updatedServer := common.RsyncServerConfig{
		Name:          name,
		IP:            ip,
		Port:          server.Port, // Garder le port d'origine
		SSHPort:       sshPort,
		Username:      username,
		Modules:       server.Modules, // Garder les modules d'origine
		DefaultModule: defaultModule,
	}
	
	// Mettre à jour la configuration
	common.AppConfig.RsyncServers[idx-1] = updatedServer
	
	// Mettre à jour également les références dans les configurations de sauvegarde
	for i, dir := range common.AppConfig.BackupDirs {
		if dir.RemoteServer != nil && dir.RemoteServer.Name == server.Name {
			common.LogInfo("Mise à jour de la référence du serveur distant dans la configuration de sauvegarde %s.", dir.Name)
			common.AppConfig.BackupDirs[i].RemoteServer = &updatedServer
		}
	}
	
	if err := common.SaveConfig(common.AppConfig); err != nil {
		common.LogError("Erreur lors de la mise à jour du serveur rsync %s: %v", name, err)
		fmt.Printf("%sErreur lors de la mise à jour de la configuration: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	common.LogInfo("Serveur rsync %s modifié avec succès.", name)
	fmt.Printf("%sServeur '%s' modifié avec succès.%s\n", colorGreen, name, colorReset)
}

// deleteRsyncServer permet de supprimer un serveur rsync
func deleteRsyncServer() {
	common.LogInfo("Début de la suppression d'un serveur rsync.")
	if len(common.AppConfig.RsyncServers) == 0 {
		common.LogWarning("Aucun serveur rsync à supprimer.")
		fmt.Printf("%sAucun serveur à supprimer.%s\n", colorYellow, colorReset)
		return
	}
	
	idxStr := readInput("Numéro du serveur à supprimer: ")
	idx, err := strconv.Atoi(idxStr)
	
	if err != nil || idx < 1 || idx > len(common.AppConfig.RsyncServers) {
		common.LogError("Numéro de serveur rsync invalide pour suppression: %s", idxStr)
		fmt.Printf("%sNuméro invalide.%s\n", colorRed, colorReset)
		return
	}
	
	// Récupérer le nom pour confirmation
	name := common.AppConfig.RsyncServers[idx-1].Name
	
	confirm := readInput(fmt.Sprintf("Êtes-vous sûr de vouloir supprimer le serveur '%s'? (o/n): ", name))
	if strings.ToLower(confirm) != "o" {
		common.LogInfo("Suppression du serveur rsync annulée par l'utilisateur pour %s.", name)
		fmt.Println("Suppression annulée.")
		return
	}
	
	// Vérifier si le serveur est utilisé par des configurations de sauvegarde
	var usedBy []string
	for _, dir := range common.AppConfig.BackupDirs {
		if dir.RemoteServer != nil && dir.RemoteServer.Name == name {
			usedBy = append(usedBy, dir.Name)
		}
	}
	
	if len(usedBy) > 0 {
		common.LogWarning("Tentative de suppression du serveur rsync %s utilisé par les configurations: %v", name, usedBy)
		fmt.Printf("%sAttention: Ce serveur est utilisé par les configurations suivantes:%s\n", 
			colorYellow, colorReset)
		for _, name := range usedBy {
			fmt.Printf("- %s\n", name)
		}
		fmt.Println("La suppression du serveur affectera ces configurations.")
		confirm := readInput("Voulez-vous vraiment continuer? (o/n): ")
		if strings.ToLower(confirm) != "o" {
			common.LogInfo("Suppression du serveur rsync annulée par l'utilisateur (utilisé par d'autres configurations) pour %s.", name)
			fmt.Println("Suppression annulée.")
			return
		}
	}

	// Supprimer l'élément
	common.AppConfig.RsyncServers = append(
		common.AppConfig.RsyncServers[:idx-1], 
		common.AppConfig.RsyncServers[idx:]...,
	)
	
	// Mettre à jour les références dans les configurations de sauvegarde
	for i, dir := range common.AppConfig.BackupDirs {
		if dir.RemoteServer != nil && dir.RemoteServer.Name == name {
			common.LogInfo("Mise à jour de la référence du serveur distant dans la configuration de sauvegarde %s.", dir.Name)
			common.AppConfig.BackupDirs[i].RemoteServer = nil
		}
	}
	
	if err := common.SaveConfig(common.AppConfig); err != nil {
		common.LogError("Erreur lors de la suppression du serveur rsync %s: %v", name, err)
		fmt.Printf("%sErreur lors de la mise à jour de la configuration: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	common.LogInfo("Serveur rsync %s supprimé avec succès.", name)
	fmt.Printf("%sServeur '%s' supprimé avec succès.%s\n", colorGreen, name, colorReset)
}

// manageRetentionPolicy permet de modifier la politique de rétention
func manageRetentionPolicy() {
	common.LogInfo("Début de la modification de la politique de rétention.")
	clearScreen()
	fmt.Printf("%sConfiguration de la politique de rétention%s\n\n", colorBold, colorReset)
	
	policy := common.AppConfig.RetentionPolicy
	
	fmt.Printf("Politique actuelle:\n")
	fmt.Printf("- Conservation quotidienne: %d jours\n", policy.KeepDaily)
	fmt.Printf("- Conservation hebdomadaire: %d semaines\n", policy.KeepWeekly)
	fmt.Printf("- Conservation mensuelle: %d mois\n\n", policy.KeepMonthly)
	
	// Permettre de modifier chaque valeur
	keepDailyStr := readInput(fmt.Sprintf("Nombre de sauvegardes quotidiennes à conserver (actuel: %d): ", policy.KeepDaily))
	keepDaily := policy.KeepDaily
	if keepDailyStr != "" {
		if days, err := strconv.Atoi(keepDailyStr); err == nil && days >= 0 {
			keepDaily = days
		} else {
			common.LogWarning("Valeur invalide pour la conservation quotidienne: %s. Garde l'actuelle.", keepDailyStr)
		}
	}
	
	keepWeeklyStr := readInput(fmt.Sprintf("Nombre de sauvegardes hebdomadaires à conserver (actuel: %d): ", policy.KeepWeekly))
	keepWeekly := policy.KeepWeekly
	if keepWeeklyStr != "" {
		if weeks, err := strconv.Atoi(keepWeeklyStr); err == nil && weeks >= 0 {
			keepWeekly = weeks
		} else {
			common.LogWarning("Valeur invalide pour la conservation hebdomadaire: %s. Garde l'actuelle.", keepWeeklyStr)
		}
	}
	
	keepMonthlyStr := readInput(fmt.Sprintf("Nombre de sauvegardes mensuelles à conserver (actuel: %d): ", policy.KeepMonthly))
	keepMonthly := policy.KeepMonthly
	if keepMonthlyStr != "" {
		if months, err := strconv.Atoi(keepMonthlyStr); err == nil && months >= 0 {
			keepMonthly = months
		} else {
			common.LogWarning("Valeur invalide pour la conservation mensuelle: %s. Garde l'actuelle.", keepMonthlyStr)
		}
	}
	
	// Mettre à jour la politique
	common.AppConfig.RetentionPolicy = common.RetentionPolicy{
		KeepDaily:   keepDaily,
		KeepWeekly:  keepWeekly,
		KeepMonthly: keepMonthly,
	}
	
	if err := common.SaveConfig(common.AppConfig); err != nil {
		common.LogError("Erreur lors de la mise à jour de la politique de rétention: %v", err)
		fmt.Printf("%sErreur lors de la mise à jour de la configuration: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	common.LogInfo("Politique de rétention mise à jour avec succès.")
	fmt.Printf("%sPolitique de rétention mise à jour avec succès.%s\n", colorGreen, colorReset)
}

// changeBackupDestination permet de modifier la destination des sauvegardes
func changeBackupDestination() {
	common.LogInfo("Début de la modification de la destination principale des sauvegardes.")
	clearScreen()
	fmt.Printf("%sModification de la destination des sauvegardes%s\n\n", colorBold, colorReset)
	
	fmt.Printf("Destination actuelle: %s\n\n", common.AppConfig.BackupDestination)
	
	// Demander la nouvelle destination
	newDest := readAndValidateInput("Nouvelle destination (vide pour annuler): ", common.IsValidPath, "Chemin invalide.") // Utilisation de common.IsValidPath
	if newDest == "" {
		common.LogInfo("Modification de la destination annulée: chemin vide.")
		fmt.Println("Modification annulée.")
		return
	}
	
	// Expandir les chemins relatifs, y compris ~/
	if strings.HasPrefix(newDest, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			newDest = filepath.Join(homeDir, newDest[2:])
		} else {
			common.LogError("Impossible d'obtenir le répertoire personnel de l'utilisateur pour la nouvelle destination: %v", err)
		}
	}
	
	// Normaliser le chemin
	newDest = filepath.Clean(newDest)
	
	// SECURITY: Vérifier si le chemin de destination est autorisé
	if !common.AppConfig.Security.IsPathAllowed(newDest) {
		common.LogSecurity("Tentative de modification de la destination vers un chemin non autorisé: %s", newDest)
		fmt.Printf("%sErreur: Le chemin de destination '%s' n'est pas autorisé dans la configuration de sécurité.%s\n", colorRed, newDest, colorReset)
		return
	}
	
	// Vérifier/créer le répertoire de destination
	if !common.DirExists(newDest) {
		fmt.Printf("Le répertoire n'existe pas. Voulez-vous le créer? (o/n): ")
		createDest := readInput("")
		if strings.ToLower(createDest) == "o" {
			if err := os.MkdirAll(newDest, 0755); err != nil {
				common.LogError("Erreur lors de la création du répertoire %s: %v", newDest, err)
				fmt.Printf("%sErreur lors de la création du répertoire: %v%s\n", colorRed, err, colorReset)
				return
			}
			common.LogInfo("Répertoire %s créé pour la nouvelle destination.", newDest)
		} else {
			common.LogInfo("Création de répertoire annulée par l'utilisateur pour la nouvelle destination.")
			fmt.Println("Modification annulée.")
			return	
		}
	}
	
	// Mettre à jour la destination
	common.AppConfig.BackupDestination = newDest
	
	if err := common.SaveConfig(common.AppConfig); err != nil {
		common.LogError("Erreur lors de la mise à jour de la destination principale: %v", err)
		fmt.Printf("%sErreur lors de la mise à jour de la configuration: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	common.LogInfo("Destination principale des sauvegardes mise à jour avec succès: %s.", newDest)
	fmt.Printf("%sDestination des sauvegardes mise à jour avec succès.%s\n", colorGreen, colorReset)
}

// manageBackupDestinations permet de gérer les destinations de sauvegarde
func manageBackupDestinations() {
	common.LogInfo("Début de la gestion des destinations de sauvegarde.")
	for {
		clearScreen()
		fmt.Printf("%sGestion des Destinations de Sauvegarde%s\n\n", colorBold, colorReset)
		
		// Afficher la liste des destinations actuelles
		fmt.Println("Destinations configurées:")
		
		if len(common.AppConfig.BackupDestinations) == 0 {
			common.LogInfo("Aucune destination de sauvegarde configurée à gérer.")
			fmt.Printf("%sAucune destination configurée%s\n\n", colorYellow, colorReset)
		} else {
			common.LogInfo("Affichage des %d destinations de sauvegarde configurées.", len(common.AppConfig.BackupDestinations))
			for i, dest := range common.AppConfig.BackupDestinations {
				defaultMark := ""
				if dest.IsDefault {
					defaultMark = fmt.Sprintf(" %s(Par défaut)%s", colorGreen, colorReset)
				}
				
				fmt.Printf("%d. %s (%s)%s\n", i+1, dest.Name, dest.Type, defaultMark)
				fmt.Printf("   Chemin: %s\n", dest.Path)
				
				// Afficher des informations supplémentaires en fonction du type
				if dest.Type == "rsync" && dest.RsyncServer != nil {
					fmt.Printf("   Serveur: %s@%s\n", dest.RsyncServer.Username, dest.RsyncServer.IP)
					if dest.RsyncServer.DefaultModule != "" {
						fmt.Printf("   Module: %s\n", dest.RsyncServer.DefaultModule)
					}
				}
				fmt.Println()
			}
		}
		
		// Menu d'options
		fmt.Printf("\n  %s1.%s Ajouter une destination\n", colorGreen, colorReset)
		fmt.Printf("  %s2.%s Modifier une destination\n", colorGreen, colorReset)
		fmt.Printf("  %s3.%s Supprimer une destination\n", colorGreen, colorReset)
		fmt.Printf("  %s4.%s Définir la destination par défaut\n", colorGreen, colorReset)
		fmt.Printf("  %s0.%s Retour\n\n", colorGreen, colorReset)
		
		choice := readInput("Votre choix: ")
		
		switch choice {
		case "1":
			addBackupDestination()
		case "2":
			editBackupDestination()
		case "3":
			deleteBackupDestination()
		case "4":
			setDefaultDestination()
		case "0":
			common.LogInfo("Retour au menu de gestion de la configuration depuis la gestion des destinations.")
			return // Suppression de readInput ici
		default:
			common.LogWarning("Option de gestion des destinations non valide: %s", choice)
			fmt.Println("Option non valide. Veuillez réessayer.")
			readInput("Appuyez sur Entrée pour continuer...")
		}
	}
}

// addBackupDestination ajoute une nouvelle destination de sauvegarde
func addBackupDestination() {
	common.LogInfo("Début de l'ajout d'une nouvelle destination de sauvegarde.")
	clearScreen()
	fmt.Printf("%sAjout d'une Nouvelle Destination de Sauvegarde%s\n\n", colorBold, colorReset)
	
	// Demander les informations de base
	name := readAndValidateInput("Nom de la destination: ", common.IsValidName, "Nom invalide.") // Utilisation de common.IsValidName
	if name == "" {
		common.LogWarning("Nom de destination vide. Annulation de l'ajout.")
		fmt.Printf("%sLe nom ne peut pas être vide.%s\n", colorRed, colorReset)
		readInput("Appuyez sur Entrée pour continuer...")
		return
	}
	
	// Vérifier si le nom existe déjà
	if _, found := common.GetBackupDestination(name); found {
		common.LogError("Une destination avec le nom '%s' existe déjà.", name)
		fmt.Printf("%sUne destination avec ce nom existe déjà.%s\n", colorRed, colorReset)
		readInput("Appuyez sur Entrée pour continuer...")
		return
	}
	
	// Type de destination
	fmt.Println("\nTypes de destination disponibles:")
	fmt.Printf("  %s1.%s Local (système de fichiers)\n", colorGreen, colorReset)
	fmt.Printf("  %s2.%s Rsync (serveur distant)\n", colorGreen, colorReset)
	
	destType := ""
	typeChoice := readInput("\nType de destination [1]: ")
	if typeChoice == "" || typeChoice == "1" {
		destType = "local"
		common.LogInfo("Type de destination sélectionné: local.")
	} else if typeChoice == "2" {
		destType = "rsync"
		common.LogInfo("Type de destination sélectionné: rsync.")
	} else {
		common.LogWarning("Type de destination non valide: %s. Utilisation du type 'local' par défaut.", typeChoice)
		fmt.Printf("%sType non valide. Utilisation du type 'local' par défaut.%s\n", colorYellow, colorReset)
		destType = "local"
	}
	
	// Informations spécifiques au type
	var path string
	var rsyncServer *common.RsyncServerConfig
	
	if destType == "local" {
		// Pour une destination locale, demander le chemin
		path = readAndValidateInput("Chemin de la destination: ", common.IsValidPath, "Chemin invalide.") // Utilisation de common.IsValidPath
		if path == "" {
			common.LogWarning("Chemin de destination local vide. Annulation de l'ajout.")
			fmt.Printf("%sLe chemin ne peut pas être vide.%s\n", colorRed, colorReset)
			readInput("Appuyez sur Entrée pour continuer...")
			return
		}
		
		// Vérifier si le répertoire existe, sinon proposer de le créer
		if !common.DirExists(path) {
			common.LogInfo("Répertoire de destination local '%s' n'existe pas. Demande de création.", path)
			createDir := readInput(fmt.Sprintf("Le répertoire %s n'existe pas. Voulez-vous le créer? (o/n): ", path))
			if strings.ToLower(createDir) == "o" || strings.ToLower(createDir) == "oui" {
				if err := os.MkdirAll(path, 0755); err != nil {
					common.LogError("Erreur lors de la création du répertoire %s: %v", path, err)
					fmt.Printf("%sErreur lors de la création du répertoire: %v%s\n", colorRed, err, colorReset)
					readInput("Appuyez sur Entrée pour continuer...")
					return
				}
				common.LogInfo("Répertoire %s créé avec succès.", path)
				fmt.Printf("%sRépertoire créé avec succès.%s\n", colorGreen, colorReset)
			} else {
				common.LogInfo("Création de répertoire annulée par l'utilisateur pour %s.", path)
				fmt.Printf("%sOpération annulée.%s\n", colorYellow, colorReset)
				readInput("Appuyez sur Entrée pour continuer...")
				return
			}
		}
	} else if destType == "rsync" {
		common.LogInfo("Configuration de la destination rsync.")
		// Pour une destination rsync, sélectionner un serveur et un module
		servers, err := common.GetRsyncServers()
		if err != nil || len(servers) == 0 {
			common.LogError("Aucun serveur rsync configuré. Impossible d'ajouter une destination rsync: %v", err)
			fmt.Printf("%sAucun serveur rsync configuré. Veuillez d'abord ajouter un serveur.%s\n", colorRed, colorReset)
			readInput("Appuyez sur Entrée pour continuer...")
			return
		}
		
		// Afficher la liste des serveurs
		fmt.Println("\nServeurs rsync disponibles:")
		for i, server := range servers {
			fmt.Printf("  %d. %s (%s@%s)\n", i+1, server.Name, server.Username, server.IP)
		}
		
		// Sélectionner un serveur
		serverChoice := readInput("\nSélectionnez un serveur [1]: ")
		serverIndex := 0
		if serverChoice != "" {
			if idx, err := strconv.Atoi(serverChoice); err == nil && idx >= 1 && idx <= len(servers) {
				serverIndex = idx - 1
			} else {
				common.LogWarning("Choix de serveur rsync non valide: %s. Utilisation du premier serveur.\n", serverChoice)
				fmt.Printf("%sChoix non valide. Utilisation du premier serveur.%s\n", colorYellow, colorReset)
			}
		}
		
		selectedServer := servers[serverIndex]
		rsyncServer = &selectedServer
		common.LogInfo("Serveur rsync sélectionné pour la destination: %s.", selectedServer.Name)
		
		// Sélectionner un module si disponible
		if len(selectedServer.Modules) > 0 {
			fmt.Println("\nModules disponibles:")
			for i, module := range selectedServer.Modules {
				fmt.Printf("  %d. %s\n", i+1, module)
			}
			
			moduleChoice := readInput(fmt.Sprintf("\nSélectionnez un module [%s]: ", selectedServer.DefaultModule))
			selectedModule := selectedServer.DefaultModule
			
			if moduleChoice != "" {
				if idx, err := strconv.Atoi(moduleChoice); err == nil && idx >= 1 && idx <= len(selectedServer.Modules) {
					selectedModule = selectedServer.Modules[idx-1]
					common.LogInfo("Module rsync sélectionné: %s.", selectedModule)
				} else {
					common.LogWarning("Choix de module rsync invalide: %s. Utilisation du module par défaut.\n", moduleChoice)
					fmt.Printf("%sChoix non valide. Utilisation du module par défaut.%s\n", colorYellow, colorReset)
				}
			}
			
			// Construire le chemin rsync
			path = fmt.Sprintf("rsync://%s@%s/%s", selectedServer.Username, selectedServer.IP, selectedModule)
		} else {
			// Si aucun module, demander un chemin distant
			path = readAndValidateInput("Chemin distant sur le serveur: ", common.IsValidPath, "Chemin invalide.") // Utilisation de common.IsValidPath
			common.LogInfo("Chemin distant sur le serveur rsync: %s.", path)
		}
	}
	
	// Demander si cette destination doit être la destination par défaut
	defaultDestStr := readInput("\nDéfinir comme destination par défaut? (o/n): ")
	isDefaultDest := strings.ToLower(defaultDestStr) == "o"
	
	// Créer la nouvelle destination
	newDest := common.BackupDestination{
		Name:        name,
		Path:        path,
		Type:        destType,
		IsDefault:   isDefaultDest,
		RsyncServer: rsyncServer,
	}
	
	// Ajouter la destination à la configuration
	if err := common.AddBackupDestination(newDest); err != nil {
		common.LogError("Erreur lors de l'ajout de la destination de sauvegarde %s: %v", newDest.Name, err)
		fmt.Printf("%sErreur lors de l'ajout de la destination: %v%s\n", colorRed, err, colorReset)
	} else {
		common.LogInfo("Destination de sauvegarde %s ajoutée avec succès.", newDest.Name)
		fmt.Printf("%sDestination '%s' ajoutée avec succès.%s\n", colorGreen, name, colorReset)
	}
	
	readInput("Appuyez sur Entrée pour continuer...")
}

// editBackupDestination permet de modifier une destination de sauvegarde existante
func editBackupDestination() {
	common.LogInfo("Début de la modification d'une destination de sauvegarde.")
	if len(common.AppConfig.BackupDestinations) == 0 {
		common.LogWarning("Aucune destination de sauvegarde à modifier.")
		fmt.Printf("%sAucune destination à modifier.%s\n", colorYellow, colorReset)
		return
	}
	
	idxStr := readInput("Numéro de la destination à modifier: ")
	idx, err := strconv.Atoi(idxStr)
	
	if err != nil || idx < 1 || idx > len(common.AppConfig.BackupDestinations) {
		common.LogError("Numéro de destination invalide pour modification: %s", idxStr)
		fmt.Printf("%sNuméro invalide.%s\n", colorRed, colorReset)
		return
	}
	
	clearScreen()
	fmt.Printf("%sModification d'une Destination de Sauvegarde%s\n\n", colorBold, colorReset)
	
	// Afficher la liste des destinations
	fmt.Println("Destinations disponibles:")
	for i, dest := range common.AppConfig.BackupDestinations {
		defaultMark := ""
		if dest.IsDefault {
			defaultMark = fmt.Sprintf(" %s(Par défaut)%s", colorGreen, colorReset)
		}
		fmt.Printf("  %d. %s (%s)%s\n", i+1, dest.Name, dest.Type, defaultMark)
	}
	
	// Sélectionner une destination
	destChoice := readInput("\nSélectionnez une destination à modifier [1]: ")
	destIndex := 0
	if destChoice != "" {
		if idx, err := strconv.Atoi(destChoice); err == nil && idx >= 1 && idx <= len(common.AppConfig.BackupDestinations) {
			destIndex = idx - 1
		} else {
			fmt.Printf("%sChoix non valide. Modification annulée.%s\n", colorRed, colorReset)
			return
		}
	}
	
	// Récupérer la destination à modifier
	destination := common.AppConfig.BackupDestinations[destIndex]
	
	// Modifier les propriétés
	fmt.Printf("\nModification de la destination: %s\n\n", destination.Name)
	
	// Nom (ne peut pas être vide)
	newName := readAndValidateInput(fmt.Sprintf("Nouveau nom [%s]: ", destination.Name), common.IsValidName, "Nom invalide.") // Utilisation de common.IsValidName
	if newName != "" {
		// Vérifier si le nouveau nom existe déjà pour une autre destination
		if newName != destination.Name {
			for _, dest := range common.AppConfig.BackupDestinations {
				if dest.Name == newName {
					fmt.Printf("%sUne destination avec ce nom existe déjà.%s\n", colorRed, colorReset)
					return
				}
			}
		}
		destination.Name = newName
	}
	
	// Pour les destinations locales, permettre de modifier le chemin
	if destination.Type == "local" {
		newPath := readAndValidateInput(fmt.Sprintf("Nouveau chemin [%s]: ", destination.Path), common.IsValidPath, "Chemin invalide.") // Utilisation de common.IsValidPath
		if newPath != "" {
			destination.Path = newPath
			
			// Vérifier si le répertoire existe, sinon proposer de le créer
			if !common.DirExists(newPath) {
				createDir := readInput(fmt.Sprintf("Le répertoire %s n'existe pas. Voulez-vous le créer? (o/n): ", newPath))
				if strings.ToLower(createDir) == "o" || strings.ToLower(createDir) == "oui" {
					if err := os.MkdirAll(newPath, 0755); err != nil {
						fmt.Printf("%sErreur lors de la création du répertoire: %v%s\n", colorRed, err, colorReset)
					} else {
						fmt.Printf("%sRépertoire créé avec succès.%s\n", colorGreen, colorReset)
					}
				}
			}
		}
	} else if destination.Type == "rsync" {
		// Pour rsync, permettre de changer le serveur ou le module
		fmt.Println("\nVoulez-vous modifier le serveur rsync?")
		changeServer := readInput("Changer le serveur? (o/n): ")
		
		if changeServer == "o" || changeServer == "O" || changeServer == "oui" || changeServer == "Oui" {
			servers, err := common.GetRsyncServers()
			if err != nil || len(servers) == 0 {
				fmt.Printf("%sAucun serveur rsync configuré.%s\n", colorRed, colorReset)
			} else {
				// Afficher la liste des serveurs
				fmt.Println("\nServeurs rsync disponibles:")
				for i, server := range servers {
					fmt.Printf("  %d. %s (%s@%s)\n", i+1, server.Name, server.Username, server.IP)
				}
				
				// Sélectionner un serveur
				serverChoice := readInput("\nSélectionnez un serveur [1]: ")
				serverIndex := 0
				if serverChoice != "" {
					if idx, err := strconv.Atoi(serverChoice); err == nil && idx >= 1 && idx <= len(servers) {
						serverIndex = idx - 1
					} else {
						fmt.Printf("%sChoix non valide.%s\n", colorYellow, colorReset)
						serverIndex = -1
					}
				}
				
				if serverIndex >= 0 {
					selectedServer := servers[serverIndex]
					destination.RsyncServer = &selectedServer
					
					// Sélectionner un module si disponible
					if len(selectedServer.Modules) > 0 {
						fmt.Println("\nModules disponibles:")
						for i, module := range selectedServer.Modules {
							fmt.Printf("  %d. %s\n", i+1, module)
						}
						
						moduleChoice := readInput(fmt.Sprintf("\nSélectionnez un module [%s]: ", selectedServer.DefaultModule))
						selectedModule := selectedServer.DefaultModule
						
						if moduleChoice != "" {
							if idx, err := strconv.Atoi(moduleChoice); err == nil && idx >= 1 && idx <= len(selectedServer.Modules) {
								selectedModule = selectedServer.Modules[idx-1]
							} else {
								fmt.Printf("%sChoix non valide. Utilisation du module par défaut.%s\n", colorYellow, colorReset)
							}
						}
						
						// Construire le chemin rsync
						destination.Path = fmt.Sprintf("rsync://%s@%s/%s", selectedServer.Username, selectedServer.IP, selectedModule)
					} else {
						fmt.Printf("%sAucun module disponible pour ce serveur.%s\n", colorRed, colorReset)
					}
				}
			}
		}
		
		// Définir comme destination par défaut
		if !destination.IsDefault {
			defaultChoice := readInput("\nDéfinir comme destination par défaut? (o/n): ")
			if defaultChoice == "o" || defaultChoice == "O" || defaultChoice == "oui" || defaultChoice == "Oui" {
				destination.IsDefault = true
				
				// Mettre à jour les autres destinations
				for i := range common.AppConfig.BackupDestinations {
					if i != destIndex {
						common.AppConfig.BackupDestinations[i].IsDefault = false
					}
				}
				
				// Mettre à jour aussi le champ BackupDestination pour compatibilité
				common.AppConfig.BackupDestination = destination.Path
			}
		}
		
		// Mettre à jour la destination
		common.AppConfig.BackupDestinations[destIndex] = destination
		
		// Sauvegarder la configuration
		if err := common.SaveConfig(common.AppConfig); err != nil {
			fmt.Printf("%sErreur lors de la sauvegarde de la configuration: %v%s\n", colorRed, err, colorReset)
		} else {
			fmt.Printf("%sDestination modifiée avec succès.%s\n", colorGreen, colorReset)
		}
		
		readInput("\nAppuyez sur Entrée pour continuer...")
	}
}

// deleteBackupDestination permet de supprimer une destination de sauvegarde
func deleteBackupDestination() {
	if len(common.AppConfig.BackupDestinations) == 0 {
		fmt.Printf("%sAucune destination à supprimer.%s\n", colorYellow, colorReset)
		return
	}
	
	idxStr := readInput("Numéro de la destination à supprimer: ")
	idx, err := strconv.Atoi(idxStr)
	
	if err != nil || idx < 1 || idx > len(common.AppConfig.BackupDestinations) {
		fmt.Printf("%sNuméro invalide.%s\n", colorRed, colorReset)
		return
	}
	
	clearScreen()
	fmt.Printf("%sSuppression d'une Destination de Sauvegarde%s\n\n", colorBold, colorReset)
	
	// Afficher la liste des destinations
	fmt.Println("Destinations disponibles:")
	for i, dest := range common.AppConfig.BackupDestinations {
		defaultMark := ""
		if dest.IsDefault {
			defaultMark = fmt.Sprintf(" %s(Par défaut)%s", colorGreen, colorReset)
		}
		fmt.Printf("  %d. %s (%s)%s\n", i+1, dest.Name, dest.Type, defaultMark)
	}
	
	// Sélectionner une destination
	destChoice := readInput("\nSélectionnez une destination à supprimer [1]: ")
	destIndex := 0
	if destChoice != "" {
		if idx, err := strconv.Atoi(destChoice); err == nil && idx >= 1 && idx <= len(common.AppConfig.BackupDestinations) {
			destIndex = idx - 1
		} else {
			fmt.Printf("%sChoix non valide. Suppression annulée.%s\n", colorRed, colorReset)
			return
		}
	}
	
	// Récupérer la destination à supprimer
	destination := common.AppConfig.BackupDestinations[destIndex]
	
	// Confirmation avant suppression
	confirm := readInput(fmt.Sprintf("\nÊtes-vous sûr de vouloir supprimer la destination '%s'? (o/n): ", destination.Name))
	if confirm != "o" && confirm != "O" && confirm != "oui" && confirm != "Oui" {
		fmt.Printf("%sSuppression annulée.%s\n", colorYellow, colorReset)
		return
	}
	
	// Vérifier si c'est la destination par défaut et s'il y a d'autres destinations
	if destination.IsDefault && len(common.AppConfig.BackupDestinations) > 1 {
		fmt.Println("\nCette destination est définie comme destination par défaut.")
		fmt.Println("Une autre destination sera automatiquement définie comme destination par défaut.")
	}
	
	// Supprimer la destination
	if err := common.DeleteBackupDestination(destination.Name); err != nil {
		fmt.Printf("%sErreur lors de la suppression de la destination: %v%s\n", colorRed, err, colorReset)
	} else {
		fmt.Printf("%sDestination supprimée avec succès.%s\n", colorGreen, colorReset)
	}
	
	return
}

// setDefaultDestination permet de définir une destination par défaut
func setDefaultDestination() {
	common.LogInfo("Début de la définition de la destination par défaut.")
	if len(common.AppConfig.BackupDestinations) == 0 {
		common.LogWarning("Aucune destination de sauvegarde à définir par défaut.")
		fmt.Printf("%sAucune destination à définir par défaut.%s\n", colorYellow, colorReset)
		return
	}
	
	fmt.Println("Destinations disponibles:")
	for i, dest := range common.AppConfig.BackupDestinations {
		defaultMark := ""
		if dest.IsDefault {
			defaultMark = " (Par défaut)"
		}
		fmt.Printf("%d. %s%s\n", i+1, dest.Name, defaultMark)
	}
	
	idxStr := readInput("Sélectionnez la destination par défaut (numéro): ")
	idx, err := strconv.Atoi(idxStr)
	
	if err != nil || idx < 1 || idx > len(common.AppConfig.BackupDestinations) {
		common.LogError("Numéro de destination invalide pour la définition par défaut: %s", idxStr)
		fmt.Printf("%sNuméro invalide.%s\n", colorRed, colorReset)
		return
	}
	
	// Désactiver le flag par défaut pour toutes les destinations
	for i := range common.AppConfig.BackupDestinations {
		common.AppConfig.BackupDestinations[i].IsDefault = false
	}
	
	// Définir la nouvelle destination par défaut
	common.AppConfig.BackupDestinations[idx-1].IsDefault = true
	common.AppConfig.BackupDestination = common.AppConfig.BackupDestinations[idx-1].Path
	
	if err := common.SaveConfig(common.AppConfig); err != nil {
		common.LogError("Erreur lors de la mise à jour de la destination par défaut: %v", err)
		fmt.Printf("%sErreur lors de la mise à jour de la configuration: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	common.LogInfo("Destination par défaut mise à jour avec succès.")
	fmt.Printf("%sDestination par défaut mise à jour avec succès.%s\n", colorGreen, colorReset)
}