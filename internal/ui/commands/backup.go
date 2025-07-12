package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Noziop/s4v3my4ss/internal/ui/display"
	"github.com/Noziop/s4v3my4ss/internal/ui/input"
	"github.com/Noziop/s4v3my4ss/internal/watch"
	"github.com/Noziop/s4v3my4ss/pkg/common"
)

// ConfigureBackup permet de configurer une nouvelle sauvegarde
func ConfigureBackup() {
	common.LogInfo("Début de la configuration d'une nouvelle sauvegarde.")
	fmt.Printf("%sConfiguration d'une nouvelle sauvegarde%s\n\n", display.ColorBold(), display.ColorReset())

	// SECURITY: Utiliser la validation pour les entrées utilisateur.
	name := input.ReadAndValidateInput("Nom de la sauvegarde: ", common.IsValidName, "Nom invalide. Utilisez uniquement des lettres, chiffres, - et _.") // Utilisation de common.IsValidName
	sourcePath := input.ReadAndValidateInput("Chemin du répertoire à surveiller: ", common.IsValidPath, "Chemin invalide ou non sécurisé.") // Utilisation de common.IsValidPath

	if sourcePath == "" {
		common.LogWarning("Configuration annulée: chemin source vide.")
		fmt.Println("Configuration annulée. Le chemin source ne peut pas être être vide.")
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
		fmt.Printf("%sErreur: Le répertoire '%s' n'existe pas.%s\n", display.ColorRed, sourcePath, display.ColorReset)
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
		
		destChoiceStr := input.ReadInput("\nEmplacement des sauvegardes (numéro ou vide pour utiliser l'emplacement par défaut): ")
		
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
		backupDestination = input.ReadAndValidateInput("Nouvel emplacement des sauvegardes: ", common.IsValidPath, "Chemin invalide ou non sécurisé.") // Utilisation de common.IsValidPath
		
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
				createDest := input.ReadInput("")
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
			newDestName := input.ReadAndValidateInput("Nom de la nouvelle destination: ", common.IsValidName, "Nom invalide.") // Utilisation de common.IsValidName
			if newDestName == "" {
				newDestName = "Destination " + strconv.Itoa(len(destinations)+1)
			}
			
			// Demander si cette destination doit être la destination par défaut
			defaultDestStr := input.ReadInput("Définir comme destination par défaut? (o/n): ")
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
	incrementalStr := input.ReadInput("Activer les sauvegardes incrémentales? (o/n): ")
	incremental := strings.ToLower(incrementalStr) == "o"
	
	// Compression
	compressStr := input.ReadInput("Activer la compression? (o/n): ")
	compression := strings.ToLower(compressStr) == "o"
	
	// Répertoires à exclure
	excludeDirsStr := input.ReadInput("Répertoires à exclure (séparés par des virgules): ")
	var excludeDirs []string
	if excludeDirsStr != "" {
		excludeDirs = strings.Split(excludeDirsStr, ",")
		for i, dir := range excludeDirs {
			excludeDirs[i] = strings.TrimSpace(dir)
			if !common.IsValidExcludePattern(excludeDirs[i]) { // Utilisation de common.IsValidExcludePattern
				common.LogError("Modèle d'exclusion de répertoire invalide: %s", excludeDirs[i])
				fmt.Printf("%sModèle d'exclusion de répertoire invalide: %s%s\n", display.ColorRed, excludeDirs[i], display.ColorReset)
				return
			}
		}
	} else {
		excludeDirs = commonExcludeDirs
	}
	
	// Fichiers à exclure
	excludeFilesStr := input.ReadInput("Fichiers à exclure (séparés par des virgules): ")
	var excludeFiles []string
	if excludeFilesStr != "" {
		excludeFiles = strings.Split(excludeFilesStr, ",")
		for i, file := range excludeFiles {
			excludeFiles[i] = strings.TrimSpace(file)
			if !common.IsValidExcludePattern(excludeFiles[i]) { // Utilisation de common.IsValidExcludePattern
				common.LogError("Modèle d'exclusion de fichier invalide: %s", excludeFiles[i])
				fmt.Printf("%sModèle d'exclusion de fichier invalide: %s%s\n", display.ColorRed, excludeFiles[i], display.ColorReset)
				return
			}
		}
	} else {
		excludeFiles = commonExcludeFiles
	}
	
	// Intervalle (0 = pas de surveillance automatique)
	intervalStr := input.ReadInput("Intervalle de sauvegarde en minutes (0 pour désactiver): ")
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
		fmt.Printf("%sErreur lors de l'ajout de la configuration: %v%s\n", display.ColorRed(), err, display.ColorReset())
		return
	}
	
	common.LogInfo("Configuration de sauvegarde %s ajoutée avec succès.", config.Name)
	fmt.Printf("%sConfiguration ajoutée avec succès.%s\n", display.ColorGreen(), display.ColorReset())
	
	// Proposer de démarrer la surveillance
	startWatchStr := input.ReadInput("Démarrer la surveillance maintenant? (o/n): ")
	if strings.ToLower(startWatchStr) == "o" {
		common.LogInfo("Démarrage de la surveillance pour %s.", config.Name)
		fmt.Printf("Démarrage de la surveillance du répertoire: %s\n", config.SourcePath)
		fmt.Println("Mode surveillance continue. La surveillance s'exécute en arrière-plan.")
		fmt.Println("Les sauvegardes continueront même si vous revenez au menu principal.")
		
		// Lancer la surveillance en arrière-plan pour ne pas bloquer l'interface
		go func() {
			if err := watch.StartWatch(config); err != nil {
				common.LogError("Erreur lors de la surveillance en arrière-plan pour %s: %v", config.Name, err)
				fmt.Printf("%sErreur lors de la surveillance: %v%s\n", display.ColorRed(), err, display.ColorReset())
			}
		}()
	}
}