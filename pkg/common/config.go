package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"strings"
)

// Config représente la configuration globale de l'application
type Config struct {
	BackupDirs       []BackupConfig     `json:"backupDirectories"`
	BackupDestinations []BackupDestination `json:"backupDestinations"` // Nouvelle liste de destinations
	BackupDestination string             `json:"backupDestination"`    // Gardé pour rétrocompatibilité
	RsyncServers     []RsyncServerConfig `json:"rsyncServers"`
	RetentionPolicy  RetentionPolicy    `json:"retentionPolicy"`
	LastUpdate       time.Time          `json:"last_update"`
}

// BackupDestination représente une destination où les sauvegardes peuvent être stockées
type BackupDestination struct {
	Name        string `json:"name"`        // Nom unique pour cette destination
	Path        string `json:"path"`        // Chemin où les sauvegardes seront stockées
	Type        string `json:"type"`        // Type: "local", "rsync", "cloud", etc.
	IsDefault   bool   `json:"isDefault"`   // Indique si c'est la destination par défaut
	RsyncServer *RsyncServerConfig `json:"rsyncServer,omitempty"` // Configuration rsync si applicable
}

// BackupConfig contient la configuration pour un répertoire à sauvegarder
type BackupConfig struct {
	SourcePath    string   `json:"sourcePath"`
	Name          string   `json:"name"`
	Compression   bool     `json:"compression"`
	IsIncremental bool     `json:"isIncremental"` // Indique si la sauvegarde doit être incrémentale
	ExcludeDirs   []string `json:"excludeDirs,omitempty"`
	ExcludeFiles  []string `json:"excludeFiles,omitempty"`
	Interval      int      `json:"interval"` // en minutes, 0 pour désactiver
	RemoteServer  *RsyncServerConfig `json:"remoteServer,omitempty"` // Serveur rsync distant
	DestinationName string `json:"destinationName,omitempty"` // Nom de la destination à utiliser (si vide, utilise la destination par défaut)
}

// RetentionPolicy définit combien de temps les sauvegardes sont conservées
type RetentionPolicy struct {
	KeepDaily   int `json:"keepDaily"`
	KeepWeekly  int `json:"keepWeekly"`
	KeepMonthly int `json:"keepMonthly"`
}

// RsyncServerConfig contient les paramètres d'un serveur rsync distant
type RsyncServerConfig struct {
	Name          string   `json:"name"`
	IP            string   `json:"ip"`
	Port          int      `json:"port"`
	SSHPort       int      `json:"sshPort"`
	Username      string   `json:"username"`
	Modules       []string `json:"modules"`
	DefaultModule string   `json:"defaultModule"`
	DefaultPath   string   `json:"defaultPath"` // Chemin par défaut sur le serveur distant
}

// BackupInfo contient les informations sur une sauvegarde
type BackupInfo struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	SourcePath   string    `json:"sourcePath"`
	BackupPath   string    `json:"backupPath"`
	Time         time.Time `json:"time"`
	Size         int64     `json:"size"`
	IsIncremental bool     `json:"isIncremental"`
	Compression  bool      `json:"compression"`
	RemoteServer *RsyncServerConfig `json:"remoteServer,omitempty"` // Serveur rsync distant si applicable
	DestinationName string `json:"destinationName,omitempty"` // Nom de la destination utilisée
}

// Variables globales
var (
	AppConfig     Config
	ConfigFile    string
	BackupInfoDir string
	TempDir       string
)

// Initialise l'application: répertoires, configuration, etc.
func InitApp() error {
	configDir, err := GetConfigDir()
	if (err != nil) {
		return fmt.Errorf("impossible de créer le répertoire de configuration: %w", err)
	}

	// Définir les chemins des fichiers de configuration
	ConfigFile = filepath.Join(configDir, "config.json")
	BackupInfoDir = filepath.Join(configDir, "backups")
	
	// Créer le répertoire des métadonnées de sauvegarde
	if err := os.MkdirAll(BackupInfoDir, 0755); err != nil {
		return fmt.Errorf("impossible de créer le répertoire des métadonnées: %w", err)
	}

	// Créer un répertoire temporaire
	tmpDir, err := GetTempDir()
	if err != nil {
		return fmt.Errorf("impossible de créer le répertoire temporaire: %w", err)
	}
	TempDir = tmpDir

	// Note: La suppression automatique du répertoire temporaire à la sortie
	// n'est pas implémentée car os.Atexit n'est disponible qu'à partir de Go 1.21
	// Le nettoyage peut être fait manuellement ou via des routines de l'application

	// Vérifier si la configuration existe, sinon créer une configuration par défaut
	if !FileExists(ConfigFile) {
		defaultConfig := Config{
			BackupDirs: []BackupConfig{},
			BackupDestination: filepath.Join(configDir, "backups"),
			BackupDestinations: []BackupDestination{
				{
					Name:      "Local",
					Path:      filepath.Join(configDir, "backups"),
					Type:      "local",
					IsDefault: true,
				},
			},
			RetentionPolicy: RetentionPolicy{
				KeepDaily:   7,
				KeepWeekly:  4,
				KeepMonthly:  3,
			},
		}
		
		if err := SaveConfig(defaultConfig); err != nil {
			return fmt.Errorf("impossible de créer la configuration par défaut: %w", err)
		}
	}

	// Charger la configuration
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("impossible de charger la configuration: %w", err)
	}
	
	// Migration: si aucune destination n'est définie mais BackupDestination existe
	if len(config.BackupDestinations) == 0 && config.BackupDestination != "" {
		config.BackupDestinations = []BackupDestination{
			{
				Name:      "Default",
				Path:      config.BackupDestination,
				Type:      "local",
				IsDefault: true,
			},
		}
		// Sauvegarder la configuration migrée
		if err := SaveConfig(config); err != nil {
			return fmt.Errorf("impossible de migrer la configuration: %w", err)
		}
	}
	
	AppConfig = config

	return nil
}

// LoadConfig charge la configuration depuis le fichier
func LoadConfig() (Config, error) {
	var config Config
	
	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		return config, err
	}
	
	if err := json.Unmarshal(data, &config); err != nil {
		return config, err
	}
	
	return config, nil
}

// SaveConfig sauvegarde la configuration dans le fichier
func SaveConfig(config Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(ConfigFile, data, 0644)
}

// AddBackupDirectory ajoute un répertoire à sauvegarder à la configuration
func AddBackupDirectory(config BackupConfig) error {
	AppConfig.BackupDirs = append(AppConfig.BackupDirs, config)
	return SaveConfig(AppConfig)
}

// GetBackupConfig récupère la configuration d'un répertoire de sauvegarde par son nom
func GetBackupConfig(name string) (BackupConfig, bool) {
	for _, cfg := range AppConfig.BackupDirs {
		if cfg.Name == name {
			return cfg, true
		}
	}
	return BackupConfig{}, false
}

// AddBackupDestination ajoute une destination de sauvegarde à la configuration
func AddBackupDestination(dest BackupDestination) error {
	// Si cette destination est définie comme par défaut, désactiver le flag pour toutes les autres
	if dest.IsDefault {
		for i := range AppConfig.BackupDestinations {
			AppConfig.BackupDestinations[i].IsDefault = false
		}
	}
	
	// Si aucune destination n'est définie comme par défaut, marquer celle-ci comme par défaut
	if len(AppConfig.BackupDestinations) == 0 {
		dest.IsDefault = true
	}
	
	AppConfig.BackupDestinations = append(AppConfig.BackupDestinations, dest)
	
	// Mettre à jour aussi le champ BackupDestination pour compatibilité
	if dest.IsDefault {
		AppConfig.BackupDestination = dest.Path
	}
	
	return SaveConfig(AppConfig)
}

// GetDefaultBackupDestination récupère la destination de sauvegarde par défaut
func GetDefaultBackupDestination() (BackupDestination, bool) {
	for _, dest := range AppConfig.BackupDestinations {
		if dest.IsDefault {
			return dest, true
		}
	}
	
	// Fallback: utiliser la première destination si aucune n'est marquée comme par défaut
	if len(AppConfig.BackupDestinations) > 0 {
		return AppConfig.BackupDestinations[0], true
	}
	
	// Si aucune destination n'est définie, essayer d'utiliser l'ancienne valeur BackupDestination
	if AppConfig.BackupDestination != "" {
		// Détecter le type de destination en fonction du préfixe
		destinationType := "local"
		destinationName := "Default (local)"
		
		if strings.HasPrefix(AppConfig.BackupDestination, "rsync://") {
			destinationType = "rsync"
			destinationName = "Default (rsync)"
		}
		
		return BackupDestination{
			Name:      destinationName,
			Path:      AppConfig.BackupDestination,
			Type:      destinationType,
			IsDefault: true,
		}, true
	}
	
	return BackupDestination{}, false
}

// GetBackupDestination récupère une destination de sauvegarde par son nom
func GetBackupDestination(name string) (BackupDestination, bool) {
	for _, dest := range AppConfig.BackupDestinations {
		if dest.Name == name {
			return dest, true
		}
	}
	return BackupDestination{}, false
}

// DeleteBackupDestination supprime une destination de sauvegarde
func DeleteBackupDestination(name string) error {
	for i, dest := range AppConfig.BackupDestinations {
		if dest.Name == name {
			// Vérifier si c'est la destination par défaut
			isDefault := dest.IsDefault
			
			// Supprimer la destination
			AppConfig.BackupDestinations = append(
				AppConfig.BackupDestinations[:i],
				AppConfig.BackupDestinations[i+1:]...,
			)
			
			// Si c'était la destination par défaut, définir une nouvelle destination par défaut
			if isDefault && len(AppConfig.BackupDestinations) > 0 {
				AppConfig.BackupDestinations[0].IsDefault = true
				AppConfig.BackupDestination = AppConfig.BackupDestinations[0].Path
			}
			
			return SaveConfig(AppConfig)
		}
	}
	
	return fmt.Errorf("destination de sauvegarde '%s' non trouvée", name)
}

// SaveBackupInfo sauvegarde les métadonnées d'une sauvegarde
func SaveBackupInfo(info BackupInfo) error {
	filename := filepath.Join(BackupInfoDir, info.ID+".json")
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filename, data, 0644)
}

// ListBackups liste toutes les sauvegardes disponibles
func ListBackups() ([]BackupInfo, error) {
	var backups []BackupInfo
	
	files, err := os.ReadDir(BackupInfoDir)
	if err != nil {
		return nil, err
	}
	
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			data, err := os.ReadFile(filepath.Join(BackupInfoDir, file.Name()))
			if err != nil {
				continue
			}
			
			var info BackupInfo
			if err := json.Unmarshal(data, &info); err != nil {
				continue
			}
			
			backups = append(backups, info)
		}
	}
	
	return backups, nil
}

// AddRsyncServer ajoute un serveur rsync à la configuration
func AddRsyncServer(server RsyncServerConfig) error {
	err := ensureConfigLoaded()
	if err != nil {
		return err
	}

	// Vérifier si un serveur avec le même nom existe déjà
	for i, s := range AppConfig.RsyncServers {
		if s.Name == server.Name {
			// Mise à jour du serveur existant
			AppConfig.RsyncServers[i] = server
			return SaveConfig(AppConfig)
		}
	}

	// Ajouter le nouveau serveur
	AppConfig.RsyncServers = append(AppConfig.RsyncServers, server)
	return SaveConfig(AppConfig)
}

// GetRsyncServers retourne la liste des serveurs rsync configurés
func GetRsyncServers() ([]RsyncServerConfig, error) {
	err := ensureConfigLoaded()
	if err != nil {
		return nil, err
	}

	return AppConfig.RsyncServers, nil
}

// GetRsyncServer retourne un serveur rsync par son nom
func GetRsyncServer(name string) (*RsyncServerConfig, error) {
	servers, err := GetRsyncServers()
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		if server.Name == name {
			return &server, nil
		}
	}

	return nil, fmt.Errorf("serveur rsync '%s' non trouvé", name)
}

// DeleteRsyncServer supprime un serveur rsync de la configuration
func DeleteRsyncServer(name string) error {
	err := ensureConfigLoaded()
	if err != nil {
		return err
	}

	for i, server := range AppConfig.RsyncServers {
		if server.Name == name {
			// Suppression du serveur
			AppConfig.RsyncServers = append(AppConfig.RsyncServers[:i], AppConfig.RsyncServers[i+1:]...)
			return SaveConfig(AppConfig)
		}
	}

	return fmt.Errorf("serveur rsync '%s' non trouvé", name)
}

// ensureConfigLoaded vérifie si la configuration est chargée et la charge si nécessaire
func ensureConfigLoaded() error {
	// Si AppConfig est vide, charger la configuration
	if len(AppConfig.BackupDirs) == 0 && len(AppConfig.BackupDestinations) == 0 && AppConfig.BackupDestination == "" {
		config, err := LoadConfig()
		if err != nil {
			return fmt.Errorf("impossible de charger la configuration: %w", err)
		}
		AppConfig = config
	}
	return nil
}

// DeleteBackup supprime une sauvegarde par son ID
func DeleteBackup(id string) error {
	// Chercher la sauvegarde
	backups, err := ListBackups()
	if err != nil {
		return fmt.Errorf("impossible de récupérer la liste des sauvegardes: %w", err)
	}
	
	var backup BackupInfo
	found := false
	for _, b := range backups {
		if b.ID == id {
			backup = b
			found = true
			break
		}
	}
	
	if !found {
		return fmt.Errorf("sauvegarde avec ID %s non trouvée", id)
	}
	
	// Supprimer le fichier de sauvegarde s'il est local
	if backup.RemoteServer == nil {
		// Si c'est un fichier local
		if FileExists(backup.BackupPath) {
			// Si c'est un fichier (sauvegarde compressée)
			if err := os.Remove(backup.BackupPath); err != nil {
				return fmt.Errorf("impossible de supprimer le fichier de sauvegarde %s: %w", backup.BackupPath, err)
			}
		} else if DirExists(backup.BackupPath) {
			// Si c'est un répertoire (sauvegarde non compressée)
			if err := os.RemoveAll(backup.BackupPath); err != nil {
				return fmt.Errorf("impossible de supprimer le répertoire de sauvegarde %s: %w", backup.BackupPath, err)
			}
		}
	}
	
	// Supprimer le fichier de métadonnées
	metaPath := filepath.Join(BackupInfoDir, id+".json")
	if err := os.Remove(metaPath); err != nil {
		return fmt.Errorf("impossible de supprimer le fichier de métadonnées %s: %w", metaPath, err)
	}
	
	return nil
}