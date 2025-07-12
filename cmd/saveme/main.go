package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"github.com/Noziop/s4v3my4ss/internal/ui"
	"github.com/Noziop/s4v3my4ss/internal/ui/commands"
	"github.com/Noziop/s4v3my4ss/pkg/common"
)

// Version de l'application
const Version = "1.0.0"

// Nom de la commande
const CommandName = "saveme"

// Variable globale pour suivre si une sauvegarde est en cours
var backupInProgress bool

func main() {
	// Configurer la gestion des signaux pour une interruption propre
	setupSignalHandling()

	// Initialisation de l'application
	if err := common.InitApp(); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur d'initialisation: %v\n", err)
		os.Exit(1)
	}

	// Si aucun argument n'est fourni, lancer le mode interactif
	if len(os.Args) < 2 {
		ui.RunInteractiveMode()
		return
	}

	// Sinon, traiter les arguments de ligne de commande
	switch os.Args[1] {
	case "watch", "--watch", "-w":
		commands.HandleWatchCommand(os.Args[2:])
	case "restore", "--restore", "-r":
		commands.HandleRestoreCommand(os.Args[2:])
	case "manage", "--manage", "-m":
		ui.HandleManageCommand(os.Args[2:])
	case "discover":
		ui.HandleDiscoverCommand(os.Args[2:])
	case "add":
		commands.HandleAddCommand(os.Args[2:])
	case "--help", "-h":
		printHelp()
	case "--version", "-v":
		fmt.Printf("S4v3my4ss version %s\n", Version)
	default:
		fmt.Fprintln(os.Stderr, "Option non reconnue:", os.Args[1])
		fmt.Fprintln(os.Stderr, "Utilisez --help pour voir les options disponibles.")
		os.Exit(1)
	}
}

// setupSignalHandling configure la gestion des signaux d'interruption (Ctrl+C)
func setupSignalHandling() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c // Attendre le signal
		fmt.Println("\nInterruption détectée...")

		if backupInProgress {
			fmt.Println("Nettoyage des fichiers temporaires en cours...")
			// TODO: Ajouter ici la logique de nettoyage spécifique si nécessaire
			fmt.Println("Les métadonnées de la sauvegarde interrompue seront conservées.")
			fmt.Println("Vous pourrez les supprimer manuellement via le menu de gestion des sauvegardes.")
		}

		fmt.Println("Au revoir!")
		os.Exit(0)
	}()
}



// Affiche l'aide de l'application
func printHelp() {
	bold := "\033[1m"
	nc := "\033[0m" // No Color

	fmt.Printf("%sUsage:%s %s [commande] [options]\n\n", bold, nc, CommandName)
	fmt.Printf("%sCommandes:%s\n", bold, nc)
	fmt.Println("  watch     Surveiller un répertoire et créer des sauvegardes")
	fmt.Println("  restore   Restaurer une sauvegarde existante")
	fmt.Println("  manage    Gérer les sauvegardes existantes")
	fmt.Println("  discover  Découvrir les serveurs rsync sur le réseau")
	fmt.Println("  add       Ajouter une nouvelle configuration (ex: add server)")
	fmt.Println("  --help    Afficher cette aide")
	fmt.Println("  --version Afficher la version")
	fmt.Println()
	fmt.Println("Exécutez sans arguments pour lancer l'interface interactive.")
}