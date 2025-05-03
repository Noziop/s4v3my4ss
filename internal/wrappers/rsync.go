package wrappers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
	Username    string
	Module      string
	Incremental bool   // Ajout pour sauvegardes incrémentales
	LinkDest    string // Chemin vers la sauvegarde précédente
}

// Execute exécute une commande rsync avec les options fournies
func ExecuteRsync(options RsyncOptions) error {
	// Vérifier que rsync est installé
	if _, err := exec.LookPath("rsync"); err != nil {
		return fmt.Errorf("rsync n'est pas installé: %v", err)
	}

	// Construire les arguments de la commande rsync
	args := []string{}

	// Options communes
	if options.Archive {
		args = append(args, "--archive")
	}
	if options.Compression {
		args = append(args, "--compress")
	}
	if options.Progress {
		args = append(args, "--progress")
	}
	if options.Delete {
		args = append(args, "--delete")
	}

	// Option pour sauvegarde incrémentale
	if options.Incremental && options.LinkDest != "" {
		args = append(args, "--link-dest="+options.LinkDest)
	}

	// Gérer les exclusions
	for _, exclude := range options.Exclude {
		args = append(args, "--exclude", exclude)
	}

	// Options pour connexion distante
	if options.Remote && options.SSHPort > 0 {
		args = append(args, "-e", fmt.Sprintf("ssh -p %d", options.SSHPort))
	}

	// Source et destination
	src := options.Source
	dest := options.Destination

	// Si c'est une destination distante mais pas formatée comme telle
	if options.Remote && !strings.Contains(dest, ":") {
		// Format pour destination distante avec module
		if options.Module != "" {
			dest = fmt.Sprintf("%s@%s::%s", options.Username, dest, options.Module)
		} else {
			// Format pour destination distante sans module (avec SSH)
			dest = fmt.Sprintf("%s@%s:", options.Username, dest)
		}
	}

	args = append(args, src, dest)

	// Exécuter la commande
	cmd := exec.Command("rsync", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
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

	// Préparer les options
	options := RsyncOptions{
		Source:      source,
		Destination: destination,
		Exclude:     excludes,
		Delete:      true,
		Archive:     true,
		Compression: compression,
		Progress:    true,
		Incremental: true, // Activer par défaut les sauvegardes incrémentales
	}

	// Si on utilise un serveur distant
	if remoteServer != nil {
		options.Remote = true
		options.SSHPort = remoteServer.SSHPort
		options.Username = remoteServer.Username
		options.Module = remoteServer.DefaultModule
		options.Destination = remoteServer.IP

		// Trouver la dernière sauvegarde pour le même chemin source
		// pour configurer le link-dest pour une sauvegarde incrémentale
		backups, err := common.ListBackups()
		if err == nil {
			var lastBackup *common.BackupInfo
			for i := range backups {
				b := &backups[i]
				// Vérifier si c'est une sauvegarde du même répertoire source
				// et vers le même serveur distant
				if b.SourcePath == source && 
				   b.RemoteServer != nil && 
				   b.RemoteServer.IP == remoteServer.IP {
					if lastBackup == nil || b.Time.After(lastBackup.Time) {
						lastBackup = b
					}
				}
			}

			// Si une sauvegarde précédente a été trouvée, configurer link-dest
			if lastBackup != nil {
				// Construire le chemin de la dernière sauvegarde
				// Format: rsync://user@host/module/path
				var linkDestPath string
				if remoteServer.DefaultModule != "" {
					// Avec module
					linkDestPath = fmt.Sprintf("rsync://%s@%s/%s/previous", 
						remoteServer.Username, 
						remoteServer.IP,
						remoteServer.DefaultModule)
				} else {
					// Sans module (avec SSH)
					linkDestPath = fmt.Sprintf("%s@%s:previous", 
						remoteServer.Username, 
						remoteServer.IP)
				}
				
				options.LinkDest = linkDestPath
				fmt.Printf("Mode sauvegarde incrémentale activé. Référence: %s\n", linkDestPath)
			} else {
				fmt.Println("Aucune sauvegarde précédente trouvée. Création d'une sauvegarde complète.")
			}
		}
	}

	// Exécuter rsync
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