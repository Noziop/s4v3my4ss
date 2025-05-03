package wrappers

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// RsyncModule représente un module rsync disponible sur un serveur
type RsyncModule struct {
	Name        string
	Description string
}

// RsyncServer représente un serveur rsync découvert
type RsyncServer struct {
	IP        string
	Hostname  string
	Port      int
	SSHPort   int
	Modules   []RsyncModule
	Available bool
}

// RsyncDiscovery gère la découverte des serveurs rsync sur le réseau
type RsyncDiscovery struct {}

// NewRsyncDiscovery crée une nouvelle instance de RsyncDiscovery
func NewRsyncDiscovery() *RsyncDiscovery {
	return &RsyncDiscovery{}
}

// ScanNetwork recherche les serveurs rsync sur le réseau spécifié
func (r *RsyncDiscovery) ScanNetwork(cidr string, timeoutSeconds int) []RsyncServer {
	// Valider le CIDR
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		fmt.Printf("Format CIDR invalide: %v\n", err)
		return nil
	}

	// Liste des serveurs découverts
	var servers []RsyncServer
	var mutex sync.Mutex

	// Canal pour les résultats
	results := make(chan RsyncServer)

	// Créer un contexte avec timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	// Nombre d'IP à scanner
	var wg sync.WaitGroup

	// Scanner les adresses IP du réseau
	go func() {
		// Convertir le masque de sous-réseau en taille
		ones, bits := ipNet.Mask.Size()
		numIPs := 1 << (bits - ones)

		// Limite du nombre de workers en parallèle
		semaphore := make(chan struct{}, 100)
		
		 // Adresse de départ
		start := ipNet.IP

		// Scanner toutes les adresses du sous-réseau
		for i := 0; i < numIPs; i++ {
			// Éviter l'adresse de réseau (première) et l'adresse de broadcast (dernière)
			if i == 0 || i == numIPs-1 {
				continue
			}

			// Incrémenter l'adresse IP
			ip := make(net.IP, len(start))
			copy(ip, start)
			
			// Calcul de la nouvelle adresse IP
			for j := len(ip) - 1; j >= 0; j-- {
				ip[j] += byte(i)
				if ip[j] != 0 {
					break
				}
			}

			// Vérifier si le contexte est terminé
			select {
			case <-ctx.Done():
				return
			default:
				// Continuer
			}

			// Limiter le nombre de goroutines
			semaphore <- struct{}{}
			wg.Add(1)
			
			// Lancer le scan de cette adresse IP
			go func(ipAddr string) {
				defer wg.Done()
				defer func() { <-semaphore }()

				// Vérifier si rsync est disponible sur cette adresse
				server := checkRsyncServer(ipAddr)
				if server.Available {
					// Obtenir le nom d'hôte
					hostname, _ := net.LookupAddr(ipAddr)
					if len(hostname) > 0 {
						server.Hostname = hostname[0]
					}

					// Vérifier si le port SSH est ouvert
					sshPort := checkSSHPort(ipAddr)
					if sshPort > 0 {
						server.SSHPort = sshPort
					}

					// Obtenir la liste des modules
					server.Modules = getRsyncModules(ipAddr)

					results <- server
				}
			}(ip.String())
		}

		// Attendre que tous les scans soient terminés
		wg.Wait()
		close(results)
	}()

	// Collecter les résultats
	for server := range results {
		mutex.Lock()
		servers = append(servers, server)
		mutex.Unlock()
	}

	return servers
}

// checkRsyncServer vérifie si un serveur rsync est disponible à l'adresse spécifiée
func checkRsyncServer(ip string) RsyncServer {
	server := RsyncServer{
		IP:        ip,
		Port:      873, // Port par défaut pour rsync
		Available: false,
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:873", ip), 500*time.Millisecond)
	if err != nil {
		return server
	}

	defer conn.Close()
	server.Available = true
	return server
}

// checkSSHPort vérifie si un port SSH est ouvert sur le serveur
// Vérifie les ports courants : 22 (défaut) et 2223 (comme mentionné par l'utilisateur)
func checkSSHPort(ip string) int {
	// Liste des ports SSH courants à vérifier
	ports := []int{22, 2223, 2222}

	for _, port := range ports {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), 500*time.Millisecond)
		if err == nil {
			conn.Close()
			return port
		}
	}

	return 0 // Pas de port SSH détecté
}

// getRsyncModules obtient la liste des modules rsync disponibles sur un serveur
func getRsyncModules(ip string) []RsyncModule {
	var modules []RsyncModule

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "rsync", "--list-only", fmt.Sprintf("rsync://%s", ip))
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return modules
	}

	if err := cmd.Start(); err != nil {
		return modules
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "\t", 2)
		
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			desc := strings.TrimSpace(parts[1])
			
			// Ignorer les lignes qui ne sont pas des modules
			if !strings.HasPrefix(name, "@") && name != "" {
				modules = append(modules, RsyncModule{
					Name:        name,
					Description: desc,
				})
			}
		}
	}

	cmd.Wait()
	return modules
}