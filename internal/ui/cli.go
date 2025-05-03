package ui

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Noziop/s4v3my4ss/internal/restore"
	"github.com/Noziop/s4v3my4ss/internal/watch"
	"github.com/Noziop/s4v3my4ss/internal/wrappers"
	"github.com/Noziop/s4v3my4ss/pkg/common"
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
		"node_modules",
		".git",
		".svn",
		".hg",
		"__pycache__",
		".venv",
		"venv",
		"env",
		"dist",
		"build",
		"target",
		"bin",
		"obj",
		".next",
		".nuxt",
		".output",
		"vendor",
		".vscode",
		".idea",
		".DS_Store",
		"coverage",
		".pytest_cache",
		".gradle",
		"tmp",
		"temp",
		"logs",
		"cache",
	}

	// Patterns de fichiers à exclure par défaut
	commonExcludeFiles = []string{
		"*.log",
		"*.tmp",
		"*.swp",
		"*.swo",
		"*.pyc",
		"*.pyo",
		"*.class",
		"*.o",
		"*.so",
		"*.exe",
		"*.dll",
		"*.db",
		"*.sqlite",
		"*.sqlite3",
		"*.pid",
		"package-lock.json",
		"yarn.lock",
		"pnpm-lock.yaml",
		".env",
		".cache",
		".env.*",
	}
)

// Constante pour le nom de la commande
const CommandName = "saveme"

var reader *bufio.Reader

// RunInteractiveMode démarre le mode interactif de l'application
func RunInteractiveMode() {
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
		case "0":
			fmt.Println("Au revoir !")
			return
		default:
			fmt.Println("Option non valide. Veuillez réessayer.")
		}
		
		fmt.Println()
		readInput("Appuyez sur Entrée pour continuer...")
	}
}

// HandleWatchCommand traite la commande 'watch' depuis la ligne de commande
func HandleWatchCommand(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: "+CommandName+" watch <nom_configuration>")
		os.Exit(1)
	}
	
	name := args[0]
	config, found := common.GetBackupConfig(name)
	
	if !found {
		fmt.Fprintf(os.Stderr, "Erreur: Configuration '%s' non trouvée\n", name)
		os.Exit(1)
	}
	
	fmt.Printf("Démarrage de la surveillance du répertoire: %s\n", config.SourcePath)
	if err := watch.StartWatch(config); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur de surveillance: %v\n", err)
		os.Exit(1)
	}
}

// HandleRestoreCommand traite la commande 'restore' depuis la ligne de commande
func HandleRestoreCommand(args []string) {
	if len(args) < 1 {
		restoreBackupInteractive()
		return
	}
	
	backupID := args[0]
	target := ""
	if len(args) > 1 {
		target = args[1]
	}
	
	if err := restore.RestoreBackup(backupID, target); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur de restauration: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Restauration terminée avec succès.")
}

// HandleManageCommand traite la commande 'manage' depuis la ligne de commande
func HandleManageCommand(args []string) {
	if len(args) == 0 {
		manageBackupsInteractive()
		return
	}
	
	subcommand := args[0]
	
	switch subcommand {
	case "list":
		listBackups()
	case "delete":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: "+CommandName+" manage delete <backup_id>")
			os.Exit(1)
		}
		deleteBackup(args[1])
	case "clean":
		cleanOldBackups()
	default:
		fmt.Fprintln(os.Stderr, "Sous-commande inconnue:", subcommand)
		fmt.Fprintln(os.Stderr, "Sous-commandes disponibles: list, delete, clean")
		os.Exit(1)
	}
}

// configureBackup permet de configurer une nouvelle sauvegarde
func configureBackup() {
	fmt.Printf("%sConfiguration d'une nouvelle sauvegarde%s\n\n", colorBold, colorReset)
	
	name := readInput("Nom de la sauvegarde: ")
	sourcePath := readInput("Chemin du répertoire à surveiller: ")
	
	// Expandir les chemins relatifs, y compris ~/
	if strings.HasPrefix(sourcePath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			sourcePath = filepath.Join(homeDir, sourcePath[2:])
		}
	}
	
	// Normaliser et vérifier le chemin
	sourcePath = filepath.Clean(sourcePath)
	if !common.DirExists(sourcePath) {
		fmt.Printf("%sErreur: Le répertoire '%s' n'existe pas.%s\n", colorRed, sourcePath, colorReset)
		return
	}
	
	// Permettre à l'utilisateur de choisir l'emplacement de sauvegarde
	backupDestination := readInput("Emplacement des sauvegardes (vide pour utiliser l'emplacement par défaut): ")
	if backupDestination != "" {
		// Expandir les chemins relatifs, y compris ~/
		if strings.HasPrefix(backupDestination, "~/") {
			homeDir, err := os.UserHomeDir()
			if err == nil {
				backupDestination = filepath.Join(homeDir, backupDestination[2:])
			}
		}
		
		// Normaliser le chemin
		backupDestination = filepath.Clean(backupDestination)
		
		// Vérifier/créer le répertoire de destination
		if !common.DirExists(backupDestination) {
			fmt.Printf("Le répertoire de destination n'existe pas. Voulez-vous le créer? (o/n): ")
			createDest := readInput("")
			if strings.ToLower(createDest) == "o" {
				if err := os.MkdirAll(backupDestination, 0755); err != nil {
					fmt.Printf("%sErreur lors de la création du répertoire: %v%s\n", colorRed, err, colorReset)
					return
				}
			} else {
				fmt.Println("Configuration annulée.")
				return
			}
		}
		
		// Mettre à jour la destination dans la configuration globale
		common.AppConfig.BackupDestination = backupDestination
		if err := common.SaveConfig(common.AppConfig); err != nil {
			fmt.Printf("%sErreur lors de la mise à jour de la configuration: %v%s\n", colorRed, err, colorReset)
			return
		}
	}
	
	// Compression
	compressStr := readInput("Activer la compression? (o/n): ")
	compression := strings.ToLower(compressStr) == "o"
	
	// Répertoires à exclure
	excludeDirsStr := readInput("Répertoires à exclure (séparés par des virgules): ")
	var excludeDirs []string
	if excludeDirsStr != "" {
		excludeDirs = strings.Split(excludeDirsStr, ",")
		for i := range excludeDirs {
			excludeDirs[i] = strings.TrimSpace(excludeDirs[i])
		}
	} else {
		excludeDirs = commonExcludeDirs
	}
	
	// Fichiers à exclure
	excludeFilesStr := readInput("Fichiers à exclure (séparés par des virgules): ")
	var excludeFiles []string
	if excludeFilesStr != "" {
		excludeFiles = strings.Split(excludeFilesStr, ",")
		for i := range excludeFiles {
			excludeFiles[i] = strings.TrimSpace(excludeFiles[i])
		}
	} else {
		excludeFiles = commonExcludeFiles
	}
	
	// Intervalle (0 = pas de surveillance automatique)
	intervalStr := readInput("Intervalle de sauvegarde en minutes (0 pour désactiver): ")
	interval := 0
	if i, err := strconv.Atoi(intervalStr); err == nil && i >= 0 {
		interval = i
	}
	
	// Créer la configuration
	config := common.BackupConfig{
		Name:         name,
		SourcePath:   sourcePath,
		Compression:  compression,
		ExcludeDirs:  excludeDirs,
		ExcludeFiles: excludeFiles,
		Interval:     interval,
	}
	
	// Ajouter à la configuration
	if err := common.AddBackupDirectory(config); err != nil {
		fmt.Printf("%sErreur lors de l'ajout de la configuration: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	fmt.Printf("%sConfiguration ajoutée avec succès.%s\n", colorGreen, colorReset)
	
	// Proposer de démarrer la surveillance
	startWatchStr := readInput("Démarrer la surveillance maintenant? (o/n): ")
	if strings.ToLower(startWatchStr) == "o" {
		fmt.Printf("Démarrage de la surveillance du répertoire: %s\n", config.SourcePath)
		fmt.Println("Mode surveillance continue. La surveillance s'exécute en arrière-plan.")
		fmt.Println("Les sauvegardes continueront même si vous revenez au menu principal.")
		
		// Lancer la surveillance en arrière-plan pour ne pas bloquer l'interface
		go func() {
			if err := watch.StartWatch(config); err != nil {
				fmt.Printf("%sErreur lors de la surveillance: %v%s\n", colorRed, err, colorReset)
			}
		}()
	}
}

// watchDirectoryInteractive permet de démarrer la surveillance d'un répertoire
func watchDirectoryInteractive() {
	fmt.Printf("%sSurveillance d'un répertoire%s\n\n", colorBold, colorReset)
	
	// Afficher les configurations disponibles
	configs := common.AppConfig.BackupDirs
	if len(configs) == 0 {
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
		fmt.Printf("%sChoix invalide.%s\n", colorRed, colorReset)
		return
	}
	
	config := configs[idx-1]
	fmt.Printf("Démarrage de la surveillance du répertoire: %s\n", config.SourcePath)
	
	// Option pour la durée de surveillance
	durationStr := readInput("Durée de surveillance en minutes (0 pour mode continu, Ctrl+C pour arrêter): ")
	duration := 0
	if dur, err := strconv.Atoi(durationStr); err == nil && dur > 0 {
		duration = dur
	}
	
	if duration > 0 {
		// Surveillance avec timer
		fmt.Printf("Surveillance pendant %d minutes...\n", duration)
		
		// Créer un canal pour communiquer avec la goroutine de surveillance
		done := make(chan bool)
		
		// Lancer la surveillance dans une goroutine
		go func() {
			if err := watch.StartWatchWithCallback(config, done); err != nil {
				fmt.Printf("%sErreur lors de la surveillance: %v%s\n", colorRed, err, colorReset)
			}
		}()
		
		// Attendre la durée spécifiée
		time.Sleep(time.Duration(duration) * time.Minute)
		
		// Signaler l'arrêt
		done <- true
		
		fmt.Printf("\n%sSurveillance terminée après %d minutes.%s\n", colorGreen, duration, colorReset)
	} else {
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
	fmt.Printf("%sRestauration d'une sauvegarde%s\n\n", colorBold, colorReset)
	
	backups, err := common.ListBackups()
	if err != nil {
		fmt.Printf("%sErreur lors de la récupération des sauvegardes: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	if len(backups) == 0 {
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
		fmt.Printf("%sChoix invalide.%s\n", colorRed, colorReset)
		return
	}
	
	backup := backups[idx-1]
	
	targetPath := readInput("Chemin de destination (vide pour restaurer à l'emplacement d'origine): ")
	if targetPath == "" {
		targetPath = backup.SourcePath
	}
	
	// Vérifier la destination
	if common.DirExists(targetPath) {
		overwriteStr := readInput(fmt.Sprintf("Le répertoire '%s' existe déjà. Écraser? (o/n): ", targetPath))
		if strings.ToLower(overwriteStr) != "o" {
			fmt.Println("Restauration annulée.")
			return
		}
	}
	
	fmt.Printf("Restauration de la sauvegarde '%s' vers '%s'...\n", backup.Name, targetPath)
	
	if err := restore.RestoreBackup(backup.ID, targetPath); err != nil {
		fmt.Printf("%sErreur lors de la restauration: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	fmt.Printf("%sRestauration terminée avec succès.%s\n", colorGreen, colorReset)
}

// manageBackupsInteractive permet de gérer les sauvegardes
func manageBackupsInteractive() {
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
			return
		default:
			fmt.Println("Option non valide. Veuillez réessayer.")
		}
		
		readInput("Appuyez sur Entrée pour continuer...")
	}
}

// listBackups affiche la liste des sauvegardes
func listBackups() {
	backups, err := common.ListBackups()
	if err != nil {
		fmt.Printf("%sErreur lors de la récupération des sauvegardes: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	if len(backups) == 0 {
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
}

// deleteBackupInteractive permet de supprimer une sauvegarde
func deleteBackupInteractive() {
	backups, err := common.ListBackups()
	if err != nil {
		fmt.Printf("%sErreur lors de la récupération des sauvegardes: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	if len(backups) == 0 {
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
		fmt.Printf("%sChoix invalide.%s\n", colorRed, colorReset)
		return
	}
	
	backup := backups[idx-1]
	
	confirmStr := readInput(fmt.Sprintf("Êtes-vous sûr de vouloir supprimer '%s'? (o/n): ", backup.Name))
	if strings.ToLower(confirmStr) != "o" {
		fmt.Println("Suppression annulée.")
		return
	}
	
	deleteBackup(backup.ID)
}

// deleteBackup supprime une sauvegarde
func deleteBackup(id string) {
	fmt.Printf("Suppression de la sauvegarde %s...\n", id)
	
	err := common.DeleteBackup(id)
	if err != nil {
		fmt.Printf("%sErreur lors de la suppression: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	fmt.Printf("%sSuppression terminée avec succès.%s\n", colorGreen, colorReset)
}

// cleanOldBackups nettoie les anciennes sauvegardes selon la politique de rétention
func cleanOldBackups() {
	fmt.Println("Nettoyage des anciennes sauvegardes...")
	// Appel au module de gestion pour nettoyer les anciennes sauvegardes
	// À implémenter plus tard
	fmt.Printf("%sNettoyage terminé.%s\n", colorGreen, colorReset)
}

// checkDependencies vérifie et installe les dépendances nécessaires
func checkDependencies() {
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
			fmt.Printf("%sOK%s\n", colorGreen, colorReset)
			continue
		}
		
		fmt.Printf("%sNon trouvé%s\n", colorYellow, colorReset)
		installStr := readInput(fmt.Sprintf("Installer %s (%s)? (o/n): ", dep.pkg, dep.desc))
		
		if strings.ToLower(installStr) != "o" {
			fmt.Println("Installation ignorée.")
			continue
		}
		
		fmt.Printf("Installation de %s...\n", dep.pkg)
		if err := common.EnsureDependency(dep.command, dep.pkg); err != nil {
			fmt.Printf("%sErreur lors de l'installation: %v%s\n", colorRed, err, colorReset)
		} else {
			fmt.Printf("%s%s installé avec succès.%s\n", colorGreen, dep.pkg, colorReset)
		}
	}
}

// Fonctions utilitaires
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// displayHeader affiche l'en-tête de l'application
func displayHeader() {
	// Essayer de lire le fichier banner.txt
	execPath, err := os.Executable()
	var bannerPath string
	if err == nil {
		// Chercher le banner par rapport à l'emplacement de l'exécutable
		bannerPath = filepath.Join(filepath.Dir(execPath), "..", "banner.txt")
	}
	
	// Si on ne trouve pas le banner à partir de l'exécutable, essayer dans le répertoire actuel
	if _, err := os.Stat(bannerPath); os.IsNotExist(err) {
		// Essayer dans le répertoire projet
		bannerPath = filepath.Join(".", "banner.txt")
		// Si toujours pas trouvé, utiliser un chemin absolu pour le développement
		if _, err := os.Stat(bannerPath); os.IsNotExist(err) {
			bannerPath = "/home/noziop/projects/s4v3my4ss/Projet Go/banner.txt"
		}
	}

	// Lire le fichier banner s'il existe
	if bannerContent, err := ioutil.ReadFile(bannerPath); err == nil {
		fmt.Printf("%s%s", colorBlue, colorBold)
		fmt.Println(string(bannerContent))
		fmt.Printf("%s\n", colorReset)
	} else {
		// Fallback sur l'ASCII art codé en dur en cas d'échec
		fmt.Printf("%s%s", colorBlue, colorBold)
		fmt.Println("  ___ _ _  _               __  __  ___ ")
		fmt.Println(" / __| | \\| |  ___  /\\ /\\ /__\\/__\\|_  )")
		fmt.Println(" \\__ \\ | .` | / -_)/ _  //_\\ / \\/ / / / ")
		fmt.Println(" |___/_|_|\\_| \\___|\\__,_/\\__/\\__//___| ")
		fmt.Printf("%s\n", colorReset)
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
		fmt.Println("0. Quitter")
		
		fmt.Print("\nVotre choix: ")
		choice := readInput("")
		
		switch choice {
		case "1":
			configureBackup()
		case "2":
			// manualBackup - à implémenter plus tard
			fmt.Println("Fonction non implémentée")
		case "3":
			restoreBackupInteractive()
		case "4":
			watchDirectoryInteractive()
		case "5":
			manageBackupsInteractive()
		case "6":
			discoverRsyncServers()
		case "0":
			fmt.Println("Au revoir!")
			return
		default:
			fmt.Printf("\n%sOption invalide. Appuyez sur Entrée pour continuer...%s", colorRed, colorReset)
			readInput("")
		}
	}
}

// discoverRsyncServers recherche et configure les serveurs rsync sur le réseau
func discoverRsyncServers() {
	clearScreen()
	fmt.Printf("%sRecherche de serveurs rsync sur le réseau%s\n\n", colorBold, colorReset)
	
	// Demander le sous-réseau à scanner
	fmt.Println("Entrez le sous-réseau à scanner (format CIDR, par exemple 192.168.0.0/24)")
	fmt.Println("ou laissez vide pour utiliser 192.168.0.0/24:")
	
	subnetCIDR := readInput("Sous-réseau: ")
	if subnetCIDR == "" {
		subnetCIDR = "192.168.0.0/24"
	}
	
	// Afficher un message d'attente
	fmt.Printf("\nRecherche de serveurs rsync sur %s...\n", subnetCIDR)
	fmt.Println("Cela peut prendre jusqu'à 30 secondes. Veuillez patienter...")
	
	// Créer l'objet de découverte
	discovery := wrappers.NewRsyncDiscovery()
	
	// Scanner le réseau (timeout de 30 secondes)
	servers := discovery.ScanNetwork(subnetCIDR, 30)
	
	if len(servers) == 0 {
		fmt.Printf("\n%sAucun serveur rsync trouvé sur le réseau.%s\n", colorYellow, colorReset)
		fmt.Println("Vérifiez que:")
		fmt.Println("1. Le service rsync est activé sur votre NAS ou serveur")
		fmt.Println("2. Le port rsync (873) est ouvert dans le pare-feu")
		fmt.Println("3. Le sous-réseau spécifié est correct")
		return
	}
	
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
			fmt.Printf("\n%sChoix invalide.%s\n", colorRed, colorReset)
		}
		return
	}
	
	// Configurer le serveur sélectionné
	selectedServer := servers[choice-1]
	configureRsyncServer(selectedServer)
}

// configureRsyncServer configure un serveur rsync sélectionné
func configureRsyncServer(server wrappers.RsyncServer) {
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
	
	name := readInput(fmt.Sprintf("Nom du serveur (défaut: %s): ", defaultName))
	if name == "" {
		name = defaultName
	}
	
	// Demander le nom d'utilisateur
	username := readInput("Nom d'utilisateur pour la connexion: ")
	
	// Si des modules sont disponibles, demander lequel utiliser
	defaultModule := ""
	if len(server.Modules) > 0 {
		fmt.Println("\nModules disponibles:")
		for i, module := range server.Modules {
			fmt.Printf("%d. %s: %s\n", i+1, module.Name, module.Description)
		}
		
		moduleChoice := readInput("\nChoisissez un module par défaut (0 pour aucun): ")
		
		if moduleIdx, err := strconv.Atoi(moduleChoice); err == nil && moduleIdx > 0 && moduleIdx <= len(server.Modules) {
			defaultModule = server.Modules[moduleIdx-1].Name
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
		fmt.Printf("\n%sErreur lors de la sauvegarde de la configuration: %v%s\n", colorRed, err, colorReset)
	} else {
		fmt.Printf("\n%sServeur rsync %s configuré avec succès!%s\n", colorGreen, name, colorReset)
	}
	
	// Proposer de configurer une sauvegarde vers ce serveur
	configBackup := readInput("\nVoulez-vous configurer une sauvegarde vers ce serveur maintenant? (o/n): ")
	
	if strings.ToLower(configBackup) == "o" {
		configureRemoteBackup(serverConfig)
	}
}

// configureRemoteBackup configure une sauvegarde vers un serveur rsync distant
func configureRemoteBackup(serverConfig common.RsyncServerConfig) {
	clearScreen()
	fmt.Printf("%sConfiguration d'une sauvegarde vers %s%s\n\n", colorBold, serverConfig.Name, colorReset)
	
	name := readInput("Nom de la sauvegarde: ")
	sourcePath := readInput("Chemin du répertoire à surveiller: ")
	
	// Expandir les chemins relatifs, y compris ~/
	if strings.HasPrefix(sourcePath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			sourcePath = filepath.Join(homeDir, sourcePath[2:])
		}
	}
	
	// Normaliser et vérifier le chemin
	sourcePath = filepath.Clean(sourcePath)
	if !common.DirExists(sourcePath) {
		fmt.Printf("%sErreur: Le répertoire '%s' n'existe pas.%s\n", colorRed, sourcePath, colorReset)
		return
	}
	
	// Sélectionner le module à utiliser
	module := serverConfig.DefaultModule
	if len(serverConfig.Modules) > 0 {
		fmt.Println("\nModules disponibles:")
		for i, mod := range serverConfig.Modules {
			fmt.Printf("%d. %s\n", i+1, mod)
		}
		
		moduleChoice := readInput("\nChoisissez un module (0 pour utiliser le paramètre par défaut): ")
		
		if moduleIdx, err := strconv.Atoi(moduleChoice); err == nil && moduleIdx > 0 && moduleIdx <= len(serverConfig.Modules) {
			module = serverConfig.Modules[moduleIdx-1]
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
	
	// Compression
	compressStr := readInput("Activer la compression? (o/n): ")
	compression := strings.ToLower(compressStr) == "o"
	
	// Exclusion de fichiers/dossiers
	excludeDirsStr := readInput("\nRépertoires à exclure (séparés par des virgules): ")
	var excludeDirs []string
	if excludeDirsStr != "" {
		excludeDirs = strings.Split(excludeDirsStr, ",")
		for i := range excludeDirs {
			excludeDirs[i] = strings.TrimSpace(excludeDirs[i])
		}
	} else {
		excludeDirs = commonExcludeDirs
	}
	
	excludeFilesStr := readInput("Fichiers à exclure (séparés par des virgules): ")
	var excludeFiles []string
	if excludeFilesStr != "" {
		excludeFiles = strings.Split(excludeFilesStr, ",")
		for i := range excludeFiles {
			excludeFiles[i] = strings.TrimSpace(excludeFiles[i])
		}
	} else {
		excludeFiles = commonExcludeFiles
	}
	
	// Intervalle de sauvegarde automatique
	intervalStr := readInput("Intervalle de sauvegarde en minutes (0 pour désactiver): ")
	interval := 0
	if intervalVal, err := strconv.Atoi(intervalStr); err == nil && intervalVal > 0 {
		interval = intervalVal
	}
	
	// Créer la configuration de sauvegarde
	backupConfig := common.BackupConfig{
		Name:         name,
		SourcePath:   sourcePath,
		Compression:  compression,
		ExcludeDirs:  excludeDirs,
		ExcludeFiles: excludeFiles,
		Interval:     interval,
		RemoteServer: &serverConfig, // Utiliser l'adresse de serverConfig pour obtenir un pointeur
	}
	
	// Mise à jour de la destination dans la configuration globale
	// Sauvegarder la destination temporairement
	prevDestination := common.AppConfig.BackupDestination
	common.AppConfig.BackupDestination = destination
	
	// Enregistrer la configuration
	if err := common.AddBackupDirectory(backupConfig); err != nil {
		fmt.Printf("%sErreur lors de l'enregistrement de la configuration: %v%s\n", colorRed, err, colorReset)
		common.AppConfig.BackupDestination = prevDestination // Restaurer l'ancienne destination
	} else {
		if err := common.SaveConfig(common.AppConfig); err != nil {
			fmt.Printf("%sErreur lors de la mise à jour de la destination: %v%s\n", colorRed, err, colorReset)
		} else {
			fmt.Printf("\n%sConfiguration de sauvegarde vers %s enregistrée avec succès!%s\n", 
				colorGreen, serverConfig.Name, colorReset)
				
			// Proposer de faire une sauvegarde immédiate
			doBackupNow := readInput("\nVoulez-vous effectuer une sauvegarde immédiate? (o/n): ")
			if strings.ToLower(doBackupNow) == "o" {
				fmt.Printf("\nDémarrage de la sauvegarde vers %s...\n", serverConfig.Name)
				
				// Générer un ID unique pour la sauvegarde
				backupID := generateBackupID(name)
				
				// Utiliser la fonction RsyncBackup pour effectuer la sauvegarde
				if err := wrappers.RsyncBackup(sourcePath, destination, excludeDirs, excludeFiles, compression, &serverConfig); err != nil {
					fmt.Printf("%sErreur lors de la sauvegarde: %v%s\n", colorRed, err, colorReset)
					return
				}
				
				// Obtenir la taille du répertoire source comme approximation
				size, err := getDirSize(sourcePath)
				if err != nil {
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
					IsIncremental: false,
					Compression:   compression,
					RemoteServer: &serverConfig, // Utiliser l'adresse de serverConfig pour obtenir un pointeur
				}
				
				// Sauvegarder les métadonnées
				if err := common.SaveBackupInfo(backupInfo); err != nil {
					fmt.Printf("%sErreur lors de l'enregistrement des métadonnées: %v%s\n", colorRed, err, colorReset)
					return
				}
				
				fmt.Printf("%sSauvegarde vers %s terminée avec succès!%s\n", colorGreen, serverConfig.Name, colorReset)
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
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}