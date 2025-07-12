package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Noziop/s4v3my4ss/internal/ui/display"
	"github.com/Noziop/s4v3my4ss/internal/ui/input"
	"github.com/Noziop/s4v3my4ss/pkg/common"
)




// ManageConfiguration permet de gérer la configuration de l'application
func ManageConfiguration() {
	common.LogInfo("Début de la gestion de la configuration.")
	for {
		display.ClearScreen()
		fmt.Printf("%sGestion de la Configuration%s\n\n", display.ColorBold(), display.ColorReset())

		fmt.Printf("  %s1.%s Afficher la configuration complète\n", display.ColorGreen(), display.ColorReset())
		fmt.Printf("  %s2.%s Modifier les répertoires sauvegardés\n", display.ColorGreen(), display.ColorReset())
		fmt.Printf("  %s3.%s Modifier/supprimer les serveurs rsync\n", display.ColorGreen(), display.ColorReset())
		fmt.Printf("  %s4.%s Modifier la politique de rétention\n", display.ColorGreen(), display.ColorReset())
		fmt.Printf("  %s5.%s Gérer les destinations de sauvegarde\n", display.ColorGreen(), display.ColorReset())
		fmt.Printf("  %s6.%s Modifier la destination principale\n", display.ColorGreen(), display.ColorReset())
		fmt.Printf("  %s0.%s Retour au menu principal\n", display.ColorGreen(), display.ColorReset())

		choice := input.ReadInput("Votre choix: ")

		switch choice {
		case "1":
			displayFullConfig()
			input.ReadInput("Appuyez sur Entrée pour continuer...")
		case "2":
			manageBackupDirectories()
		case "3":
			manageRsyncServers()
		case "4":
			manageRetentionPolicy()
			input.ReadInput("Appuyez sur Entrée pour continuer...")
		case "5":
			manageBackupDestinations()
		case "6":
			changeBackupDestination()
			input.ReadInput("Appuyez sur Entrée pour continuer...")
		case "0":
			common.LogInfo("Retour au menu principal depuis la gestion de la configuration.")
			return
		default:
			common.LogWarning("Option de gestion de configuration non valide: %s", choice)
			fmt.Println("Option non valide. Veuillez réessayer.")
			input.ReadInput("Appuyez sur Entrée pour continuer...")
		}
	}
}

// displayFullConfig affiche la configuration complète de l'application
func displayFullConfig() {
	common.LogInfo("Affichage de la configuration complète.")
	display.ClearScreen()
	fmt.Printf("%sConfiguration Complète de l'Application%s\n\n", display.ColorBold(), display.ColorReset())

	// Afficher les informations générales
	fmt.Printf("%sInformations générales:%s", display.ColorBold(), display.ColorReset())
	fmt.Printf("Destination des sauvegardes: %s\n", common.AppConfig.BackupDestination)
	fmt.Printf("Dernière mise à jour: %s\n\n", common.AppConfig.LastUpdate.Format("02/01/2006 15:04:05"))

	// Afficher la politique de rétention
	fmt.Printf("%sPolitique de rétention:%s\n", display.ColorBold(), display.ColorReset())
	fmt.Printf("Conservation quotidienne: %d jours\n", common.AppConfig.RetentionPolicy.KeepDaily)
	fmt.Printf("Conservation hebdomadaire: %d semaines\n", common.AppConfig.RetentionPolicy.KeepWeekly)
	fmt.Printf("Conservation mensuelle: %d mois\n\n", common.AppConfig.RetentionPolicy.KeepMonthly)

	// Afficher les répertoires sauvegardés
	display.DisplayConfigList(common.AppConfig.BackupDirs, "Répertoires sauvegardés", func(i int, item interface{}) string {
		dir := item.(common.BackupConfig)
		incr := "Non"
		if dir.IsIncremental {
			incr = "Oui"
		}
		comp := "Non"
		if dir.Compression {
			comp = "Oui"
		}
		
		out := fmt.Sprintf("%d. %s (%s)\n", i+1, dir.Name, dir.SourcePath)
		out += fmt.Sprintf("   Compression: %s, Incrémental: %s, Intervalle: %d min\n", comp, incr, dir.Interval)

		// Afficher les exclusions si présentes
		if len(dir.ExcludeDirs) > 0 || len(dir.ExcludeFiles) > 0 {
			out += "   Exclusions:\n"
			if len(dir.ExcludeDirs) > 0 {
				out += fmt.Sprintf("   - Répertoires: %s\n", strings.Join(dir.ExcludeDirs, ", "))
			}
			if len(dir.ExcludeFiles) > 0 {
				out += fmt.Sprintf("   - Fichiers: %s\n", strings.Join(dir.ExcludeFiles, ", "))
			}
		}

		// Afficher le serveur distant si configuré
		if dir.RemoteServer != nil {
			out += fmt.Sprintf("   Serveur distant: %s (%s)\n", dir.RemoteServer.Name, dir.RemoteServer.IP)
			if dir.RemoteServer.DefaultModule != "" {
				out += fmt.Sprintf("   Module: %s\n", dir.RemoteServer.DefaultModule)
			}
		}
		return out
	})

	// Afficher les serveurs rsync
	display.DisplayConfigList(common.AppConfig.RsyncServers, "Serveurs rsync configurés", func(i int, item interface{}) string {
		server := item.(common.RsyncServerConfig)
		out := fmt.Sprintf("%d. %s (%s)\n", i+1, server.Name, server.IP)
		out += fmt.Sprintf("   Port: %d, Port SSH: %d, Utilisateur: %s\n", server.Port, server.SSHPort, server.Username)
		if len(server.Modules) > 0 {
			out += fmt.Sprintf("   Modules disponibles: %s\n", strings.Join(server.Modules, ", "))
			if server.DefaultModule != "" {
				out += fmt.Sprintf("   Module par défaut: %s\n", server.DefaultModule)
			}
		}
		return out
	})
	common.LogInfo("Configuration complète affichée.")
}

// manageBackupDirectories permet de gérer les répertoires à sauvegarder
func manageBackupDirectories() {
	common.LogInfo("Début de la gestion des répertoires de sauvegarde.")
	for {
		display.ClearScreen()
		fmt.Printf("%sGestion des répertoires sauvegardés%s\n\n", display.ColorBold(), display.ColorReset())

		// Afficher les répertoires sauvegardés
        display.DisplayConfigList(common.AppConfig.BackupDirs, "Répertoires configurés", func(i int, item interface{}) string {
            dir := item.(common.BackupConfig)
            incr := "Non"
            if dir.IsIncremental {
                incr = "Oui"
            }
            comp := "Non"
            if dir.Compression {
                comp = "Oui"
            }
            return fmt.Sprintf("%d. %s (%s)   Compression: %s, Incrémental: %s, Intervalle: %d min",
                i+1, dir.Name, dir.SourcePath, comp, incr, dir.Interval)
        })

		fmt.Printf("  %s1.%s Ajouter un répertoire\n", display.ColorGreen(), display.ColorReset())
		fmt.Printf("  %s2.%s Modifier un répertoire\n", display.ColorGreen(), display.ColorReset())
		fmt.Printf("  %s3.%s Supprimer un répertoire\n", display.ColorGreen(), display.ColorReset())
		fmt.Printf("  %s0.%s Retour\n\n", display.ColorGreen(), display.ColorReset())

		choice := input.ReadInput("Votre choix: ")

		switch choice {
		case "1":
			//configureBackup() // Utiliser la fonction existante
			common.LogInfo("Ajout d'un répertoire de sauvegarde demandé.")
			return
		case "2":
			editBackupDirectory()
			input.ReadInput("Appuyez sur Entrée pour continuer...")
		case "3":
			deleteBackupDirectory()
			input.ReadInput("Appuyez sur Entrée pour continuer...")
		case "0":
			common.LogInfo("Retour au menu de gestion de la configuration.")
			return
		default:
			common.LogWarning("Option de gestion des répertoires de sauvegarde non valide: %s", choice)
			fmt.Println("Option non valide.")
			input.ReadInput("Appuyez sur Entrée pour continuer...")
		}
	}
}

// editBackupDirectory permet de modifier un répertoire de sauvegarde existant
func editBackupDirectory() {
	common.LogInfo("Début de la modification d'un répertoire de sauvegarde.")
	if len(common.AppConfig.BackupDirs) == 0 {
		input.DisplayMessage(false, "Aucun répertoire à modifier.")
		return
	}

	idxStr := input.ReadInput("Numéro du répertoire à modifier: ")
	idx, err := strconv.Atoi(idxStr)

	if err != nil || idx < 1 || idx > len(common.AppConfig.BackupDirs) {
		input.DisplayMessage(true, "Numéro invalide.")
		return
	}

	// Récupérer la configuration à modifier
	dir := common.AppConfig.BackupDirs[idx-1]
	common.LogInfo("Modification du répertoire de sauvegarde: %s.", dir.Name)
	fmt.Printf("%sModification de la configuration '%s'%s\n\n", display.ColorBold(), dir.Name, display.ColorReset())

	// Permettre de modifier chaque propriété
	name := input.ReadStringInput(fmt.Sprintf("Nouveau nom (actuel: %s): ", dir.Name), dir.Name, common.IsValidName, "Nom invalide.")
	sourcePath := input.ReadStringInput(fmt.Sprintf("Nouveau chemin (actuel: %s): ", dir.SourcePath), dir.SourcePath, common.IsValidPath, "Chemin invalide.")
	incremental := input.ReadBoolInput("Activer les sauvegardes incrémentales?", dir.IsIncremental)
	compression := input.ReadBoolInput("Activer la compression?", dir.Compression)

	// Répertoires à exclure
	fmt.Printf("Répertoires exclus actuels: %s\n", strings.Join(dir.ExcludeDirs, ", "))
	excludeDirsStr := input.ReadInput("Nouveaux répertoires à exclure (séparés par des virgules, vide pour garder les actuels): ")
	excludeDirs := dir.ExcludeDirs
	if excludeDirsStr != "" {
		excludeDirs = strings.Split(excludeDirsStr, ",")
		for i, d := range excludeDirs {
			excludeDirs[i] = strings.TrimSpace(d)
		}
	}

	// Fichiers à exclure
	fmt.Printf("Fichiers exclus actuels: %s\n", strings.Join(dir.ExcludeFiles, ", "))
	excludeFilesStr := input.ReadInput("Nouveaux fichiers à exclure (séparés par des virgules, vide pour garder les actuels): ")
	excludeFiles := dir.ExcludeFiles
	if excludeFilesStr != "" {
		excludeFiles = strings.Split(excludeFilesStr, ",")
		for i, f := range excludeFiles {
			excludeFiles[i] = strings.TrimSpace(f)
		}
	}

	// Intervalle
	interval := input.ReadIntInput("Nouvel intervalle en minutes", dir.Interval)

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
		input.DisplayMessage(true, "Erreur lors de la mise à jour de la configuration du répertoire %s: %v", name, err)
		return
	}

	input.DisplayMessage(false, "Configuration '%s' modifiée avec succès.", name)
}

// deleteBackupDirectory permet de supprimer un répertoire de sauvegarde
func deleteBackupDirectory() {
	common.LogInfo("Début de la suppression d'un répertoire de sauvegarde.")
	if len(common.AppConfig.BackupDirs) == 0 {
		input.DisplayMessage(false, "Aucun répertoire à supprimer.")
		return
	}

	idxStr := input.ReadInput("Numéro du répertoire à supprimer: ")
	idx, err := strconv.Atoi(idxStr)

	if err != nil || idx < 1 || idx > len(common.AppConfig.BackupDirs) {
		input.DisplayMessage(true, "Numéro invalide.")
		return
	}

	// Récupérer le nom pour confirmation
	name := common.AppConfig.BackupDirs[idx-1].Name

	if !input.ConfirmAction(fmt.Sprintf("Êtes-vous sûr de vouloir supprimer la configuration '%s'?", name)) {
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
		input.DisplayMessage(true, "Erreur lors de la suppression de la configuration du répertoire %s: %v", name, err)
		return
	}

	input.DisplayMessage(false, "Configuration '%s' supprimée avec succès.", name)
}

// manageRsyncServers permet de gérer les serveurs rsync configurés
func manageRsyncServers() {
	common.LogInfo("Début de la gestion des serveurs rsync.")
	for {
		display.ClearScreen()
		fmt.Printf("%sGestion des serveurs rsync%s\n\n", display.ColorBold(), display.ColorReset())

		// Afficher les serveurs configurés
        display.DisplayConfigList(common.AppConfig.RsyncServers, "Serveurs configurés", func(i int, item interface{}) string {
            server := item.(common.RsyncServerConfig)
            out := fmt.Sprintf("%d. %s (%s)\n", i+1, server.Name, server.IP)
            out += fmt.Sprintf("   Port: %d, Port SSH: %d, Utilisateur: %s\n", server.Port, server.SSHPort, server.Username)
            if len(server.Modules) > 0 {
                out += fmt.Sprintf("   Modules: %s\n", strings.Join(server.Modules, ", "))
            }
            return out
        })

		fmt.Printf("  %s1.%s Rechercher et ajouter un serveur\n", display.ColorGreen(), display.ColorReset())
		fmt.Printf("  %s2.%s Modifier un serveur\n", display.ColorGreen(), display.ColorReset())
		fmt.Printf("  %s3.%s Supprimer un serveur\n", display.ColorGreen(), display.ColorReset())
		fmt.Printf("  %s0.%s Retour\n\n", display.ColorGreen(), display.ColorReset())

		choice := input.ReadInput("Votre choix: ")

		switch choice {
		case "1":
			//discoverRsyncServers()
			common.LogInfo("Recherche et ajout de serveur rsync demandé.")
			return
		case "2":
			editRsyncServer()
			input.ReadInput("Appuyez sur Entrée pour continuer...")
		case "3":
			deleteRsyncServer()
			input.ReadInput("Appuyez sur Entrée pour continuer...")
		case "0":
			common.LogInfo("Retour au menu de gestion de la configuration.")
			return
		default:
			common.LogWarning("Option de gestion des serveurs rsync non valide: %s", choice)
			fmt.Println("Option non valide. Veuillez réessayer.")
			input.ReadInput("Appuyez sur Entrée pour continuer...")
		}
	}
}

// editRsyncServer permet de modifier un serveur rsync existant
func editRsyncServer() {
	common.LogInfo("Début de la modification d'un serveur rsync.")
	if len(common.AppConfig.RsyncServers) == 0 {
		input.DisplayMessage(false, "Aucun serveur rsync à modifier.")
		return
	}

	idxStr := input.ReadInput("Numéro du serveur à modifier: ")
	idx, err := strconv.Atoi(idxStr)

	if err != nil || idx < 1 || idx > len(common.AppConfig.RsyncServers) {
		input.DisplayMessage(true, "Numéro invalide.")
		return
	}

	// Récupérer la configuration à modifier
	server := common.AppConfig.RsyncServers[idx-1]
	common.LogInfo("Modification du serveur rsync: %s.", server.Name)
	fmt.Printf("%sModification du serveur '%s'%s\n\n", display.ColorBold(), server.Name, display.ColorReset())

	// Permettre de modifier chaque propriété
	name := input.ReadStringInput(fmt.Sprintf("Nouveau nom (actuel: %s): ", server.Name), server.Name, common.IsValidName, "Nom invalide.")
	ip := input.ReadStringInput(fmt.Sprintf("Nouvelle adresse IP (actuel: %s): ", server.IP), server.IP, func(s string) bool { return s != "" }, "Adresse IP invalide.") // Simple validation non vide
	username := input.ReadStringInput(fmt.Sprintf("Nouveau nom d'utilisateur (actuel: %s): ", server.Username), server.Username, common.IsValidName, "Nom d'utilisateur invalide.")
	sshPort := input.ReadIntInput(fmt.Sprintf("Nouveau port SSH (actuel: %d)", server.SSHPort), server.SSHPort)

	// Si des modules sont disponibles, permettre de modifier le module par défaut
	defaultModule := server.DefaultModule
	if len(server.Modules) > 0 {
		fmt.Printf("Modules disponibles: %s\n", strings.Join(server.Modules, ", "))
		fmt.Printf("Module par défaut actuel: %s\n", defaultModule)

		fmt.Println("\nVoulez-vous changer le module par défaut?")
		moduleChoice := input.ReadInput("\nChoisissez un module (vide pour garder l'actuel): ")

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
		input.DisplayMessage(true, "Erreur lors de la mise à jour du serveur rsync %s: %v", name, err)
		return
	}

	input.DisplayMessage(false, "Serveur '%s' modifié avec succès.", name)
}

// deleteRsyncServer permet de supprimer un serveur rsync
func deleteRsyncServer() {
	common.LogInfo("Début de la suppression d'un serveur rsync.")
	if len(common.AppConfig.RsyncServers) == 0 {
		input.DisplayMessage(false, "Aucun serveur rsync à supprimer.")
		return
	}

	idxStr := input.ReadInput("Numéro du serveur à supprimer: ")
	idx, err := strconv.Atoi(idxStr)

	if err != nil || idx < 1 || idx > len(common.AppConfig.RsyncServers) {
		input.DisplayMessage(true, "Numéro invalide.")
		return
	}

	// Récupérer le nom pour confirmation
	name := common.AppConfig.RsyncServers[idx-1].Name

	if !input.ConfirmAction(fmt.Sprintf("Êtes-vous sûr de vouloir supprimer le serveur '%s'?", name)) {
		common.LogInfo("Suppression du serveur rsync annulée par l'utilisateur pour %s.", name)
		fmt.Println("Suppression annulée.")
		return
	}

	// Supprimer l'élément
	common.AppConfig.RsyncServers = append(
		common.AppConfig.RsyncServers[:idx-1],
		common.AppConfig.RsyncServers[idx:]...,
	)

	if err := common.SaveConfig(common.AppConfig); err != nil {
		input.DisplayMessage(true, "Erreur lors de la suppression du serveur rsync %s: %v", name, err)
		return
	}

	input.DisplayMessage(false, "Serveur '%s' supprimé avec succès.", name)
}

// manageRetentionPolicy permet de modifier la politique de rétention
func manageRetentionPolicy() {
	common.LogInfo("Début de la modification de la politique de rétention.")
	display.ClearScreen()
	fmt.Printf("%sModification de la Politique de Rétention%s\n\n", display.ColorBold(), display.ColorReset())

	// Afficher la politique actuelle
	fmt.Printf("Politique actuelle:\n")
	fmt.Printf("  Conservation quotidienne: %d jours\n", common.AppConfig.RetentionPolicy.KeepDaily)
	fmt.Printf("  Conservation hebdomadaire: %d semaines\n", common.AppConfig.RetentionPolicy.KeepWeekly)
	fmt.Printf("  Conservation mensuelle: %d mois\n\n", common.AppConfig.RetentionPolicy.KeepMonthly)

	// Demander les nouvelles valeurs
	daily := input.ReadIntInput("Nouvelle conservation quotidienne", common.AppConfig.RetentionPolicy.KeepDaily)
	weekly := input.ReadIntInput("Nouvelle conservation hebdomadaire", common.AppConfig.RetentionPolicy.KeepWeekly)
	monthly := input.ReadIntInput("Nouvelle conservation mensuelle", common.AppConfig.RetentionPolicy.KeepMonthly)

	common.AppConfig.RetentionPolicy.KeepDaily = daily
	common.AppConfig.RetentionPolicy.KeepWeekly = weekly
	common.AppConfig.RetentionPolicy.KeepMonthly = monthly

	// Sauvegarder la configuration
	if err := common.SaveConfig(common.AppConfig); err != nil {
		input.DisplayMessage(true, "Erreur lors de la sauvegarde de la politique de rétention: %v", err)
		return
	}

	input.DisplayMessage(false, "Politique de rétention mise à jour avec succès.")
}

// manageBackupDestinations permet de gérer les destinations de sauvegarde
func manageBackupDestinations() {
	common.LogInfo("Début de la gestion des destinations de sauvegarde.")
	for {
		display.ClearScreen()
		fmt.Printf("%sGestion des Destinations de Sauvegarde%s\n\n", display.ColorBold(), display.ColorReset())

		// Afficher les destinations existantes
		display.DisplayConfigList(common.AppConfig.BackupDestinations, "Destinations configurées", func(i int, item interface{}) string {
			dest := item.(common.BackupDestination)
			defaultStr := ""
			if dest.IsDefault {
				defaultStr = " (Par défaut)"
			}
			return fmt.Sprintf("%d. %s (%s) - %s%s", i+1, dest.Name, dest.Type, dest.Path, defaultStr)
		})

		fmt.Printf("  %s1.%s Ajouter une destination\n", display.ColorGreen(), display.ColorReset())
		fmt.Printf("  %s2.%s Modifier une destination\n", display.ColorGreen(), display.ColorReset())
		fmt.Printf("  %s3.%s Supprimer une destination\n", display.ColorGreen(), display.ColorReset())
		fmt.Printf("  %s4.%s Définir une destination par défaut\n", display.ColorGreen(), display.ColorReset())
		fmt.Printf("  %s0.%s Retour\n\n", display.ColorGreen(), display.ColorReset())

		choice := input.ReadInput("Votre choix: ")

		switch choice {
		case "1":
			addBackupDestination()
			input.ReadInput("Appuyez sur Entrée pour continuer...")
		case "2":
			editBackupDestination()
			input.ReadInput("Appuyez sur Entrée pour continuer...")
		case "3":
			deleteBackupDestination()
			input.ReadInput("Appuyez sur Entrée pour continuer...")
		case "4":
			setDefaultBackupDestination()
			input.ReadInput("Appuyez sur Entrée pour continuer...")
		case "0":
			common.LogInfo("Retour au menu de gestion de la configuration.")
			return
		default:
			common.LogWarning("Option de gestion des destinations de sauvegarde non valide: %s", choice)
			fmt.Println("Option non valide.")
			input.ReadInput("Appuyez sur Entrée pour continuer...")
		}
	}
}

// addBackupDestination permet d'ajouter une nouvelle destination de sauvegarde
func addBackupDestination() {
	common.LogInfo("Ajout d'une nouvelle destination de sauvegarde.")
	fmt.Printf("\n%sAjout d'une nouvelle destination%s\n\n", display.ColorBold(), display.ColorReset())

	name := input.ReadAndValidateInput("Nom de la destination: ", common.IsValidName, "Nom invalide.")
	path := input.ReadAndValidateInput("Chemin de la destination: ", common.IsValidPath, "Chemin invalide.")

	destType := "local"
	if strings.HasPrefix(path, "rsync://") {
		destType = "rsync"
	}

	isDefault := input.ReadBoolInput("Définir comme destination par défaut?", false)

	newDest := common.BackupDestination{
		Name:      name,
		Path:      path,
		Type:      destType,
		IsDefault: isDefault,
	}

	if err := common.AddBackupDestination(newDest); err != nil{
		input.DisplayMessage(true, "Erreur lors de l'ajout de la destination de sauvegarde %s: %v", name, err)
		return
	}

	input.DisplayMessage(false, "Destination '%s' ajoutée avec succès.", name)
}

// editBackupDestination permet de modifier une destination de sauvegarde existante
func editBackupDestination() {
	common.LogInfo("Début de la modification d'une destination de sauvegarde.")
	if len(common.AppConfig.BackupDestinations) == 0 {
		input.DisplayMessage(false, "Aucune destination de sauvegarde à modifier.")
		return
	}

	idxStr := input.ReadInput("Numéro de la destination à modifier: ")
	idx, err := strconv.Atoi(idxStr)

	if err != nil || idx < 1 || idx > len(common.AppConfig.BackupDestinations) {
		input.DisplayMessage(true, "Numéro invalide.")
		return
	}

	dest := common.AppConfig.BackupDestinations[idx-1]
	common.LogInfo("Modification de la destination de sauvegarde: %s.", dest.Name)
	fmt.Printf("\n%sModification de la destination '%s'%s\n\n", display.ColorBold(), dest.Name, display.ColorReset())

	name := input.ReadStringInput(fmt.Sprintf("Nouveau nom (actuel: %s): ", dest.Name), dest.Name, common.IsValidName, "Nom invalide.")
	path := input.ReadStringInput(fmt.Sprintf("Nouveau chemin (actuel: %s): ", dest.Path), dest.Path, common.IsValidPath, "Chemin invalide.")

	destType := dest.Type
	if strings.HasPrefix(path, "rsync://") {
		destType = "rsync"
	} else {
		destType = "local"
	}

	isDefault := input.ReadBoolInput(fmt.Sprintf("Définir comme destination par défaut? (actuel: %t)", dest.IsDefault), dest.IsDefault)

	updatedDest := common.BackupDestination{
		Name:      name,
		Path:      path,
		Type:      destType,
		IsDefault: isDefault,
	}

	if err := common.UpdateBackupDestination(dest.Name, updatedDest); err != nil {
		input.DisplayMessage(true, "Erreur lors de la mise à jour de la destination de sauvegarde %s: %v", name, err)
		return
	}

	input.DisplayMessage(false, "Destination '%s' mise à jour avec succès.", name)
}
// deleteBackupDestination permet de supprimer une destination de sauvegarde
func deleteBackupDestination() {
	common.LogInfo("Début de la suppression d'une destination de sauvegarde.")
	if len(common.AppConfig.BackupDestinations) == 0 {
		input.DisplayMessage(false, "Aucune destination de sauvegarde à supprimer.")
		return
	}

	idxStr := input.ReadInput("Numéro de la destination à supprimer: ")
	idx, err := strconv.Atoi(idxStr)

	if err != nil || idx < 1 || idx > len(common.AppConfig.BackupDestinations) {
		input.DisplayMessage(true, "Numéro invalide.")
		return
	}

	name := common.AppConfig.BackupDestinations[idx-1].Name

	if !input.ConfirmAction(fmt.Sprintf("Êtes-vous sûr de vouloir supprimer la destination '%s'?", name)) {
		common.LogInfo("Suppression de la destination de sauvegarde annulée par l'utilisateur pour %s.", name)
		fmt.Println("Suppression annulée.")
		return
	}

	if err := common.DeleteBackupDestination(name); err != nil {
		input.DisplayMessage(true, "Erreur lors de la suppression de la destination de sauvegarde %s: %v", name, err)
		return
	}

	input.DisplayMessage(false, "Destination '%s' supprimée avec succès.", name)
}

// setDefaultBackupDestination permet de définir une destination de sauvegarde par défaut
func setDefaultBackupDestination() {
	common.LogInfo("Début de la définition de la destination de sauvegarde par défaut.")
	if len(common.AppConfig.BackupDestinations) == 0 {
		input.DisplayMessage(false, "Aucune destination de sauvegarde à définir par défaut.")
		return
	}

	idxStr := input.ReadInput("Numéro de la destination à définir par défaut: ")
	idx, err := strconv.Atoi(idxStr)

	if err != nil || idx < 1 || idx > len(common.AppConfig.BackupDestinations) {
		input.DisplayMessage(true, "Numéro invalide.")
		return
	}

	name := common.AppConfig.BackupDestinations[idx-1].Name

	if err := common.SetDefaultBackupDestination(name); err != nil {
		input.DisplayMessage(true, "Erreur lors de la définition de la destination par défaut %s: %v", name, err)
		return
	}

	input.DisplayMessage(false, "Destination par défaut définie sur '%s'.", name)
}

// changeBackupDestination permet de modifier la destination principale de sauvegarde
func changeBackupDestination() {
	common.LogInfo("Début de la modification de la destination principale de sauvegarde.")
	display.ClearScreen()
	fmt.Printf("%sModification de la Destination Principale de Sauvegarde%s\n\n", display.ColorBold(), display.ColorReset())

	fmt.Printf("Destination actuelle: %s\n\n", common.AppConfig.BackupDestination)

	// Demander la nouvelle destination
	newDest := input.ReadStringInput("Nouvelle destination (vide pour annuler): ", common.AppConfig.BackupDestination, common.IsValidPath, "Chemin invalide.")

	if newDest == common.AppConfig.BackupDestination {
		common.LogInfo("Modification de la destination annulée: pas de changement.")
		fmt.Println("Modification annulée.")
		return
	}

	common.AppConfig.BackupDestination = newDest

	if err := common.SaveConfig(common.AppConfig); err != nil {
		input.DisplayMessage(true, "Erreur lors de la sauvegarde de la nouvelle destination principale: %v", err)
		return
	}

	input.DisplayMessage(false, "Destination principale modifiée avec succès: %s", newDest)
}