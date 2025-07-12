package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/Noziop/s4v3my4ss/pkg/common"
)

// HandleAddCommand traite la commande 'add' et ses sous-commandes.
func HandleAddCommand(args []string) {
	common.LogInfo("Traitement de la commande 'add' avec les arguments: %v", args)
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Erreur: La commande 'add' nécessite une sous-commande (ex: server).")
		os.Exit(1)
	}

	switch args[0] {
	case "server":
		HandleAddServerCommand(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Erreur: Sous-commande '%s' non reconnue pour 'add'.\n", args[0])
		os.Exit(1)
	}
}

// HandleAddServerCommand gère la logique pour 'add server'.
func HandleAddServerCommand(args []string) {
	addServerCmd := flag.NewFlagSet("add server", flag.ExitOnError)

	// Définition des flags
	name := addServerCmd.String("name", "", "(Requis) Nom unique pour le serveur.")
	ip := addServerCmd.String("ip", "", "(Requis) Adresse IP du serveur.")
	user := addServerCmd.String("user", "", "(Requis) Nom d'utilisateur pour la connexion.")
	sshPort := addServerCmd.Int("ssh-port", 22, "Port SSH pour la connexion.")
	keyPath := addServerCmd.String("key-path", "", "Chemin vers la clé privée SSH.")
	defaultModule := addServerCmd.String("default-module", "", "Module rsync par défaut à utiliser.")

	addServerCmd.Parse(args)

	// Validation des arguments requis
	if *name == "" || *ip == "" || *user == "" {
		fmt.Fprintln(os.Stderr, "Erreur: Les flags --name, --ip, et --user sont requis.")
		addServerCmd.Usage()
		os.Exit(1)
	}

	// Création de la configuration du serveur
	serverConfig := common.RsyncServerConfig{
		Name:              *name,
		IP:                *ip,
		Username:          *user,
		SSHPort:           *sshPort,
		SSHPrivateKeyPath: *keyPath,
		DefaultModule:     *defaultModule,
		Port:              873, // Port rsync par défaut
	}

	// Ajout du serveur à la configuration globale
	if err := common.AddRsyncServer(serverConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur lors de l'ajout du serveur: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Serveur rsync '%s' ajouté avec succès.\n", *name)
	common.LogInfo("Serveur rsync '%s' ajouté via la ligne de commande.", *name)
}
