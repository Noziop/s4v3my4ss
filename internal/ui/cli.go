package ui

import (
	"fmt"
	"os"

	"github.com/Noziop/s4v3my4ss/pkg/common"
	"github.com/Noziop/s4v3my4ss/internal/setup"
	"github.com/Noziop/s4v3my4ss/internal/ui/backup"
	"github.com/Noziop/s4v3my4ss/internal/ui/commands"
	"github.com/Noziop/s4v3my4ss/internal/ui/config"
	"github.com/Noziop/s4v3my4ss/internal/ui/discovery"
	"github.com/Noziop/s4v3my4ss/internal/ui/display"
	"github.com/Noziop/s4v3my4ss/internal/ui/input"
)



// RunInteractiveMode démarre le mode interactif de l'application
func RunInteractiveMode() {
	common.LogInfo("Démarrage du mode interactif.")
	display.DisplayHeader()
	fmt.Printf("\033[1mSystème de Sauvegarde et Restauration Automatique\033[0m\n\n")
	// SECURITY: Vérifier si l'application est exécutée en tant que root.
	if os.Geteuid() == 0 {
		common.LogSecurity("AVERTISSEMENT: Application exécutée en tant que root.")
		input.DisplayMessage(true, "AVERTISSEMENT: Vous exécutez cette application en tant que root. Il est fortement recommandé d'utiliser un utilisateur avec des privilèges moindres pour des raisons de sécurité.")
	}

	for {
		display.ClearScreen()
		display.DisplayMainMenu()

		choice := input.ReadInput("Votre choix: ")

		switch choice {
		case "1":
			commands.ConfigureBackup()
		case "2":
			commands.WatchDirectoryInteractive()
		case "3":
			commands.RestoreBackupInteractive(false)
		case "4":
			backup.ManageBackupsInteractive()
		case "5":
			setup.CheckDependencies()
		case "6":	
			discovery.DiscoverRsyncServers()
		case "7":
			config.ManageConfiguration()
		case "0":
			common.LogInfo("Quitter l'application.")
			fmt.Println("Au revoir !")
			return
		default:
			common.LogWarning("Option de menu non valide: %s", choice)
			fmt.Println("Option non valide. Veuillez réessayer.")
		}

		fmt.Println()
		input.ReadInput("Appuyez sur Entrée pour continuer...")
	}
}

// HandleManageCommand traite la commande 'manage' depuis la ligne de commande
func HandleManageCommand(args []string) {
	common.LogInfo("Traitement de la commande 'manage' avec les arguments: %v", args)
	if len(args) == 0 {
		common.LogInfo("Aucun argument fourni pour manage. Lancement du mode interactif.")
		backup.ManageBackupsInteractive()
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "list":
		common.LogInfo("Exécution de la sous-commande manage list.")
		backup.ListBackups()
	case "delete":
		common.LogInfo("Exécution de la sous-commande manage delete.")
		if len(args) < 2 {
			common.LogError("Utilisation incorrecte de manage delete: ID de sauvegarde manquant.")
			fmt.Fprintln(os.Stderr, "Usage: " + common.CommandName + " manage delete <backup_id>")
			os.Exit(1)
		}
		backupID := args[1]
		if !common.IsValidName(backupID) { // Utilisation de common.IsValidName
			input.DisplayMessage(true, "ID de sauvegarde invalide: %s", backupID)
			os.Exit(1)
	}
		backup.DeleteBackup(backupID)
	case "clean":
		common.LogInfo("Exécution de la sous-commande manage clean.")
		//cleanOldBackups()
	default:
		common.LogWarning("Sous-commande manage inconnue: %s", subcommand)
		fmt.Fprintln(os.Stderr, "Sous-commande inconnue:", subcommand)
		fmt.Fprintln(os.Stderr, "Sous-commandes disponibles: list, delete, clean")
		os.Exit(1)
	}
}

// HandleDiscoverCommand traite la commande 'discover' depuis la ligne de commande
func HandleDiscoverCommand(args []string) {
	common.LogInfo("Traitement de la commande 'discover' avec les arguments: %v", args)
	subnetCIDR := ""
	if len(args) > 0 {
		subnetCIDR = args[0]
		if !common.IsValidSubnet(subnetCIDR) {
			input.DisplayMessage(true, "Format CIDR invalide (ex: 192.168.0.0/24): %s", subnetCIDR)
			os.Exit(1)
		}
	} else {
		// TODO: Essayer de déterminer le sous-réseau local automatiquement
		subnetCIDR = "192.168.0.0/24"
		common.LogInfo("Aucun sous-réseau fourni, utilisation de la valeur par défaut: %s", subnetCIDR)
	}

	discovery.DiscoverRsyncServers()
}

// HandleAddCommand traite la commande 'add' et ses sous-commandes.
func HandleAddCommand(args []string) {
	common.LogInfo("Traitement de la commande 'add' avec les arguments: %v", args)
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Erreur: La commande 'add' nécessite une sous-commande (ex: server).")
		os.Exit(1)
	}

	switch args[0] {
	case "server":
		commands.HandleAddServerCommand(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Erreur: Sous-commande '%s' non reconnue pour 'add'.\n", args[0])
		os.Exit(1)
	}
}
