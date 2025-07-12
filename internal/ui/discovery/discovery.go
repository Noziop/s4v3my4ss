package discovery

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Noziop/s4v3my4ss/internal/ui/display"
	"github.com/Noziop/s4v3my4ss/internal/ui/input"
	"github.com/Noziop/s4v3my4ss/internal/wrappers"
	"github.com/Noziop/s4v3my4ss/pkg/common"
)

// DiscoverRsyncServers recherche et configure les serveurs rsync sur le réseau (mode interactif)
func DiscoverRsyncServers() {
	common.LogInfo("Début de la découverte des serveurs rsync.")
	display.ClearScreen()
	fmt.Printf("%sRecherche de serveurs rsync sur le réseau%s\n\n", display.ColorBold(), display.ColorReset())

	// Demander le sous-réseau à scanner
	fmt.Println("Entrez le sous-réseau à scanner (format CIDR, par exemple 192.168.0.0/24)")
	fmt.Println("ou laissez vide pour utiliser 192.168.0.0/24:")

	subnetCIDR := input.ReadAndValidateInput("Sous-réseau: ", common.IsValidSubnet, "Format CIDR invalide (ex: 192.168.0.0/24).") // Utilisation de common.IsValidSubnet
	if subnetCIDR == "" {
		subnetCIDR = "192.168.0.0/24"
		common.LogInfo("Sous-réseau par défaut utilisé pour la découverte: %s", subnetCIDR)
	}

	servers := runDiscovery(subnetCIDR)

	if len(servers) == 0 {
		common.LogInfo("Aucun serveur rsync trouvé sur le réseau %s.", subnetCIDR)
		fmt.Printf("\n%sAucun serveur rsync trouvé sur le réseau.%s\n", display.ColorYellow(), display.ColorReset())
		fmt.Println("Vérifiez que:")
		fmt.Println("1. Le service rsync est activé sur votre NAS ou serveur")
		fmt.Println("2. Le port rsync (873) est ouvert dans le pare-feu")
		fmt.Println("3. Le sous-réseau spécifié est correct")
		return
	}

	common.LogInfo("%d serveurs rsync trouvés sur le réseau %s.", len(servers), subnetCIDR)
	// Afficher les serveurs trouvés
	fmt.Printf("\n%s%d serveurs rsync trouvés:%s\n\n", display.ColorGreen(), len(servers), display.ColorReset())

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
	choiceStr := input.ReadInput("\nChoisissez un serveur à configurer (0 pour annuler): ")
	choice, err := strconv.Atoi(choiceStr)

	if err != nil || choice < 1 || choice > len(servers) {
		if choiceStr != "0" {
			common.LogWarning("Choix de serveur invalide pour la configuration: %s", choiceStr)
			fmt.Printf("\n%sChoix invalide.%s\n", display.ColorRed(), display.ColorReset())
		}
		return
	}

	// Configurer le serveur sélectionné
	selectedServer := servers[choice-1]
	common.LogInfo("Serveur %s sélectionné pour la configuration.", selectedServer.IP)
	configureRsyncServer(selectedServer)
}

// runDiscovery exécute la logique de découverte de réseau et retourne les serveurs trouvés.
func runDiscovery(subnetCIDR string) []wrappers.RsyncServer {
	// Afficher un message d'attente
	fmt.Printf("\nRecherche de serveurs rsync sur %s...\n", subnetCIDR)
	fmt.Println("Cela peut prendre jusqu'à 30 secondes. Veuillez patienter...")

	// Créer l'objet de découverte
	discovery := wrappers.NewRsyncDiscovery()

	// Scanner le réseau (timeout de 30 secondes)
	return discovery.ScanNetwork(subnetCIDR, 30)
}

// configureRsyncServer configure un serveur rsync sélectionné
func configureRsyncServer(server wrappers.RsyncServer) {
	common.LogInfo("Début de la configuration du serveur rsync: %s.", server.IP)
	display.ClearScreen()
	serverName := server.IP
	if server.Hostname != "" {
		serverName = server.Hostname
	}

	fmt.Printf("%sConfiguration du serveur rsync: %s (%s)%s\n\n",
		display.ColorBold, serverName, server.IP, display.ColorReset)

	// Demander un nom pour le serveur
	defaultName := serverName
	if server.Hostname != "" {
		// Extraire le nom d'hôte sans le domaine si possible
		parts := strings.Split(server.Hostname, ".")
		defaultName = parts[0]
	}

	name := input.ReadAndValidateInput(fmt.Sprintf("Nom du serveur (défaut: %s): ", defaultName), common.IsValidName, "Nom invalide.") // Utilisation de common.IsValidName
	if name == "" {
		name = defaultName
	}

	// Demander le nom d'utilisateur
	username := input.ReadAndValidateInput("Nom d'utilisateur pour la connexion: ", common.IsValidName, "Nom d'utilisateur invalide.") // Utilisation de common.IsValidName

	// Si des modules sont disponibles, demander lequel utiliser
	defaultModule := ""
	if len(server.Modules) > 0 {
		common.LogInfo("Modules disponibles pour %s: %v", server.IP, server.Modules)
		fmt.Println("\nModules disponibles:")
		for i, module := range server.Modules {
			fmt.Printf("%d. %s: %s\n", i+1, module.Name, module.Description)
		}

		moduleChoice := input.ReadInput("\nChoisissez un module par défaut (0 pour aucun): ")

		if moduleIdx, err := strconv.Atoi(moduleChoice); err == nil && moduleIdx > 0 && moduleIdx <= len(server.Modules) {
			defaultModule = server.Modules[moduleIdx-1].Name
			common.LogInfo("Module par défaut sélectionné: %s", defaultModule)
		} else {
			common.LogWarning("Choix de module invalide ou aucun module sélectionné: %s", moduleChoice)
		}
	}

	// Créer la configuration du serveur
	rsincModuleNames := make([]string, 0, len(server.Modules))
	for _, module := range server.Modules {
		rsincModuleNames = append(rsincModuleNames, module.Name)
	}

	serverConfig := common.RsyncServerConfig{
		Name:          name,
		IP:            server.IP,
		Port:          server.Port,
		SSHPort:       server.SSHPort,
		Username:      username,
		Modules:       rsincModuleNames,
		DefaultModule: defaultModule,
	}

	// Enregistrer le serveur dans la configuration
	if err := common.AddRsyncServer(serverConfig); err != nil {
		common.LogError("Erreur lors de la sauvegarde de la configuration du serveur rsync %s: %v", serverConfig.Name, err)
		fmt.Printf("\n%sErreur lors de la sauvegarde de la configuration: %v%s\n", display.ColorRed(), err, display.ColorReset())
	} else {
		common.LogInfo("Serveur rsync %s configuré avec succès.", serverConfig.Name)
		fmt.Printf("\n%sServeur rsync %s configuré avec succès!%s\n", display.ColorGreen, name, display.ColorReset)
	}

	// Proposer de configurer une sauvegarde vers ce serveur
	configBackup := input.ReadInput("\nVoulez-vous configurer une sauvegarde vers ce serveur maintenant? (o/n): ")

	if strings.ToLower(configBackup) == "o" {
		common.LogInfo("Configuration d'une sauvegarde distante vers %s demandée.", serverConfig.Name)
		//TODO: commands.ConfigureRemoteBackup(serverConfig)
	} else {
		common.LogInfo("Configuration d'une sauvegarde distante vers %s annulée.", serverConfig.Name)
	}
}