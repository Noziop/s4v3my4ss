package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config représente la configuration globale de l'application
type Config struct {
	BackupDirs        []BackupConfig     `json:"backupDirectories"`
	BackupDestination string             `json:"backupDestination"`
	RsyncServers      []RsyncServerConfig `json:"rsyncServers"`
	RetentionPolicy   RetentionPolicy    `json:"retentionPolicy"`
	LastUpdate        time.Time          `json:"last_update"`
}

// BackupConfig contient la configuration pour un répertoire à sauvegarder
type BackupConfig struct {
	SourcePath   string   `json:"sourcePath"`
	Name         string   `json:"name"`
	Compression  bool     `json:"compression"`
	ExcludeDirs  []string `json:"excludeDirs,omitempty"`
	ExcludeFiles []string `json:"excludeFiles,omitempty"`
	Interval     int      `json:"interval"` // en minutes, 0 pour désactiver
	RemoteServer *RsyncServerConfig `json:"remoteServer,omitempty"` // Serveur rsync distant
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
			RetentionPolicy: RetentionPolicy{
				KeepDaily:   7,
				KeepWeekly:  4,
				KeepMonthly: 3,
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
	if len(AppConfig.BackupDirs) == 0 && AppConfig.BackupDestination == "" {
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