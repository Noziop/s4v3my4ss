package watch

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/Noziop/s4v3my4ss/internal/backup"
	"github.com/Noziop/s4v3my4ss/internal/wrappers"
	"github.com/Noziop/s4v3my4ss/pkg/common"
)

// Constantes pour la configuration du watcher
const (
	// Le délai minimum entre deux sauvegardes (pour éviter les sauvegardes trop fréquentes)
	minBackupInterval = 10 * time.Second
	// Le délai d'attente après des modifications avant de déclencher une sauvegarde
	waitAfterChanges = 5 * time.Second
)

// Watcher gère la surveillance d'un répertoire
type Watcher struct {
	// Config contient la configuration du répertoire à surveiller
	Config common.BackupConfig
	// Le wrapper inotify pour la détection des modifications
	inotify *wrappers.InotifyWrapper
	// Canal pour signaler qu'une sauvegarde doit être déclenchée
	triggerBackup chan struct{}
	// Contexte pour arrêter la surveillance
	ctx context.Context
	// Fonction pour annuler la surveillance
	cancel context.CancelFunc
	// Mutex pour protéger l'accès à lastBackupTime
	mu sync.Mutex
	// Date de la dernière sauvegarde
	lastBackupTime time.Time
	// Indique si une modification est en attente
	pendingChanges bool
	// Timer pour déclencher une sauvegarde après un délai
	backupTimer *time.Timer
}

// NewWatcher crée un nouveau watcher pour un répertoire
func NewWatcher(config common.BackupConfig) (*Watcher, error) {
	inotify, err := wrappers.NewInotifyWrapper()
	if err != nil {
		return nil, fmt.Errorf("impossible de créer le watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Watcher{
		Config:        config,
		inotify:       inotify,
		triggerBackup: make(chan struct{}, 1),
		ctx:           ctx,
		cancel:        cancel,
		lastBackupTime: time.Time{}, // Zéro = jamais fait de sauvegarde
	}, nil
}

// StartWatch démarre la surveillance du répertoire
func StartWatch(config common.BackupConfig) error {
	// Vérifier que le répertoire à surveiller existe
	if !common.DirExists(config.SourcePath) {
		return fmt.Errorf("le répertoire à surveiller n'existe pas: %s", config.SourcePath)
	}

	// Créer un nouveau watcher
	watcher, err := NewWatcher(config)
	if err != nil {
		return err
	}

	// Démarrer la surveillance
	return watcher.Start()
}

// StartWatchWithCallback démarre la surveillance du répertoire avec un canal pour l'arrêter
func StartWatchWithCallback(config common.BackupConfig, done <-chan bool) error {
	// Vérifier que le répertoire à surveiller existe
	if !common.DirExists(config.SourcePath) {
		return fmt.Errorf("le répertoire à surveiller n'existe pas: %s", config.SourcePath)
	}

	// Créer un nouveau watcher
	watcher, err := NewWatcher(config)
	if err != nil {
		return err
	}

	// Créer une goroutine qui attendra le signal d'arrêt
	go func() {
		// Attendre le signal d'arrêt
		<-done
		// Arrêter la surveillance
		watcher.Stop()
		fmt.Println("Surveillance arrêtée.")
	}()

	// Démarrer la surveillance
	return watcher.Start()
}

// Start démarre la surveillance du répertoire
func (w *Watcher) Start() error {
	fmt.Printf("Démarrage de la surveillance du répertoire: %s\n", w.Config.SourcePath)

	// Effectuer une sauvegarde initiale
	if err := w.performBackup(); err != nil {
		fmt.Printf("Erreur lors de la sauvegarde initiale: %v\n", err)
		// On continue même en cas d'erreur
	}

	// Démarrer la goroutine qui gère les sauvegardes
	go w.backupManager()

	// Démarrer la surveillance des fichiers
	return w.inotify.WatchDirectory(w.ctx, w.Config.SourcePath, true, w.handleFileChange)
}

// Stop arrête la surveillance
func (w *Watcher) Stop() {
	w.cancel()
	// Arrêter le timer s'il est actif
	if w.backupTimer != nil {
		w.backupTimer.Stop()
	}
}

// handleFileChange est appelé chaque fois qu'un fichier est modifié
func (w *Watcher) handleFileChange(event wrappers.WatchEvent) {
	// Ignorer les fichiers temporaires et cachés
	baseName := filepath.Base(event.Path)
	if len(baseName) > 0 && (baseName[0] == '.' || baseName[0] == '~' || baseName[len(baseName)-1] == '~') {
		return
	}
	
	// Vérifier si le fichier fait partie des exclusions
	for _, exclude := range w.Config.ExcludeFiles {
		matched, err := filepath.Match(exclude, baseName)
		if err == nil && matched {
			return
		}
	}
	
	// Vérifier si le répertoire fait partie des exclusions
	dirPath := filepath.Dir(event.Path)
	for _, exclude := range w.Config.ExcludeDirs {
		if filepath.Base(dirPath) == exclude {
			return
		}
	}

	// Afficher l'événement
	fmt.Printf("Modification détectée: %s (%s)\n", event.Path, event.EventType)

	w.mu.Lock()
	defer w.mu.Unlock()

	// Marquer qu'il y a des modifications en attente
	w.pendingChanges = true

	// Si un timer est déjà actif, le réinitialiser
	if w.backupTimer != nil {
		w.backupTimer.Stop()
	}

	// Créer un nouveau timer pour déclencher la sauvegarde après un délai
	w.backupTimer = time.AfterFunc(waitAfterChanges, func() {
		// Ne déclencher que si assez de temps s'est écoulé depuis la dernière sauvegarde
		w.mu.Lock()
		defer w.mu.Unlock()
		
		if time.Since(w.lastBackupTime) >= minBackupInterval {
			// Envoyer un signal pour déclencher la sauvegarde
			select {
			case w.triggerBackup <- struct{}{}:
				// Signal envoyé avec succès
			default:
				// Canal déjà plein, une sauvegarde est déjà prévue
			}
		}
	})
}

// backupManager gère les sauvegardes en fonction des signaux reçus
func (w *Watcher) backupManager() {
	for {
		select {
		case <-w.ctx.Done():
			// Contexte annulé, arrêter la surveillance
			return
		case <-w.triggerBackup:
			// Recevoir un signal pour déclencher une sauvegarde
			err := w.performBackup()
			if err != nil {
				fmt.Printf("Erreur lors de la sauvegarde: %v\n", err)
			}
		}
	}
}

// performBackup effectue une sauvegarde du répertoire
func (w *Watcher) performBackup() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	if !w.pendingChanges && !w.lastBackupTime.IsZero() {
		// Pas de modifications depuis la dernière sauvegarde
		return nil
	}

	fmt.Printf("Démarrage de la sauvegarde de %s...\n", w.Config.SourcePath)
	
	// Appeler le module de backup pour créer une sauvegarde
	backupConfig := backup.BackupConfig{
		SourcePath:   w.Config.SourcePath,
		Name:         w.Config.Name,
		Compression:  w.Config.Compression,
		ExcludeDirs:  w.Config.ExcludeDirs,
		ExcludeFiles: w.Config.ExcludeFiles,
		Incremental:  true, // Forcer les sauvegardes incrémentales pour la surveillance automatique
	}
	
	err := backup.CreateBackup(backupConfig)
	if err != nil {
		return err
	}
	
	// Mettre à jour la date de la dernière sauvegarde
	w.lastBackupTime = time.Now()
	w.pendingChanges = false
	
	fmt.Printf("Sauvegarde terminée avec succès à %s.\n", w.lastBackupTime.Format("15:04:05"))
	return nil
}