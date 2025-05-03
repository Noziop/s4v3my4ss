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
		case "7":
			manageConfiguration()
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
		IsIncremental: incremental, // Correction du nom du champ (de Incremental à IsIncremental)
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
	if bannerContent, err := os.ReadFile(bannerPath); err == nil {
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
		// Bannière de secours si le fichier n'est pas trouvé
		fmt.Print(colorBold)
		fmt.Println("  ___ _ _  _               __  __  ___ ")
		fmt.Println(" / __| | \\| |  ___  /\\ /\\ /__\\/__\\|_  )")
		fmt.Println(" \\__ \\ | .` | / -_)/ _  //_\\ / \\/ / / / ")
		fmt.Println(" |___/_|_|\\_| \\___|\\__,_/\\__/\\__//___| ")
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

// manageConfiguration permet de gérer la configuration de l'application
func manageConfiguration() {
	for {
		clearScreen()
		fmt.Printf("%sGestion de la Configuration%s\n\n", colorBold, colorReset)
		
		fmt.Printf("  %s1.%s Afficher la configuration complète\n", colorGreen, colorReset)
		fmt.Printf("  %s2.%s Modifier les répertoires sauvegardés\n", colorGreen, colorReset)
		fmt.Printf("  %s3.%s Modifier/supprimer les serveurs rsync\n", colorGreen, colorReset)
		fmt.Printf("  %s4.%s Modifier la politique de rétention\n", colorGreen, colorReset)
		fmt.Printf("  %s5.%s Modifier la destination des sauvegardes\n", colorGreen, colorReset)
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
			changeBackupDestination()
		case "0":
			return
		default:
			fmt.Println("Option non valide. Veuillez réessayer.")
		}
		
		readInput("Appuyez sur Entrée pour continuer...")
	}
}

// displayFullConfig affiche la configuration complète de l'application
func displayFullConfig() {
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
		fmt.Printf("%sAucun répertoire configuré.%s\n\n", colorYellow, colorReset)
	} else {
		for i, dir := range common.AppConfig.BackupDirs {
			fmt.Printf("%d. %s (%s)\n", i+1, dir.Name, dir.SourcePath)
			fmt.Printf("   Compression: %v, Incrémental: %v, Intervalle: %d minutes\n", 
				dir.Compression, dir.IsIncremental, dir.Interval)
			
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
		fmt.Printf("%sAucun serveur rsync configuré.%s\n", colorYellow, colorReset)
	} else {
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
}

// manageBackupDirectories permet de gérer les répertoires à sauvegarder
func manageBackupDirectories() {
	for {
		clearScreen()
		fmt.Printf("%sGestion des répertoires sauvegardés%s\n\n", colorBold, colorReset)
		
		// Afficher les répertoires sauvegardés
		if len(common.AppConfig.BackupDirs) == 0 {
			fmt.Printf("%sAucun répertoire configuré.%s\n\n", colorYellow, colorReset)
		} else {
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
			return
		case "2":
			editBackupDirectory()
		case "3":
			deleteBackupDirectory()
		case "0":
			return
		default:
			fmt.Println("Option non valide.")
		}
		
		readInput("Appuyez sur Entrée pour continuer...")
	}
}

// editBackupDirectory permet de modifier un répertoire de sauvegarde existant
func editBackupDirectory() {
	if len(common.AppConfig.BackupDirs) == 0 {
		fmt.Printf("%sAucun répertoire à modifier.%s\n", colorYellow, colorReset)
		return
	}
	
	idxStr := readInput("Numéro du répertoire à modifier: ")
	idx, err := strconv.Atoi(idxStr)
	
	if err != nil || idx < 1 || idx > len(common.AppConfig.BackupDirs) {
		fmt.Printf("%sNuméro invalide.%s\n", colorRed, colorReset)
		return
	}
	
	// Récupérer la configuration à modifier
	dir := common.AppConfig.BackupDirs[idx-1]
	
	fmt.Printf("%sModification de la configuration '%s'%s\n\n", colorBold, dir.Name, colorReset)
	
	// Permettre de modifier chaque propriété
	fmt.Printf("Nom actuel: %s\n", dir.Name)
	name := readInput("Nouveau nom (vide pour garder l'actuel): ")
	if name == "" {
		name = dir.Name
	}
	
	fmt.Printf("Chemin actuel: %s\n", dir.SourcePath)
	sourcePath := readInput("Nouveau chemin (vide pour garder l'actuel): ")
	if sourcePath == "" {
		sourcePath = dir.SourcePath
	} else {
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
			fmt.Printf("%sAttention: Le répertoire '%s' n'existe pas.%s\n", colorYellow, sourcePath, colorReset)
			confirm := readInput("Continuer quand même? (o/n): ")
			if strings.ToLower(confirm) != "o" {
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
		for i := range excludeDirs {
			excludeDirs[i] = strings.TrimSpace(excludeDirs[i])
		}
	}
	
	// Fichiers à exclure
	fmt.Printf("Fichiers exclus actuels: %s\n", strings.Join(dir.ExcludeFiles, ", "))
	excludeFilesStr := readInput("Nouveaux fichiers à exclure (séparés par des virgules, vide pour garder les actuels): ")
	excludeFiles := dir.ExcludeFiles
	if excludeFilesStr != "" {
		excludeFiles = strings.Split(excludeFilesStr, ",")
		for i := range excludeFiles {
			excludeFiles[i] = strings.TrimSpace(excludeFiles[i])
		}
	}
	
	// Intervalle
	fmt.Printf("Intervalle actuel: %d minutes\n", dir.Interval)
	intervalStr := readInput("Nouvel intervalle en minutes (vide pour garder l'actuel): ")
	interval := dir.Interval
	if intervalStr != "" {
		if i, err := strconv.Atoi(intervalStr); err == nil && i >= 0 {
			interval = i
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
		fmt.Printf("%sErreur lors de la mise à jour de la configuration: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	fmt.Printf("%sConfiguration '%s' modifiée avec succès.%s\n", colorGreen, name, colorReset)
}

// deleteBackupDirectory permet de supprimer un répertoire de sauvegarde
func deleteBackupDirectory() {
	if len(common.AppConfig.BackupDirs) == 0 {
		fmt.Printf("%sAucun répertoire à supprimer.%s\n", colorYellow, colorReset)
		return
	}
	
	idxStr := readInput("Numéro du répertoire à supprimer: ")
	idx, err := strconv.Atoi(idxStr)
	
	if err != nil || idx < 1 || idx > len(common.AppConfig.BackupDirs) {
		fmt.Printf("%sNuméro invalide.%s\n", colorRed, colorReset)
		return
	}
	
	// Récupérer le nom pour confirmation
	name := common.AppConfig.BackupDirs[idx-1].Name
	
	confirm := readInput(fmt.Sprintf("Êtes-vous sûr de vouloir supprimer la configuration '%s'? (o/n): ", name))
	if strings.ToLower(confirm) != "o" {
		fmt.Println("Suppression annulée.")
		return
	}
	
	// Supprimer l'élément
	common.AppConfig.BackupDirs = append(
		common.AppConfig.BackupDirs[:idx-1], 
		common.AppConfig.BackupDirs[idx:]...
	)
	
	if err := common.SaveConfig(common.AppConfig); err != nil {
		fmt.Printf("%sErreur lors de la mise à jour de la configuration: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	fmt.Printf("%sConfiguration '%s' supprimée avec succès.%s\n", colorGreen, name, colorReset)
}

// manageRsyncServers permet de gérer les serveurs rsync configurés
func manageRsyncServers() {
	for {
		clearScreen()
		fmt.Printf("%sGestion des serveurs rsync%s\n\n", colorBold, colorReset)
		
		// Afficher les serveurs configurés
		if len(common.AppConfig.RsyncServers) == 0 {
			fmt.Printf("%sAucun serveur rsync configuré.%s\n\n", colorYellow, colorReset)
		} else {
			fmt.Printf("%sServeurs configurés:%s\n", colorBold, colorReset)
			for i, server := range common.AppConfig.RsyncServers {
				fmt.Printf("%d. %s (%s)\n", i+1, server.Name, server.IP)
				fmt.Printf("   Utilisateur: %s, Port SSH: %d\n", server.Username, server.SSHPort)
				if len(server.Modules) > 0 {
					fmt.Printf("   Modules: %s\n", strings.Join(server.Modules, ", "))
				}
			}
			fmt.Println()
		}
		
		fmt.Printf("  %s1.%s Rechercher et ajouter un serveur\n", colorGreen, colorReset)
		fmt.Printf("  %s2.%s Modifier un serveur\n", colorGreen, colorReset)
		fmt.Printf("  %s3.%s Supprimer un serveur\n", colorGreen, colorReset)
		fmt.Printf("  %s0.%s Retour\n\n", colorGreen, colorReset)
		
		choice := readInput("Votre choix: ")
		
		switch choice {
		case "1":
			discoverRsyncServers() // Utiliser la fonction existante
			return
		case "2":
			editRsyncServer()
		case "3":
			deleteRsyncServer()
		case "0":
			return
		default:
			fmt.Println("Option non valide.")
		}
		
		readInput("Appuyez sur Entrée pour continuer...")
	}
}

// editRsyncServer permet de modifier un serveur rsync existant
func editRsyncServer() {
	if len(common.AppConfig.RsyncServers) == 0 {
		fmt.Printf("%sAucun serveur à modifier.%s\n", colorYellow, colorReset)
		return
	}
	
	idxStr := readInput("Numéro du serveur à modifier: ")
	idx, err := strconv.Atoi(idxStr)
	
	if err != nil || idx < 1 || idx > len(common.AppConfig.RsyncServers) {
		fmt.Printf("%sNuméro invalide.%s\n", colorRed, colorReset)
		return
	}
	
	// Récupérer la configuration à modifier
	server := common.AppConfig.RsyncServers[idx-1]
	
	fmt.Printf("%sModification du serveur '%s'%s\n\n", colorBold, server.Name, colorReset)
	
	// Permettre de modifier chaque propriété
	fmt.Printf("Nom actuel: %s\n", server.Name)
	name := readInput("Nouveau nom (vide pour garder l'actuel): ")
	if name == "" {
		name = server.Name
	}
	
	fmt.Printf("Adresse IP actuelle: %s\n", server.IP)
	ip := readInput("Nouvelle adresse IP (vide pour garder l'actuelle): ")
	if ip == "" {
		ip = server.IP
	}
	
	fmt.Printf("Nom d'utilisateur actuel: %s\n", server.Username)
	username := readInput("Nouveau nom d'utilisateur (vide pour garder l'actuel): ")
	if username == "" {
		username = server.Username
	}
	
	fmt.Printf("Port SSH actuel: %d\n", server.SSHPort)
	sshPortStr := readInput("Nouveau port SSH (vide pour garder l'actuel): ")
	sshPort := server.SSHPort
	if sshPortStr != "" {
		if port, err := strconv.Atoi(sshPortStr); err == nil && port > 0 {
			sshPort = port
		}
	}
	
	// Si des modules sont disponibles, permettre de modifier le module par défaut
	defaultModule := server.DefaultModule
	if len(server.Modules) > 0 {
		fmt.Printf("Modules disponibles: %s\n", strings.Join(server.Modules, ", "))
		fmt.Printf("Module par défaut actuel: %s\n", defaultModule)
		
		fmt.Println("\nVoulez-vous changer le module par défaut?")
		for i, module := range server.Modules {
			fmt.Printf("%d. %s\n", i+1, module)
		}
		fmt.Printf("0. Aucun module par défaut\n")
		
		moduleChoice := readInput("\nChoisissez un module (vide pour garder l'actuel): ")
		
		if moduleChoice != "" {
			if moduleIdx, err := strconv.Atoi(moduleChoice); err == nil {
				if moduleIdx == 0 {
					defaultModule = ""
				} else if moduleIdx > 0 && moduleIdx <= len(server.Modules) {
					defaultModule = server.Modules[moduleIdx-1]
				}
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
			common.AppConfig.BackupDirs[i].RemoteServer = &updatedServer
		}
	}
	
	if err := common.SaveConfig(common.AppConfig); err != nil {
		fmt.Printf("%sErreur lors de la mise à jour de la configuration: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	fmt.Printf("%sServeur '%s' modifié avec succès.%s\n", colorGreen, name, colorReset)
}

// deleteRsyncServer permet de supprimer un serveur rsync
func deleteRsyncServer() {
	if len(common.AppConfig.RsyncServers) == 0 {
		fmt.Printf("%sAucun serveur à supprimer.%s\n", colorYellow, colorReset)
		return
	}
	
	idxStr := readInput("Numéro du serveur à supprimer: ")
	idx, err := strconv.Atoi(idxStr)
	
	if err != nil || idx < 1 || idx > len(common.AppConfig.RsyncServers) {
		fmt.Printf("%sNuméro invalide.%s\n", colorRed, colorReset)
		return
	}
	
	// Récupérer le serveur pour vérifier les dépendances
	server := common.AppConfig.RsyncServers[idx-1]
	
	// Vérifier si le serveur est utilisé par des configurations de sauvegarde
	var usedBy []string
	for _, dir := range common.AppConfig.BackupDirs {
		if dir.RemoteServer != nil && dir.RemoteServer.Name == server.Name {
			usedBy = append(usedBy, dir.Name)
		}
	}
	
	if len(usedBy) > 0 {
		fmt.Printf("%sAttention: Ce serveur est utilisé par les configurations suivantes:%s\n", 
			colorYellow, colorReset)
		for _, name := range usedBy {
			fmt.Printf("- %s\n", name)
		}
		fmt.Println("La suppression du serveur affectera ces configurations.")
	}
	
	confirm := readInput(fmt.Sprintf("Êtes-vous sûr de vouloir supprimer le serveur '%s'? (o/n): ", server.Name))
	if strings.ToLower(confirm) != "o" {
		fmt.Println("Suppression annulée.")
		return
	}
	
	// Supprimer l'élément
	common.AppConfig.RsyncServers = append(
		common.AppConfig.RsyncServers[:idx-1], 
		common.AppConfig.RsyncServers[idx:]...
	)
	
	// Mettre à jour les références dans les configurations de sauvegarde
	for i, dir := range common.AppConfig.BackupDirs {
		if dir.RemoteServer != nil && dir.RemoteServer.Name == server.Name {
			common.AppConfig.BackupDirs[i].RemoteServer = nil
		}
	}
	
	if err := common.SaveConfig(common.AppConfig); err != nil {
		fmt.Printf("%sErreur lors de la mise à jour de la configuration: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	fmt.Printf("%sServeur '%s' supprimé avec succès.%s\n", colorGreen, server.Name, colorReset)
}

// manageRetentionPolicy permet de modifier la politique de rétention
func manageRetentionPolicy() {
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
		}
	}
	
	keepWeeklyStr := readInput(fmt.Sprintf("Nombre de sauvegardes hebdomadaires à conserver (actuel: %d): ", policy.KeepWeekly))
	keepWeekly := policy.KeepWeekly
	if keepWeeklyStr != "" {
		if weeks, err := strconv.Atoi(keepWeeklyStr); err == nil && weeks >= 0 {
			keepWeekly = weeks
		}
	}
	
	keepMonthlyStr := readInput(fmt.Sprintf("Nombre de sauvegardes mensuelles à conserver (actuel: %d): ", policy.KeepMonthly))
	keepMonthly := policy.KeepMonthly
	if keepMonthlyStr != "" {
		if months, err := strconv.Atoi(keepMonthlyStr); err == nil && months >= 0 {
			keepMonthly = months
		}
	}
	
	// Mettre à jour la politique
	common.AppConfig.RetentionPolicy = common.RetentionPolicy{
		KeepDaily:   keepDaily,
		KeepWeekly:  keepWeekly,
		KeepMonthly: keepMonthly,
	}
	
	if err := common.SaveConfig(common.AppConfig); err != nil {
		fmt.Printf("%sErreur lors de la mise à jour de la configuration: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	fmt.Printf("%sPolitique de rétention mise à jour avec succès.%s\n", colorGreen, colorReset)
}

// changeBackupDestination permet de modifier la destination des sauvegardes
func changeBackupDestination() {
	clearScreen()
	fmt.Printf("%sModification de la destination des sauvegardes%s\n\n", colorBold, colorReset)
	
	fmt.Printf("Destination actuelle: %s\n\n", common.AppConfig.BackupDestination)
	
	// Demander la nouvelle destination
	newDest := readInput("Nouvelle destination (vide pour annuler): ")
	if newDest == "" {
		fmt.Println("Modification annulée.")
		return
	}
	
	// Expandir les chemins relatifs, y compris ~/
	if strings.HasPrefix(newDest, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			newDest = filepath.Join(homeDir, newDest[2:])
		}
	}
	
	// Normaliser le chemin
	newDest = filepath.Clean(newDest)
	
	// Vérifier/créer le répertoire de destination
	if !common.DirExists(newDest) {
		createDest := readInput("Le répertoire n'existe pas. Voulez-vous le créer? (o/n): ")
		if strings.ToLower(createDest) == "o" {
			if err := os.MkdirAll(newDest, 0755); err != nil {
				fmt.Printf("%sErreur lors de la création du répertoire: %v%s\n", colorRed, err, colorReset)
				return
			}
		} else {
			fmt.Println("Modification annulée.")
			return
		}
	}
	
	// Mettre à jour la destination
	common.AppConfig.BackupDestination = newDest
	
	if err := common.SaveConfig(common.AppConfig); err != nil {
		fmt.Printf("%sErreur lors de la mise à jour de la configuration: %v%s\n", colorRed, err, colorReset)
		return
	}
	
	fmt.Printf("%sDestination des sauvegardes mise à jour avec succès.%s\n", colorGreen, colorReset)
}