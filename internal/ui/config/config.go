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
			return // Suppression de input.ReadInput ici
		default:
			common.LogWarning("Option de gestion de configuration non valide: %s", choice)
			fmt.Println("Option non valide. Veuillez réessayer.")
		}

		input.ReadInput("Appuyez sur Entrée pour continuer...")
	}
}

// displayFullConfig affiche la configuration complète de l'application
func displayFullConfig() {
	common.LogInfo("Affichage de la configuration complète.")
	display.ClearScreen()
	fmt.Printf("%sConfiguration Complète de l'Application%s\n\n", display.ColorBold, display.ColorReset)

	// Afficher les informations générales
	fmt.Printf("%sInformations générales:%s\n", display.ColorBold, display.ColorReset)
	fmt.Printf("Destination des sauvegardes: %s\n", common.AppConfig.BackupDestination)
	fmt.Printf("Dernière mise à jour: %s\n\n", common.AppConfig.LastUpdate.Format("02/01/2006 15:04:05"))

	// Afficher la politique de rétention
	fmt.Printf("%sPolitique de rétention:%s\n", display.ColorBold, display.ColorReset)
	fmt.Printf("Conservation quotidienne: %d jours\n", common.AppConfig.RetentionPolicy.KeepDaily)
	fmt.Printf("Conservation hebdomadaire: %d semaines\n", common.AppConfig.RetentionPolicy.KeepWeekly)
	fmt.Printf("Conservation mensuelle: %d mois\n\n", common.AppConfig.RetentionPolicy.KeepMonthly)

	// Afficher les répertoires sauvegardés
	fmt.Printf("%sRépertoires sauvegardés:%s\n", display.ColorBold, display.ColorReset)
	if len(common.AppConfig.BackupDirs) == 0 {
		common.LogInfo("Aucun répertoire de sauvegarde configuré à afficher.")
		fmt.Printf("%sAucun répertoire configuré.%s\n\n", display.ColorYellow, display.ColorReset)
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
	fmt.Printf("%sServeurs rsync configurés:%s\n", display.ColorBold, display.ColorReset)
	if len(common.AppConfig.RsyncServers) == 0 {
		common.LogInfo("Aucun serveur rsync configuré à afficher.")
		fmt.Printf("%sAucun serveur rsync configuré.%s\n\n", display.ColorYellow(), display.ColorReset())
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
		display.ClearScreen()
		fmt.Printf("%sGestion des répertoires sauvegardés%s\n\n", display.ColorBold, display.ColorReset)

		// Afficher les répertoires sauvegardés
		if len(common.AppConfig.BackupDirs) == 0 {
			common.LogWarning("Aucun répertoire de sauvegarde configuré à gérer.")
			fmt.Printf("%sAucun répertoire configuré.%s\n\n", display.ColorYellow, display.ColorReset)
		} else {
			common.LogInfo("Affichage des %d répertoires de sauvegarde configurés pour gestion.", len(common.AppConfig.BackupDirs))
			fmt.Printf("%sRépertoires configurés:%s\n", display.ColorBold, display.ColorReset)
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
		case "3":
			deleteBackupDirectory()
		case "0":
			common.LogInfo("Retour au menu de gestion de la configuration.")
			return // Suppression de input.ReadInput ici
		default:
			common.LogWarning("Option de gestion des répertoires de sauvegarde non valide: %s", choice)
			fmt.Println("Option non valide.")
		}

		input.ReadInput("Appuyez sur Entrée pour continuer...")
	}
}

// editBackupDirectory permet de modifier un répertoire de sauvegarde existant
func editBackupDirectory() {
	common.LogInfo("Début de la modification d'un répertoire de sauvegarde.")
	if len(common.AppConfig.BackupDirs) == 0 {
		common.LogWarning("Aucun répertoire de sauvegarde à modifier.")
		fmt.Printf("%sAucun répertoire à modifier.%s\n", display.ColorYellow, display.ColorReset)
		return
	}

	idxStr := input.ReadInput("Numéro du répertoire à modifier: ")
	idx, err := strconv.Atoi(idxStr)

	if err != nil || idx < 1 || idx > len(common.AppConfig.BackupDirs) {
		common.LogError("Numéro de répertoire invalide pour modification: %s", idxStr)
		fmt.Printf("%sNuméro invalide.%s\n", display.ColorRed, display.ColorReset)
		return
	}

	// Récupérer la configuration à modifier
	dir := common.AppConfig.BackupDirs[idx-1]
	common.LogInfo("Modification du répertoire de sauvegarde: %s.", dir.Name)
	fmt.Printf("%sModification de la configuration '%s'%s\n\n", display.ColorBold, dir.Name, display.ColorReset)

	// Permettre de modifier chaque propriété
	fmt.Printf("Nom actuel: %s\n", dir.Name)
	name := input.ReadAndValidateInput("Nouveau nom (vide pour garder l'actuel): ", common.IsValidName, "Nom invalide.") // Utilisation de common.IsValidName
	if name == "" {
		name = dir.Name
	}

	fmt.Printf("Chemin actuel: %s\n", dir.SourcePath)
	sourcePath := input.ReadAndValidateInput("Nouveau chemin (vide pour garder l'actuel): ", common.IsValidPath, "Chemin invalide.") // Utilisation de common.IsValidPath
	if sourcePath == "" {
		sourcePath = dir.SourcePath
	}

	// Option pour une sauvegarde incrémentale
	incStr := "n"
	if dir.IsIncremental {
		incStr = "o"
	}
	fmt.Printf("Sauvegarde incrémentale actuelle: %s\n", incStr)
	incrementalStr := input.ReadInput("Activer les sauvegardes incrémentales? (o/n, vide pour garder l'actuel): ")
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
	compressStr := input.ReadInput("Activer la compression? (o/n, vide pour garder l'actuel): ")
	compression := dir.Compression
	if compressStr != "" {
		compression = strings.ToLower(compressStr) == "o"
	}

	// Répertoires à exclure
	fmt.Printf("Répertoires exclus actuels: %s\n", strings.Join(dir.ExcludeDirs, ", "))
	excludeDirsStr := input.ReadInput("Nouveaux répertoires à exclure (séparés par des virgules, vide pour garder les actuels): ")
	excludeDirs := dir.ExcludeDirs
	if excludeDirsStr != "" {
		excludeDirs = strings.Split(excludeDirsStr, ",")
		for i, dir := range excludeDirs {
			excludeDirs[i] = strings.TrimSpace(dir)
		}
	}

	// Fichiers à exclure
	fmt.Printf("Fichiers exclus actuels: %s\n", strings.Join(dir.ExcludeFiles, ", "))
	excludeFilesStr := input.ReadInput("Nouveaux fichiers à exclure (séparés par des virgules, vide pour garder les actuels): ")
	excludeFiles := dir.ExcludeFiles
	if excludeFilesStr != "" {
		excludeFiles = strings.Split(excludeFilesStr, ",")
		for i, file := range excludeFiles {
			excludeFiles[i] = strings.TrimSpace(file)
		}
	}

	// Intervalle
	fmt.Printf("Intervalle actuel: %d minutes\n", dir.Interval)
	intervalStr := input.ReadInput("Nouvel intervalle en minutes (vide pour garder l'actuel): ")
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
		fmt.Printf("%sErreur lors de la mise à jour de la configuration: %v%s\n", display.ColorRed, err, display.ColorReset)
		return
	}

	common.LogInfo("Configuration du répertoire %s modifiée avec succès.", name)
	fmt.Printf("%sConfiguration '%s' modifiée avec succès.%s\n", display.ColorGreen, name, display.ColorReset)
}

// deleteBackupDirectory permet de supprimer un répertoire de sauvegarde
func deleteBackupDirectory() {
	common.LogInfo("Début de la suppression d'un répertoire de sauvegarde.")
	if len(common.AppConfig.BackupDirs) == 0 {
		common.LogWarning("Aucun répertoire de sauvegarde à supprimer.")
		fmt.Printf("%sAucun répertoire à supprimer.%s\n", display.ColorYellow, display.ColorReset)
		return
	}

	idxStr := input.ReadInput("Numéro du répertoire à supprimer: ")
	idx, err := strconv.Atoi(idxStr)

	if err != nil || idx < 1 || idx > len(common.AppConfig.BackupDirs) {
		common.LogError("Numéro de répertoire invalide pour suppression: %s", idxStr)
		fmt.Printf("%sNuméro invalide.%s\n", display.ColorRed, display.ColorReset)
		return
	}

	// Récupérer le nom pour confirmation
	name := common.AppConfig.BackupDirs[idx-1].Name

	confirm := input.ReadInput(fmt.Sprintf("Êtes-vous sûr de vouloir supprimer la configuration '%s'? (o/n): ", name))
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
		fmt.Printf("%sErreur lors de la mise à jour de la configuration: %v%s\n", display.ColorRed, err, display.ColorReset)
		return
	}

	common.LogInfo("Configuration du répertoire %s supprimée avec succès.", name)
	fmt.Printf("%sConfiguration '%s' supprimée avec succès.%s\n", display.ColorGreen, name, display.ColorReset)
}

// manageRsyncServers permet de gérer les serveurs rsync configurés
func manageRsyncServers() {
	common.LogInfo("Début de la gestion des serveurs rsync.")
	for {
		display.ClearScreen()
		fmt.Printf("%sGestion des serveurs rsync%s\n\n", display.ColorBold, display.ColorReset)

		// Afficher les serveurs configurés
		if len(common.AppConfig.RsyncServers) == 0 {
			common.LogWarning("Aucun serveur rsync configuré à gérer.")
			fmt.Printf("%sAucun serveur rsync configuré.%s\n\n", display.ColorYellow(), display.ColorReset())
		} else {
			common.LogInfo("Affichage des %d serveurs rsync configurés pour gestion.", len(common.AppConfig.RsyncServers))
			fmt.Printf("%sServeurs configurés:%s\n", display.ColorBold, display.ColorReset)
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

		fmt.Printf("  %s1.%s Rechercher et ajouter un serveur\n", display.ColorGreen, display.ColorReset)
		fmt.Printf("  %s2.%s Modifier un serveur\n", display.ColorGreen, display.ColorReset)
		fmt.Printf("  %s3.%s Supprimer un serveur\n", display.ColorGreen, display.ColorReset)
		fmt.Printf("  %s0.%s Retour\n\n", display.ColorGreen, display.ColorReset)

		choice := input.ReadInput("Votre choix: ")

		switch choice {
		case "1":
			//discoverRsyncServers()
			common.LogInfo("Recherche et ajout de serveur rsync demandé.")
			return
		case "2":
			editRsyncServer()
		case "3":
			deleteRsyncServer()
		case "0":
			common.LogInfo("Retour au menu de gestion de la configuration.")
			return // Suppression de input.ReadInput ici
		default:
			common.LogWarning("Option de gestion des serveurs rsync non valide: %s", choice)
			fmt.Println("Option non valide. Veuillez réessayer.")
		}

		input.ReadInput("Appuyez sur Entrée pour continuer...")
	}
}

// editRsyncServer permet de modifier un serveur rsync existant
func editRsyncServer() {
	common.LogInfo("Début de la modification d'un serveur rsync.")
	if len(common.AppConfig.RsyncServers) == 0 {
		common.LogWarning("Aucun serveur rsync à modifier.")
		fmt.Printf("%sAucun serveur à modifier.%s\n", display.ColorYellow, display.ColorReset)
		return
	}

	idxStr := input.ReadInput("Numéro du serveur à modifier: ")
	idx, err := strconv.Atoi(idxStr)

	if err != nil || idx < 1 || idx > len(common.AppConfig.RsyncServers) {
		common.LogError("Numéro de serveur rsync invalide pour modification: %s", idxStr)
		fmt.Printf("%sNuméro invalide.%s\n", display.ColorRed, display.ColorReset)
		return
	}

	// Récupérer la configuration à modifier
	server := common.AppConfig.RsyncServers[idx-1]
	common.LogInfo("Modification du serveur rsync: %s.", server.Name)
	fmt.Printf("%sModification du serveur '%s'%s\n\n", display.ColorBold, server.Name, display.ColorReset)

	// Permettre de modifier chaque propriété
	fmt.Printf("Nom actuel: %s\n", server.Name)
	name := input.ReadAndValidateInput("Nouveau nom (vide pour garder l'actuel): ", common.IsValidName, "Nom invalide.") // Utilisation de common.IsValidName
	if name == "" {
		name = server.Name
	}

	fmt.Printf("Adresse IP actuelle: %s\n", server.IP)
	ip := input.ReadInput("Nouvelle adresse IP (vide pour garder l'actuelle): ")
	if ip == "" {
		ip = server.IP
	}

	fmt.Printf("Nom d'utilisateur actuel: %s\n", server.Username)
	username := input.ReadAndValidateInput("Nouveau nom d'utilisateur (vide pour garder l'actuel): ", common.IsValidName, "Nom d'utilisateur invalide.") // Utilisation de common.IsValidName
	if username == "" {
		username = server.Username
	}

	fmt.Printf("Port SSH actuel: %d\n", server.SSHPort)
	sshPortStr := input.ReadInput("Nouveau port SSH (vide pour garder l'actuel): ")
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
		common.LogError("Erreur lors de la mise à jour du serveur rsync %s: %v", name, err)
		fmt.Printf("%sErreur lors de la mise à jour de la configuration: %v%s\n", display.ColorRed, err, display.ColorReset)
		return
	}

	common.LogInfo("Serveur rsync %s modifié avec succès.", name)
	fmt.Printf("%sServeur '%s' modifié avec succès.%s\n", display.ColorGreen, name, display.ColorReset)
}

// deleteRsyncServer permet de supprimer un serveur rsync
func deleteRsyncServer() {
	common.LogInfo("Début de la suppression d'un serveur rsync.")
	if len(common.AppConfig.RsyncServers) == 0 {
		common.LogWarning("Aucun serveur rsync à supprimer.")
		fmt.Printf("%sAucun serveur à supprimer.%s\n", display.ColorYellow, display.ColorReset)
		return
	}

	idxStr := input.ReadInput("Numéro du serveur à supprimer: ")
	idx, err := strconv.Atoi(idxStr)

	if err != nil || idx < 1 || idx > len(common.AppConfig.RsyncServers) {
		common.LogError("Numéro de serveur rsync invalide pour suppression: %s", idxStr)
		fmt.Printf("%sNuméro invalide.%s\n", display.ColorRed, display.ColorReset)
		return
	}

	// Récupérer le nom pour confirmation
	name := common.AppConfig.RsyncServers[idx-1].Name

	confirm := input.ReadInput(fmt.Sprintf("Êtes-vous sûr de vouloir supprimer le serveur '%s'? (o/n): ", name))
	if strings.ToLower(confirm) != "o" {
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
		common.LogError("Erreur lors de la suppression du serveur rsync %s: %v", name, err)
		fmt.Printf("%sErreur lors de la mise à jour de la configuration: %v%s\n", display.ColorRed, err, display.ColorReset)
		return
	}

	common.LogInfo("Serveur rsync %s supprimé avec succès.", name)
	fmt.Printf("%sServeur '%s' supprimé avec succès.%s\n", display.ColorGreen, name, display.ColorReset)
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
	dailyStr := input.ReadInput(fmt.Sprintf("Nouvelle conservation quotidienne (actuel: %d): ", common.AppConfig.RetentionPolicy.KeepDaily))
	weeklyStr := input.ReadInput(fmt.Sprintf("Nouvelle conservation hebdomadaire (actuel: %d): ", common.AppConfig.RetentionPolicy.KeepWeekly))
	monthlyStr := input.ReadInput(fmt.Sprintf("Nouvelle conservation mensuelle (actuel: %d): ", common.AppConfig.RetentionPolicy.KeepMonthly))

	// Mettre à jour les valeurs si elles sont valides
	if daily, err := strconv.Atoi(dailyStr); err == nil && daily >= 0 {
		common.AppConfig.RetentionPolicy.KeepDaily = daily
		common.LogInfo("Politique de rétention quotidienne mise à jour: %d", daily)
	} else if dailyStr != "" {
		common.LogWarning("Valeur de rétention quotidienne invalide: %s", dailyStr)
	}

	if weekly, err := strconv.Atoi(weeklyStr); err == nil && weekly >= 0 {
		common.AppConfig.RetentionPolicy.KeepWeekly = weekly
		common.LogInfo("Politique de rétention hebdomadaire mise à jour: %d", weekly)
	} else if weeklyStr != "" {
		common.LogWarning("Valeur de rétention hebdomadaire invalide: %s", weeklyStr)
	}

	if monthly, err := strconv.Atoi(monthlyStr); err == nil && monthly >= 0 {
		common.AppConfig.RetentionPolicy.KeepMonthly = monthly
		common.LogInfo("Politique de rétention mensuelle mise à jour: %d", monthly)
	} else if monthlyStr != "" {
		common.LogWarning("Valeur de rétention mensuelle invalide: %s", monthlyStr)
	}

	// Sauvegarder la configuration
	if err := common.SaveConfig(common.AppConfig); err != nil {
		common.LogError("Erreur lors de la sauvegarde de la politique de rétention: %v", err)
		fmt.Printf("%sErreur lors de la sauvegarde de la configuration: %v%s\n", display.ColorRed, err, display.ColorReset)
		return
	}

	common.LogInfo("Politique de rétention mise à jour avec succès.")
	fmt.Printf("\n%sPolitique de rétention mise à jour avec succès.%s\n", display.ColorGreen(), display.ColorReset())
}

// manageBackupDestinations permet de gérer les destinations de sauvegarde
func manageBackupDestinations() {
	common.LogInfo("Début de la gestion des destinations de sauvegarde.")
	for {
		display.ClearScreen()
		fmt.Printf("%sGestion des Destinations de Sauvegarde%s\n\n", display.ColorBold, display.ColorReset)

		// Afficher les destinations existantes
		if len(common.AppConfig.BackupDestinations) == 0 {
			common.LogWarning("Aucune destination de sauvegarde configurée à gérer.")
			fmt.Printf("%sAucune destination configurée.%s\n\n", display.ColorYellow, display.ColorReset)
		} else {
			common.LogInfo("Affichage des %d destinations de sauvegarde configurées pour gestion.", len(common.AppConfig.BackupDestinations))
			fmt.Println("Destinations configurées:")
			for i, dest := range common.AppConfig.BackupDestinations {
				defaultStr := ""
				if dest.IsDefault {
					defaultStr = " (Par défaut)"
				}
				fmt.Printf("%d. %s (%s) - %s%s\n", i+1, dest.Name, dest.Type, dest.Path, defaultStr)
			}
			fmt.Println()
		}

		fmt.Printf("  %s1.%s Ajouter une destination\n", display.ColorGreen, display.ColorReset)
		fmt.Printf("  %s2.%s Modifier une destination\n", display.ColorGreen, display.ColorReset)
		fmt.Printf("  %s3.%s Supprimer une destination\n", display.ColorGreen, display.ColorReset)
		fmt.Printf("  %s4.%s Définir une destination par défaut\n", display.ColorGreen, display.ColorReset)
		fmt.Printf("  %s0.%s Retour\n\n", display.ColorGreen, display.ColorReset)

		choice := input.ReadInput("Votre choix: ")

		switch choice {
		case "1":
			addBackupDestination()
		case "2":
			editBackupDestination()
		case "3":
			deleteBackupDestination()
		case "4":
			setDefaultBackupDestination()
		case "0":
			common.LogInfo("Retour au menu de gestion de la configuration.")
			return
		default:
			common.LogWarning("Option de gestion des destinations de sauvegarde non valide: %s", choice)
			fmt.Println("Option non valide.")
		}

		input.ReadInput("Appuyez sur Entrée pour continuer...")
	}
}

// addBackupDestination permet d'ajouter une nouvelle destination de sauvegarde
func addBackupDestination() {
	common.LogInfo("Ajout d'une nouvelle destination de sauvegarde.")
	fmt.Printf("\n%sAjout d'une nouvelle destination%s\n\n", display.ColorBold, display.ColorReset)

	name := input.ReadAndValidateInput("Nom de la destination: ", common.IsValidName, "Nom invalide.")
	path := input.ReadAndValidateInput("Chemin de la destination: ", common.IsValidPath, "Chemin invalide.")

	destType := "local"
	if strings.HasPrefix(path, "rsync://") {
		destType = "rsync"
	}

	defaultStr := input.ReadInput("Définir comme destination par défaut? (o/n): ")
	isDefault := strings.ToLower(defaultStr) == "o"

	newDest := common.BackupDestination{
		Name:      name,
		Path:      path,
		Type:      destType,
		IsDefault: isDefault,
	}

	if err := common.AddBackupDestination(newDest); err != nil{
		common.LogError("Erreur lors de l'ajout de la destination de sauvegarde %s: %v", name, err)
		fmt.Printf("%sErreur lors de l'ajout de la destination: %v%s\n", display.ColorRed(), err, display.ColorReset())
		return
	}

	common.LogInfo("Destination de sauvegarde %s ajoutée avec succès.", name)
	fmt.Printf("\n%sDestination '%s' ajoutée avec succès.%s\n", display.ColorGreen, name, display.ColorReset)
}

// editBackupDestination permet de modifier une destination de sauvegarde existante
func editBackupDestination() {
	common.LogInfo("Début de la modification d'une destination de sauvegarde.")
	if len(common.AppConfig.BackupDestinations) == 0 {
		common.LogWarning("Aucune destination de sauvegarde à modifier.")
		fmt.Printf("%sAucune destination à modifier.%s\n", display.ColorYellow, display.ColorReset)
		return
	}

	idxStr := input.ReadInput("Numéro de la destination à modifier: ")
	idx, err := strconv.Atoi(idxStr)

	if err != nil || idx < 1 || idx > len(common.AppConfig.BackupDestinations) {
		common.LogError("Numéro de destination invalide pour modification: %s", idxStr)
		fmt.Printf("%sNuméro invalide.%s\n", display.ColorRed, display.ColorReset)
		return
	}

	dest := common.AppConfig.BackupDestinations[idx-1]
	common.LogInfo("Modification de la destination de sauvegarde: %s.", dest.Name)
	fmt.Printf("\n%sModification de la destination '%s'%s\n\n", display.ColorBold, dest.Name, display.ColorReset)

	name := input.ReadAndValidateInput(fmt.Sprintf("Nouveau nom (actuel: %s): ", dest.Name), common.IsValidName, "Nom invalide.")
	if name == "" {
		name = dest.Name
	}

	path := input.ReadAndValidateInput(fmt.Sprintf("Nouveau chemin (actuel: %s): ", dest.Path), common.IsValidPath, "Chemin invalide.")
	if path == "" {
		path = dest.Path
	}

	destType := "local"
	if strings.HasPrefix(path, "rsync://") {
		destType = "rsync"
	}

	defaultStr := input.ReadInput(fmt.Sprintf("Définir comme destination par défaut? (actuel: %t, o/n): ", dest.IsDefault))
	isDefault := dest.IsDefault
	if defaultStr != "" {
		isDefault = strings.ToLower(defaultStr) == "o"
	}

	updatedDest := common.BackupDestination{
		Name:      name,
		Path:      path,
		Type:      destType,
		IsDefault: isDefault,
	}

	if err := common.UpdateBackupDestination(dest.Name, updatedDest); err != nil {
		common.LogError("Erreur lors de la mise à jour de la destination de sauvegarde %s: %v", name, err)
		fmt.Printf("%sErreur lors de la mise à jour de la destination: %v%s\n", display.ColorRed, err, display.ColorReset)
		return
	}

	common.LogInfo("Destination de sauvegarde %s mise à jour avec succès.", name)
	fmt.Printf("\n%sDestination '%s' mise à jour avec succès.%s\n", display.ColorGreen, name, display.ColorReset)
}

// deleteBackupDestination permet de supprimer une destination de sauvegarde
func deleteBackupDestination() {
	common.LogInfo("Début de la suppression d'une destination de sauvegarde.")
	if len(common.AppConfig.BackupDestinations) == 0 {
		common.LogWarning("Aucune destination de sauvegarde à supprimer.")
		fmt.Printf("%sAucune destination à supprimer.%s\n", display.ColorYellow, display.ColorReset)
		return
	}

	idxStr := input.ReadInput("Numéro de la destination à supprimer: ")
	idx, err := strconv.Atoi(idxStr)

	if err != nil || idx < 1 || idx > len(common.AppConfig.BackupDestinations) {
		common.LogError("Numéro de destination invalide pour suppression: %s", idxStr)
		fmt.Printf("%sNuméro invalide.%s\n", display.ColorRed, display.ColorReset)
		return
	}

	name := common.AppConfig.BackupDestinations[idx-1].Name

	confirm := input.ReadInput(fmt.Sprintf("Êtes-vous sûr de vouloir supprimer la destination '%s'? (o/n): ", name))
	if strings.ToLower(confirm) != "o" {
		common.LogInfo("Suppression de la destination de sauvegarde annulée par l'utilisateur pour %s.", name)
		fmt.Println("Suppression annulée.")
		return
	}

	if err := common.DeleteBackupDestination(name); err != nil {
		common.LogError("Erreur lors de la suppression de la destination de sauvegarde %s: %v", name, err)
		fmt.Printf("%sErreur lors de la suppression de la destination: %v%s\n", display.ColorRed, err, display.ColorReset)
		return
	}

	common.LogInfo("Destination de sauvegarde %s supprimée avec succès.", name)
	fmt.Printf("\n%sDestination '%s' supprimée avec succès.%s\n", display.ColorGreen(), name, display.ColorReset())
}

// setDefaultBackupDestination permet de définir une destination de sauvegarde par défaut
func setDefaultBackupDestination() {
	common.LogInfo("Début de la définition de la destination de sauvegarde par défaut.")
	if len(common.AppConfig.BackupDestinations) == 0 {
		common.LogWarning("Aucune destination de sauvegarde à définir par défaut.")
		fmt.Printf("%sAucune destination à définir par défaut.%s\n", display.ColorYellow, display.ColorReset)
		return
	}

	idxStr := input.ReadInput("Numéro de la destination à définir par défaut: ")
	idx, err := strconv.Atoi(idxStr)

	if err != nil || idx < 1 || idx > len(common.AppConfig.BackupDestinations) {
		common.LogError("Numéro de destination invalide pour la définition par défaut: %s", idxStr)
		fmt.Printf("%sNuméro invalide.%s\n", display.ColorRed, display.ColorReset)
		return
	}

	name := common.AppConfig.BackupDestinations[idx-1].Name

	if err := common.SetDefaultBackupDestination(name); err != nil {
		common.LogError("Erreur lors de la définition de la destination par défaut %s: %v", name, err)
		fmt.Printf("%sErreur lors de la définition de la destination par défaut: %v%s\n", display.ColorRed, err, display.ColorReset)
		return
	}

	common.LogInfo("Destination de sauvegarde par défaut définie sur %s.", name)
	fmt.Printf("\n%sDestination par défaut définie sur '%s'.%s\n", display.ColorGreen, name, display.ColorReset)
}

// changeBackupDestination permet de modifier la destination principale de sauvegarde
func changeBackupDestination() {
	common.LogInfo("Début de la modification de la destination principale de sauvegarde.")
	display.ClearScreen()
	fmt.Printf("%sModification de la Destination Principale de Sauvegarde%s\n\n", display.ColorBold, display.ColorReset)

	fmt.Printf("Destination actuelle: %s\n\n", common.AppConfig.BackupDestination)

	// Demander la nouvelle destination
	newDest := input.ReadAndValidateInput("Nouvelle destination (vide pour annuler): ", common.IsValidPath, "Chemin invalide.") // Utilisation de common.IsValidPath
	if newDest == "" {
		common.LogInfo("Modification de la destination annulée: chemin vide.")
		fmt.Println("Modification annulée.")
		return
	}

	common.AppConfig.BackupDestination = newDest

	if err := common.SaveConfig(common.AppConfig); err != nil {
		common.LogError("Erreur lors de la sauvegarde de la nouvelle destination principale: %v", err)
		fmt.Printf("%sErreur lors de la sauvegarde de la configuration: %v%s\n", display.ColorRed, err, display.ColorReset)
		return
	}

	common.LogInfo("Destination principale de sauvegarde modifiée: %s", newDest)
	fmt.Printf("\n%sDestination principale modifiée avec succès.%s\n", display.ColorGreen(), newDest, display.ColorReset())
}