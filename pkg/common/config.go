package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

// Config représente la configuration globale de l'application
type Config struct {
	BackupDirs         []BackupConfig      `json:"backupDirectories"`
	BackupDestinations []BackupDestination  `json:"backupDestinations"`
	BackupDestination  string              `json:"backupDestination,omitempty"` // Gardé pour rétrocompatibilité
	RsyncServers       []RsyncServerConfig `json:"rsyncServers"`
	RetentionPolicy    RetentionPolicy     `json:"retentionPolicy"`
	Security           SecurityConfig      `json:"security,omitempty"` // Configuration de sécurité
	EncryptionKey      string              `json:"encryptionKey,omitempty"`   // Clé de chiffrement (pour une future implémentation)
	LastUpdate         time.Time           `json:"last_update"`
}

// SecurityConfig contient les paramètres de sécurité de l'application.
// Permet de restreindre les chemins de sauvegarde pour éviter les écritures dans des répertoires sensibles.
type SecurityConfig struct {
	AllowedBackupPaths []string `json:"allowedBackupPaths,omitempty"` // Liste blanche des chemins de destination autorisés
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
	Encrypt       bool     `json:"encrypt,omitempty"` // Activer le chiffrement pour cette sauvegarde
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

// RetentionInterval représente les intervalles de rétention (quotidien, hebdomadaire, mensuel)
type RetentionInterval string

const (
	Daily   RetentionInterval = "daily"
	Weekly  RetentionInterval = "weekly"
	Monthly RetentionInterval = "monthly"
)

// RsyncServerConfig contient les paramètres d'un serveur rsync distant
type RsyncServerConfig struct {
	Name                  string   `json:"name"`
	IP                    string   `json:"ip"`
	Port                  int      `json:"port"`
	SSHPort               int      `json:"sshPort"`
	Username              string   `json:"username"`
	// SECURITY: Le chemin vers la clé privée SSH est plus sécurisé que de stocker un mot de passe.
	// L'utilisateur doit s'assurer que ce fichier est protégé avec des permissions restrictives (ex: 600).
	SSHPrivateKeyPath     string   `json:"sshPrivateKeyPath,omitempty"`
	// SECURITY: Spécifier l'empreinte de la clé de l'hôte SSH pour prévenir les attaques Man-in-the-Middle.
	// L'empreinte peut être obtenue avec `ssh-keyscan` ou lors de la première connexion manuelle.
	SSHHostKeyFingerprint string   `json:"sshHostKeyFingerprint,omitempty"`
	Modules               []string `json:"modules"`
	DefaultModule         string   `json:"defaultModule"`
	DefaultPath           string   `json:"defaultPath"` // Chemin par défaut sur le serveur distant
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
	// Initialiser le logger au démarrage de l'application
	if err := InitLogger(); err != nil {
		return fmt.Errorf("impossible d'initialiser le logger: %w", err)
	}
	defer CloseLogger() // S'assurer que le logger est fermé à la sortie de l'application

	configDir, err := GetConfigDir()
	if (err != nil) {
		LogError("Impossible de créer le répertoire de configuration: %v", err)
		return fmt.Errorf("impossible de créer le répertoire de configuration: %w", err)
	}

	// Définir les chemins des fichiers de configuration
	ConfigFile = filepath.Join(configDir, "config.json")
	BackupInfoDir = filepath.Join(configDir, "backups")
	
	// Créer le répertoire des métadonnées de sauvegarde
	if err := os.MkdirAll(BackupInfoDir, 0755); err != nil {
		LogError("Impossible de créer le répertoire des métadonnées: %v", err)
		return fmt.Errorf("impossible de créer le répertoire des métadonnées: %w", err)
	}

	// Créer un répertoire temporaire
	tmpDir, err := GetTempDir()
	if err != nil {
		LogError("Impossible de créer le répertoire temporaire: %v", err)
		return fmt.Errorf("impossible de créer le répertoire temporaire: %w", err)
	}
	TempDir = tmpDir

	// Note: La suppression automatique du répertoire temporaire à la sortie
	// n'est pas implémentée car os.Atexit n'est disponible qu'à partir de Go 1.21
	// Le nettoyage peut être fait manuellement ou via des routines de l'application

	// Vérifier si la configuration existe, sinon créer une configuration par défaut
	if !FileExists(ConfigFile) {
		LogInfo("Fichier de configuration non trouvé. Création d'une configuration par défaut.")
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
			LogError("Impossible de créer la configuration par défaut: %v", err)
			return fmt.Errorf("impossible de créer la configuration par défaut: %w", err)
		}
		LogInfo("Configuration par défaut créée avec succès.")
	}

	// Charger la configuration
	config, err := LoadConfig()
	if err != nil {
		LogError("Impossible de charger la configuration: %v", err)
		return fmt.Errorf("impossible de charger la configuration: %w", err)
	}
	
	// Migration: si aucune destination n'est définie mais BackupDestination existe
	if len(config.BackupDestinations) == 0 && config.BackupDestination != "" {
		LogInfo("Migration de l'ancienne destination de sauvegarde vers le nouveau format.")
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
			LogError("Impossible de migrer la configuration: %v", err)
			return fmt.Errorf("impossible de migrer la configuration: %w", err)
		}
		LogInfo("Migration de la configuration terminée avec succès.")
	}
	
	// SECURITY: Valider la configuration après le chargement
	if err := config.ValidateConfig(); err != nil {
		LogSecurity("Configuration invalide détectée: %v", err)
		return fmt.Errorf("la configuration est invalide: %w", err)
	}

	AppConfig = config
	LogInfo("Configuration chargée et validée avec succès.")

	return nil
}

// LoadConfig charge la configuration depuis le fichier
func LoadConfig() (Config, error) {
	var config Config
	
	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		LogError("Impossible de lire le fichier de configuration %s: %v", ConfigFile, err)
		return config, err
	}
	
	if err := json.Unmarshal(data, &config); err != nil {
		LogError("Impossible de désérialiser le fichier de configuration %s: %v", ConfigFile, err)
		return config, err
	}
	
	LogInfo("Configuration chargée depuis %s.", ConfigFile)
	return config, nil
}

// SaveConfig sauvegarde la configuration dans le fichier
func SaveConfig(config Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		LogError("Impossible de sérialiser la configuration: %v", err)
		return err
	}
	
	// SECURITY: Restreindre les permissions du fichier de configuration
	if err := os.WriteFile(ConfigFile, data, 0600); err != nil {
		LogError("Impossible d'écrire le fichier de configuration %s: %v", ConfigFile, err)
		return err
	}
	LogSecurity("Configuration sauvegardée dans %s avec des permissions restreintes (0600).", ConfigFile)
	return nil
}

// AddBackupDirectory ajoute un répertoire à sauvegarder à la configuration
func AddBackupDirectory(config BackupConfig) error {
	if err := AddConfigItem(&AppConfig.BackupDirs, config, "Name"); err != nil {
		LogError("Impossible d'ajouter le répertoire de sauvegarde '%s': %v", config.Name, err)
		return err
	}
	if err := SaveConfig(AppConfig); err != nil {
		LogError("Impossible d'ajouter le répertoire de sauvegarde '%s': %v", config.Name, err)
		return err
	}
	LogSecurity("Répertoire de sauvegarde '%s' ajouté à la configuration.", config.Name)
	return nil
}

// GetBackupConfig récupère la configuration d'un répertoire de sauvegarde par son nom
func GetBackupConfig(name string) (BackupConfig, bool) {
	if item, found := GetConfigItem(&AppConfig.BackupDirs, name, "Name"); found {
		return item.(BackupConfig), true
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

	if err := AddConfigItem(&AppConfig.BackupDestinations, dest, "Name"); err != nil {
		LogError("Impossible d'ajouter la destination de sauvegarde '%s': %v", dest.Name, err)
		return err
	}

	// Mettre à jour aussi le champ BackupDestination pour compatibilité
	if dest.IsDefault {
		AppConfig.BackupDestination = dest.Path
	}

	if err := SaveConfig(AppConfig); err != nil {
		LogError("Impossible d'ajouter la destination de sauvegarde '%s': %v", dest.Name, err)
		return err
	}
	LogSecurity("Destination de sauvegarde '%s' ajoutée à la configuration.", dest.Name)
	return nil
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
	if item, found := GetConfigItem(&AppConfig.BackupDestinations, name, "Name"); found {
		return item.(BackupDestination), true
	}
	return BackupDestination{}, false
}

// DeleteBackupDestination supprime une destination de sauvegarde
func DeleteBackupDestination(name string) error {
	// Trouver la destination à supprimer pour vérifier si c'est la destination par défaut
	var isDefault bool
	if item, found := GetConfigItem(&AppConfig.BackupDestinations, name, "Name"); found {
		isDefault = item.(BackupDestination).IsDefault
	} else {
		LogError("Destination de sauvegarde '%s' non trouvée pour suppression.", name)
		return fmt.Errorf("destination de sauvegarde '%s' non trouvée", name)
	}

	if err := DeleteConfigItem(&AppConfig.BackupDestinations, name, "Name"); err != nil {
		LogError("Impossible de supprimer la destination de sauvegarde '%s': %v", name, err)
		return err
	}

	// Si c'était la destination par défaut, définir une nouvelle destination par défaut
	if isDefault && len(AppConfig.BackupDestinations) > 0 {
		AppConfig.BackupDestinations[0].IsDefault = true
		AppConfig.BackupDestination = AppConfig.BackupDestinations[0].Path
	}

	if err := SaveConfig(AppConfig); err != nil {
		LogError("Impossible de supprimer la destination de sauvegarde '%s': %v", name, err)
		return err
	}
	LogSecurity("Destination de sauvegarde '%s' supprimée de la configuration.", name)
	return nil
}

// UpdateBackupDestination met à jour une destination de sauvegarde
func UpdateBackupDestination(name string, newDest BackupDestination) error {
	if err := UpdateConfigItem(&AppConfig.BackupDestinations, name, newDest, "Name"); err != nil {
		LogError("Impossible de mettre à jour la destination de sauvegarde '%s': %v", name, err)
		return err
	}
	if err := SaveConfig(AppConfig); err != nil {
		LogError("Impossible de mettre à jour la destination de sauvegarde '%s': %v", name, err)
		return err
	}
	LogSecurity("Destination de sauvegarde '%s' mise à jour dans la configuration.", name)
	return nil
}

// SetDefaultBackupDestination définit une destination de sauvegarde par défaut
func SetDefaultBackupDestination(name string) error {
	found := false
	for i := range AppConfig.BackupDestinations {
		if AppConfig.BackupDestinations[i].Name == name {
			AppConfig.BackupDestinations[i].IsDefault = true
			AppConfig.BackupDestination = AppConfig.BackupDestinations[i].Path
			found = true
		} else {
			AppConfig.BackupDestinations[i].IsDefault = false
		}
	}

	if !found {
		LogError("Destination de sauvegarde '%s' non trouvée pour la définir par défaut.", name)
		return fmt.Errorf("destination de sauvegarde '%s' non trouvée", name)
	}

	// Sauvegarder la configuration après avoir mis à jour les flags IsDefault
	if err := SaveConfig(AppConfig); err != nil {
		LogError("Impossible de définir la destination par défaut sur '%s': %v", name, err)
		return err
	}
	LogSecurity("Destination de sauvegarde par défaut définie sur '%s'.", name)
	return nil
}



// AddRsyncServer ajoute un serveur rsync à la configuration
func AddRsyncServer(server RsyncServerConfig) error {
	err := ensureConfigLoaded()
	if err != nil {
		LogError("Impossible de charger la configuration pour ajouter un serveur rsync: %v", err)
		return err
	}

	if err := AddConfigItem(&AppConfig.RsyncServers, server, "Name"); err != nil {
		LogError("Impossible d'ajouter le serveur rsync '%s': %v", server.Name, err)
		return err
	}

	if err := SaveConfig(AppConfig); err != nil {
		LogError("Impossible d'ajouter le serveur rsync '%s': %v", server.Name, err)
		return err
	}
	LogSecurity("Serveur rsync '%s' ajouté à la configuration.", server.Name)
	return nil
}

// GetRsyncServers retourne la liste des serveurs rsync configurés
func GetRsyncServers() ([]RsyncServerConfig, error) {
	err := ensureConfigLoaded()
	if err != nil {
		LogError("Impossible de charger la configuration pour récupérer les serveurs rsync: %v", err)
		return nil, err
	}

	return AppConfig.RsyncServers, nil
}

// GetRsyncServer retourne un serveur rsync par son nom
func GetRsyncServer(name string) (*RsyncServerConfig, error) {
	err := ensureConfigLoaded()
	if err != nil {
		LogError("Impossible de charger la configuration pour récupérer les serveurs rsync: %v", err)
		return nil, err
	}

	if item, found := GetConfigItem(&AppConfig.RsyncServers, name, "Name"); found {
		server := item.(RsyncServerConfig)
		return &server, nil
	}
	LogError("Serveur rsync '%s' non trouvé.", name)
	return nil, fmt.Errorf("serveur rsync '%s' non trouvé", name)
}

// DeleteRsyncServer supprime un serveur rsync de la configuration
func DeleteRsyncServer(name string) error {
	err := ensureConfigLoaded()
	if err != nil {
		LogError("Impossible de charger la configuration pour supprimer un serveur rsync: %v", err)
		return err
	}

	if err := DeleteConfigItem(&AppConfig.RsyncServers, name, "Name"); err != nil {
		LogError("Impossible de supprimer le serveur rsync '%s': %v", name, err)
		return err
	}

	if err := SaveConfig(AppConfig); err != nil {
		LogError("Impossible de supprimer le serveur rsync '%s': %v", name, err)
		return err
	}
	LogSecurity("Serveur rsync '%s' supprimé de la configuration.", name)
	return nil
}

// ensureConfigLoaded vérifie si la configuration est chargée et la charge si nécessaire
func ensureConfigLoaded() error {
	// Si AppConfig est vide, charger la configuration
	if len(AppConfig.BackupDirs) == 0 && len(AppConfig.BackupDestinations) == 0 && AppConfig.BackupDestination == "" {
		config, err := LoadConfig()
		if err != nil {
			LogError("Impossible de charger la configuration dans ensureConfigLoaded: %v", err)
			return fmt.Errorf("impossible de charger la configuration: %w", err)
		}
		AppConfig = config
	}
	return nil
}

func (sc *SecurityConfig) IsPathAllowed(path string) bool {
	// Si aucune restriction n'est définie, tous les chemins sont autorisés.
	if len(sc.AllowedBackupPaths) == 0 {
		LogInfo("Aucune restriction de chemin de sauvegarde définie. Chemin '%s' autorisé par défaut.", path)
		return true
	}

	// Nettoyer le chemin à vérifier
	cleanPath := filepath.Clean(path)

	// Vérifier si le chemin est un sous-répertoire d'un des chemins autorisés.
	for _, allowedPath := range sc.AllowedBackupPaths {
		cleanAllowedPath := filepath.Clean(allowedPath)
		if strings.HasPrefix(cleanPath, cleanAllowedPath) {
			LogInfo("Chemin de sauvegarde '%s' autorisé car il est sous '%s'.", path, allowedPath)
			return true
		}
	}
	LogSecurity("Chemin de sauvegarde '%s' bloqué car non autorisé.", path)
	return false
}

// ValidateConfig vérifie la validité de la configuration chargée.
func (c *Config) ValidateConfig() error {
	LogInfo("Validation de la configuration...")
	for _, dir := range c.BackupDirs {
		if !IsValidName(dir.Name) {
			LogError("Nom de configuration de sauvegarde invalide: %s", dir.Name)
			return fmt.Errorf("nom de configuration de sauvegarde invalide: %s", dir.Name)
		}
		if !IsValidPath(dir.SourcePath) {
			LogError("Chemin source invalide dans la configuration '%s': %s", dir.Name, dir.SourcePath)
			return fmt.Errorf("chemin source invalide dans la configuration '%s': %s", dir.Name, dir.SourcePath)
		}
	}

	for _, dest := range c.BackupDestinations {
		if !IsValidName(dest.Name) {
			LogError("Nom de destination invalide: %s", dest.Name)
			return fmt.Errorf("nom de destination invalide: %s", dest.Name)
		}
		if !IsValidPath(dest.Path) {
			LogError("Chemin de destination invalide pour '%s': %s", dest.Name, dest.Path)
			return fmt.Errorf("chemin de destination invalide pour '%s': %s", dest.Name, dest.Path)
		}
	}

	for _, server := range c.RsyncServers {
		if !IsValidName(server.Name) {
			LogError("Nom de serveur rsync invalide: %s", server.Name)
			return fmt.Errorf("nom de serveur rsync invalide: %s", server.Name)
		}
	}
	LogInfo("Configuration validée avec succès.")
	return nil
}

// AddConfigItem ajoute un élément à une slice de configuration de manière générique.
// slicePtr doit être un pointeur vers une slice (ex: *[]BackupConfig).
// item doit être l'élément à ajouter.
// nameField est le nom du champ de l'élément qui contient le nom unique (ex: "Name").
func AddConfigItem(slicePtr interface{}, item interface{}, nameField string) error {
	sliceVal := reflect.ValueOf(slicePtr).Elem()
	itemVal := reflect.ValueOf(item)

	// Vérifier si l'élément existe déjà par son nom
	itemName := itemVal.FieldByName(nameField).String()
	for i := 0; i < sliceVal.Len(); i++ {
		if sliceVal.Index(i).FieldByName(nameField).String() == itemName {
			// Mettre à jour l'élément existant
			sliceVal.Index(i).Set(itemVal)
			return nil // Succès de la mise à jour
		}
	}

	// Ajouter le nouvel élément
	sliceVal.Set(reflect.Append(sliceVal, itemVal))
	return nil
}

// UpdateConfigItem met à jour un élément dans une slice de configuration de manière générique.
// slicePtr doit être un pointeur vers une slice.
// oldItemName est le nom de l'élément à mettre à jour.
// newItem est le nouvel élément.
// nameField est le nom du champ de l'élément qui contient le nom unique.
func UpdateConfigItem(slicePtr interface{}, oldItemName string, newItem interface{}, nameField string) error {
	sliceVal := reflect.ValueOf(slicePtr).Elem()
	newItemVal := reflect.ValueOf(newItem)

	for i := 0; i < sliceVal.Len(); i++ {
		if sliceVal.Index(i).FieldByName(nameField).String() == oldItemName {
			sliceVal.Index(i).Set(newItemVal)
			return nil
		}
	}
	return fmt.Errorf("élément '%s' non trouvé pour mise à jour", oldItemName)
}

// DeleteConfigItem supprime un élément d'une slice de configuration de manière générique.
// slicePtr doit être un pointeur vers une slice.
// itemName est le nom de l'élément à supprimer.
// nameField est le nom du champ de l'élément qui contient le nom unique.
func DeleteConfigItem(slicePtr interface{}, itemName string, nameField string) error {
	sliceVal := reflect.ValueOf(slicePtr).Elem()

	for i := 0; i < sliceVal.Len(); i++ {
		if sliceVal.Index(i).FieldByName(nameField).String() == itemName {
			sliceVal.Set(reflect.AppendSlice(sliceVal.Slice(0, i), sliceVal.Slice(i+1, sliceVal.Len())))
			return nil
		}
	}
	return fmt.Errorf("élément '%s' non trouvé pour suppression", itemName)
}

// GetConfigItem récupère un élément d'une slice de configuration de manière générique.
// slicePtr doit être un pointeur vers une slice.
// itemName est le nom de l'élément à récupérer.
// nameField est le nom du champ de l'élément qui contient le nom unique.
func GetConfigItem(slicePtr interface{}, itemName string, nameField string) (interface{}, bool) {
	sliceVal := reflect.ValueOf(slicePtr).Elem()

	for i := 0; i < sliceVal.Len(); i++ {
		if sliceVal.Index(i).FieldByName(nameField).String() == itemName {
			return sliceVal.Index(i).Interface(), true
		}
	}
	return nil, false
}

