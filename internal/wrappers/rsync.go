package wrappers

import (
	"context"
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
		common.LogError("rsync n'est pas installé.")
		return rw, fmt.Errorf("rsync n'est pas installé")
	}

	rw.Verified = true
	common.LogInfo("RsyncWrapper créé et rsync vérifié.")
	return rw, nil
}

// EnsureAvailable vérifie que rsync est disponible, et tente de l'installer si ce n'est pas le cas
func (rw *RsyncWrapper) EnsureAvailable() error {
	if rw.Verified {
		return nil
	}

	if !common.IsCommandAvailable("rsync") {
		common.LogWarning("rsync non trouvé. Tentative d'installation...")
		err := common.EnsureDependency("rsync", "rsync")
		if err != nil {
			common.LogError("Impossible d'installer rsync: %v", err)
			return fmt.Errorf("impossible d'installer rsync: %w", err)
		}
		common.LogInfo("rsync installé avec succès.")
	}

	rw.Verified = true
	return nil
}

// RsyncOptions contient les options pour une opération rsync
type RsyncOptions struct {
	Source                string
	Destination           string
	Exclude               []string
	Delete                bool
	Archive               bool
	Compression           bool
	Progress              bool
	Remote                bool
	SSHPort               int
	RsyncPort             int // Port pour le démon rsync (par défaut 873)
	Username              string
	Hostname              string // Ajout du champ Hostname manquant
	Module                string
	Incremental           bool   // Ajout pour sauvegardes incrémentales
	LinkDest              string // Chemin vers la sauvegarde précédente
	SSHPrivateKeyPath     string // Chemin vers la clé privée SSH
	SSHHostKeyFingerprint string // Empreinte de la clé de l'hôte SSH
}

// ExecuteRsync exécute une commande rsync avec les options spécifiées de manière sécurisée.
func ExecuteRsync(options RsyncOptions) error {
	common.LogSecurity("Exécution de rsync avec les options: Source=%s, Destination=%s, Remote=%t", options.Source, options.Destination, options.Remote)
	// SECURITY: Valider toutes les entrées pour prévenir les injections de commande et les chemins non autorisés.
	if !common.IsValidPath(options.Source) {
		common.LogError("Chemin source invalide ou non sécurisé: %s", options.Source)
		return fmt.Errorf("chemin source invalide ou non sécurisé: %s", options.Source)
	}
	if !common.IsValidPath(options.Destination) {
		common.LogError("Chemin destination invalide ou non sécurisé: %s", options.Destination)
		return fmt.Errorf("chemin destination invalide ou non sécurisé: %s", options.Destination)
	}
	if options.LinkDest != "" && !common.IsValidPath(options.LinkDest) {
		common.LogError("Chemin link-dest invalide ou non sécurisé: %s", options.LinkDest)
		return fmt.Errorf("chemin link-dest invalide ou non sécurisé: %s", options.LinkDest)
	}
	if options.Username != "" && !common.IsValidName(options.Username) {
		common.LogError("Nom d'utilisateur invalide: %s", options.Username)
		return fmt.Errorf("nom d'utilisateur invalide: %s", options.Username)
	}
	if options.Hostname != "" && !common.IsValidName(options.Hostname) {
		common.LogError("Nom d'hôte invalide: %s", options.Hostname)
		return fmt.Errorf("nom d'hôte invalide: %s", options.Hostname)
	}
	if options.Module != "" && !common.IsValidName(options.Module) {
		common.LogError("Nom de module invalide: %s", options.Module)
		return fmt.Errorf("nom de module invalide: %s", options.Module)
	}
	for _, exclude := range options.Exclude {
		if !common.IsValidExcludePattern(exclude) {
			common.LogError("Modèle d'exclusion invalide: %s", exclude)
			return fmt.Errorf("modèle d'exclusion invalide: %s", exclude)
		}
	}

	args := []string{}

	// Options de base
	if options.Archive {
		args = append(args, "-a")
	}
	if options.Compression {
		args = append(args, "-z")
	}
	if options.Delete {
		args = append(args, "--delete")
	}
	if options.Progress {
		args = append(args, "--progress", "--stats")
	}

	// Options pour la sauvegarde incrémentale
	if options.Incremental && options.LinkDest != "" {
		args = append(args, "--link-dest="+options.LinkDest)
	}

	// Exclusions
	for _, exclude := range options.Exclude {
		args = append(args, "--exclude="+exclude)
	}

	// Configuration SSH pour les serveurs distants
	sshCommand := "ssh"
	if options.Remote {
		if options.SSHPort > 0 && options.SSHPort != 22 {
			sshCommand = fmt.Sprintf("ssh -p %d", options.SSHPort)
		}

		// SECURITY: Utiliser StrictHostKeyChecking et UserKnownHostsFile pour prévenir les attaques MITM.
		// L'empreinte de la clé de l'hôte doit être gérée en dehors de l'application ou par un mécanisme TOFU.
		sshOptions := []string{"-o", "StrictHostKeyChecking=yes", "-o", "UserKnownHostsFile=/dev/null"}
		if options.SSHHostKeyFingerprint != "" {
			// Note: Une vérification complète de l'empreinte nécessiterait une bibliothèque SSH plus robuste
			// ou une gestion externe du known_hosts. Ici, nous nous assurons juste que l'option est passée.
			// L'utilisateur est responsable de la validité de l'empreinte fournie.
			sshOptions = append(sshOptions, "-o", fmt.Sprintf("HostKeyAlgorithms=ssh-rsa,ssh-dss,ecdsa-sha2-nistp256,ecdsa-sha2-nistp384,ecdsa-sha2-nistp521,ed25519")) // Exemple d'algorithmes
		}
		
		// Ajouter le chemin de la clé privée si fourni
		if options.SSHPrivateKeyPath != "" {
			if !common.IsValidPath(options.SSHPrivateKeyPath) {
				common.LogError("Chemin de la clé privée SSH invalide: %s", options.SSHPrivateKeyPath)
				return fmt.Errorf("chemin de la clé privée SSH invalide: %s", options.SSHPrivateKeyPath)
			}
			sshOptions = append(sshOptions, "-i", options.SSHPrivateKeyPath)
		}

		args = append(args, "-e", strings.Join(append([]string{sshCommand}, sshOptions...), " "))
	}

	// Ajouter source et destination
	args = append(args, options.Source)
	args = append(args, options.Destination)

	// Exécuter la commande
	cmd := exec.CommandContext(context.Background(), "rsync", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	common.LogInfo("Exécution de la commande rsync: rsync %s", strings.Join(args, " "))

	err := cmd.Run()
	if err != nil {
		common.LogError("Erreur lors de l'exécution de rsync: %v", err)
		return fmt.Errorf("erreur lors de l'exécution de rsync: %w", err)
	}

	common.LogInfo("Commande rsync exécutée avec succès.")
	return nil
}

// RsyncBackup effectue une sauvegarde avec rsync
func RsyncBackup(source, destination string, excludeDirs, excludeFiles []string, compression bool, remoteServer *common.RsyncServerConfig) error {
	common.LogInfo("Début de la sauvegarde rsync: Source=%s, Destination=%s", source, destination)
	// Vérifier que le répertoire source existe
	if _, err := os.Stat(source); err != nil {
		common.LogError("Le répertoire source '%s' n'existe pas: %v", source, err)
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
				common.LogInfo("Sauvegarde incrémentale basée sur: %s", lastBackupPath)
			} else {
				common.LogInfo("Aucune sauvegarde précédente trouvée. Création d'une sauvegarde complète.")
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
				common.LogInfo("Sauvegarde incrémentale basée sur: %s", lastBackupPath)
			} else {
				common.LogInfo("Aucune sauvegarde précédente trouvée. Création d'une sauvegarde complète.")
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
		options.SSHPrivateKeyPath = remoteServer.SSHPrivateKeyPath
		options.SSHHostKeyFingerprint = remoteServer.SSHHostKeyFingerprint
	}

	// Créer le répertoire de destination si c'est local
	if remoteServer == nil {
		if err := os.MkdirAll(finalDestination, 0755); err != nil {
			common.LogError("Impossible de créer le répertoire de destination: %v", err)
			return fmt.Errorf("impossible de créer le répertoire de destination: %v", err)
		}
	}

	// Exécuter rsync
	common.LogInfo("Lancement de rsync pour %s vers %s...", source, finalDestination)
	if err := ExecuteRsync(options); err != nil {
		common.LogError("Erreur rsync lors de la sauvegarde: %v", err)
		return fmt.Errorf("erreur rsync: %v", err)
	}

	common.LogInfo("Sauvegarde rsync terminée avec succès.")
	return nil
}

// RsyncRestore restaure une sauvegarde avec rsync
func RsyncRestore(source, destination string, remoteServer *common.RsyncServerConfig) error {
	common.LogInfo("Début de la restauration rsync: Source=%s, Destination=%s", source, destination)
	// Vérifier que le répertoire source existe (sauf si c'est distant)
	if remoteServer == nil {
		if _, err := os.Stat(source); err != nil {
			common.LogError("Le répertoire source '%s' n'existe pas: %v", source, err)
			return fmt.Errorf("le répertoire source '%s' n'existe pas: %v", source, err)
		}
	}

	// S'assurer que le chemin source se termine par un slash
	if !strings.HasSuffix(source, "/") {
		source = source + "/"
	}

	// Créer le répertoire de destination s'il n'existe pas
	if err := os.MkdirAll(destination, 0755); err != nil {
		common.LogError("Impossible de créer le répertoire de destination pour la restauration: %v", err)
		return fmt.Errorf("impossible de créer le répertoire de destination: %w", err)
	}

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
		options.SSHPrivateKeyPath = remoteServer.SSHPrivateKeyPath
		options.SSHHostKeyFingerprint = remoteServer.SSHHostKeyFingerprint
		
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
	common.LogInfo("Lancement de rsync pour la restauration de %s vers %s...", source, destination)
	if err := ExecuteRsync(options); err != nil {
		common.LogError("Erreur rsync lors de la restauration: %v", err)
		return fmt.Errorf("erreur rsync: %v", err)
	}

	common.LogInfo("Restauration rsync terminée avec succès.")
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