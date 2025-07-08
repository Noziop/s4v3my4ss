package wrappers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Noziop/s4v3my4ss/pkg/common"
)

// Variable globale pour suivre l'état de la sauvegarde
var isBackupInProgress = false

// GetBackupStatus renvoie si une sauvegarde est en cours
func GetBackupStatus() bool {
	return isBackupInProgress
}

// SetBackupStatus définit si une sauvegarde est en cours
func SetBackupStatus(status bool) {
	isBackupInProgress = status
}

// RsyncWrapper gère les appels à rsync pour la création et restauration de sauvegardes
type RsyncWrapper struct {
	// Vérifié indique si rsync est disponible sur le système
	Verified bool
}

// NewRsyncWrapper crée une nouvelle instance de RsyncWrapper
func NewRsyncWrapper() (*RsyncWrapper, error) {
	rw := &RsyncWrapper{
		Verified: false,
	}

	// Vérifier si rsync est disponible
	if !common.IsCommandAvailable("rsync") {
		return rw, fmt.Errorf("rsync n'est pas installé")
	}

	rw.Verified = true
	return rw, nil
}

// EnsureAvailable vérifie que rsync est disponible, et tente de l'installer si ce n'est pas le cas
func (rw *RsyncWrapper) EnsureAvailable() error {
	if rw.Verified {
		return nil
	}

	if !common.IsCommandAvailable("rsync") {
		err := common.EnsureDependency("rsync", "rsync")
		if err != nil {
			return fmt.Errorf("impossible d'installer rsync: %w", err)
		}
	}

	rw.Verified = true
	return nil
}

// RsyncOptions contient les options pour une opération rsync
type RsyncOptions struct {
	Source      string
	Destination string
	Exclude     []string
	Delete      bool
	Archive     bool
	Compression bool
	Progress    bool
	Remote      bool
	SSHPort     int
	RsyncPort   int      // Port pour le démon rsync (par défaut 873)
	Username    string
	Hostname    string   // Ajout du champ Hostname manquant
	Module      string
	Incremental bool   // Ajout pour sauvegardes incrémentales
	LinkDest    string // Chemin vers la sauvegarde précédente
}

// ExecuteRsync exécute une commande rsync avec les options spécifiées
func ExecuteRsync(options RsyncOptions) error {
    // Déterminer le mode basé sur la destination
    if strings.Contains(options.Destination, "::") {
        // Mode daemon rsync - essayer seulement le port 873
        return executeDaemonRsync(options)
    } else if options.Remote {
        // Mode SSH - essayer les ports SSH
        return executeSSHRsync(options)
    }
    
    // Mode local
    return executeLocalRsync(options)
}

// RsyncBackup effectue une sauvegarde avec rsync
func RsyncBackup(source, destination string, excludeDirs, excludeFiles []string, compression bool, remoteServer *common.RsyncServerConfig) error {
	// Vérifier que le répertoire source existe
	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("le répertoire source '%s' n'existe pas: %v", source, err)
	}

	// S'assurer que le chemin source se termine par un slash
	if !strings.HasSuffix(source, "/") {
		source = source + "/"
	}

	// Construire la liste des exclusions
	excludes := []string{}
	for _, dir := range excludeDirs {
		excludes = append(excludes, dir)
	}
	for _, file := range excludeFiles {
		excludes = append(excludes, file)
	}

	// Générer un horodatage unique pour cette sauvegarde
	timestamp := fmt.Sprintf("%s", strings.Replace(
		strings.Replace(
			strings.Replace(
				time.Now().Format("2006-01-02 15:04:05"), 
				" ", "_", -1), 
			":", "-", -1), 
		"+", "", -1))

	// Construire le chemin de destination final avec timestamp pour éviter d'écraser les sauvegardes existantes
	finalDestination := destination
	var lastBackupPath string
	var isIncremental bool

	// Si on utilise un serveur distant, préparer des chemins spécifiques
	if remoteServer != nil {
		// Extraire le nom du répertoire source pour l'utiliser comme base du nom du répertoire de sauvegarde
		sourceBaseName := filepath.Base(strings.TrimSuffix(source, "/"))
		
		// Construire le chemin de destination sur le serveur distant incluant le timestamp
		if remoteServer.DefaultModule != "" {
			// Format pour destination distante avec module rsync
			// Format: user@host::module/sourceBaseName/timestamp/
			finalDestination = fmt.Sprintf("%s::%s/%s_%s_%s/",
				remoteServer.IP,
				remoteServer.DefaultModule,
				sourceBaseName,
				timestamp,
				common.GenerateRandomString(6))
			
			if remoteServer.Username != "" {
				finalDestination = fmt.Sprintf("%s@%s", remoteServer.Username, finalDestination)
			}
		} else {
			// Format pour destination SSH: user@host:path/sourceBaseName/timestamp/
			finalDestination = fmt.Sprintf("%s@%s:%s/%s_%s_%s/",
				remoteServer.Username,
				remoteServer.IP,
				remoteServer.DefaultPath,
				sourceBaseName,
				timestamp,
				common.GenerateRandomString(6))
		}

		// Chercher la dernière sauvegarde pour créer une sauvegarde incrémentale
		backups, err := common.ListBackups()
		if err == nil {
			var lastBackup *common.BackupInfo
			for i := range backups {
				b := &backups[i]
				// Vérifier si c'est une sauvegarde du même répertoire source vers le même serveur distant
				if b.SourcePath == source && 
					b.RemoteServer != nil && 
					b.RemoteServer.IP == remoteServer.IP {
						if lastBackup == nil || b.Time.After(lastBackup.Time) {
							lastBackup = b
						}
				}
			}

			if lastBackup != nil {
				// Utiliser le chemin de la dernière sauvegarde pour link-dest
				lastBackupPath = lastBackup.BackupPath
				isIncremental = true
				fmt.Printf("Sauvegarde incrémentale basée sur: %s\n", lastBackupPath)
			} else {
				fmt.Println("Aucune sauvegarde précédente trouvée. Création d'une sauvegarde complète.")
			}
		}
	} else {
		// Pour les sauvegardes locales, créer un sous-répertoire avec timestamp
		finalDestination = filepath.Join(destination, timestamp)
		
		// Vérifier s'il existe déjà des sauvegardes pour ce répertoire
		backups, err := common.ListBackups()
		if err == nil {
			var lastBackup *common.BackupInfo
			for i := range backups {
				b := &backups[i]
				// Vérifier si c'est une sauvegarde du même répertoire source
				if b.SourcePath == source && b.RemoteServer == nil {
					if lastBackup == nil || b.Time.After(lastBackup.Time) {
						lastBackup = b
					}
				}
			}

			if lastBackup != nil {
				// Utiliser le chemin de la dernière sauvegarde pour link-dest
				lastBackupPath = lastBackup.BackupPath
				isIncremental = true
				fmt.Printf("Sauvegarde incrémentale basée sur: %s\n", lastBackupPath)
			} else {
				fmt.Println("Aucune sauvegarde précédente trouvée. Création d'une sauvegarde complète.")
			}
		}
	}

	// Préparer les options
	options := RsyncOptions{
		Source:      source,
		Destination: finalDestination,
		Exclude:     excludes,
		Delete:      true,
		Archive:     true,
		Compression: compression,
		Progress:    true,
	}

	// Configurer pour sauvegarde incrémentale si possible
	if isIncremental && lastBackupPath != "" {
		options.Incremental = true
		options.LinkDest = lastBackupPath
	}

	// Si on utilise un serveur distant
	if remoteServer != nil {
		options.Remote = true
		options.SSHPort = remoteServer.SSHPort
		options.Username = remoteServer.Username
		options.Module = remoteServer.DefaultModule
	}

	// Créer le répertoire de destination si c'est local
	if remoteServer == nil {
		if err := os.MkdirAll(finalDestination, 0755); err != nil {
			return fmt.Errorf("impossible de créer le répertoire de destination: %v", err)
		}
	}

	// Exécuter rsync
	fmt.Printf("Sauvegarde de %s vers %s...\n", source, finalDestination)
	if err := ExecuteRsync(options); err != nil {
		return fmt.Errorf("erreur rsync: %v", err)
	}

	return nil
}

// RsyncRestore restaure une sauvegarde avec rsync
func RsyncRestore(source, destination string, remoteServer *common.RsyncServerConfig) error {
	// Vérifier que le répertoire source existe (sauf si c'est distant)
	if remoteServer == nil {
		if _, err := os.Stat(source); err != nil {
			return fmt.Errorf("le répertoire source '%s' n'existe pas: %v", source, err)
		}
	}

	// S'assurer que le chemin source se termine par un slash
	if !strings.HasSuffix(source, "/") {
		source = source + "/"
	}

	// Créer le répertoire de destination s'il n'existe pas
	os.MkdirAll(destination, 0755)

	// Préparer les options
	options := RsyncOptions{
		Source:      source,
		Destination: destination,
		Archive:     true,
		Progress:    true,
	}

	// Si on utilise un serveur distant comme source
	if remoteServer != nil {
		options.Remote = true
		options.SSHPort = remoteServer.SSHPort
		options.Username = remoteServer.Username
		
		// La source devient une référence distante
		if remoteServer.DefaultModule != "" {
			options.Source = fmt.Sprintf("%s@%s::%s/", 
				remoteServer.Username, 
				remoteServer.IP, 
				remoteServer.DefaultModule)
		} else {
			options.Source = fmt.Sprintf("%s@%s:", 
				remoteServer.Username, 
				remoteServer.IP)
		}
	}

	// Exécuter rsync
	if err := ExecuteRsync(options); err != nil {
		return fmt.Errorf("erreur rsync: %v", err)
	}

	return nil
}

// BackupOptions contient les options pour la création de sauvegarde
type BackupOptions struct {
	// Incremental indique s'il s'agit d'une sauvegarde incrémentielle
	Incremental bool
	// LinkDest est le chemin de la sauvegarde précédente pour les sauvegardes incrémentielles
	LinkDest string
	// ExcludeDirs est la liste des répertoires à exclure
	ExcludeDirs []string
	// ExcludeFiles est la liste des fichiers à exclure
	ExcludeFiles []string
}