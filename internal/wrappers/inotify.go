package wrappers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/Noziop/s4v3my4ss/pkg/common"
)

// WatchEvent représente un événement de modification de fichier
type WatchEvent struct {
	// Path est le chemin du fichier modifié
	Path string
	// EventType est le type d'événement (create, modify, delete, etc.)
	EventType string
	// IsDir indique si c'est un répertoire
	IsDir bool
}

// WatchCallback est la signature de la fonction de rappel appelée quand un fichier est modifié
type WatchCallback func(event WatchEvent)

// InotifyWrapper gère la surveillance des fichiers via inotify ou fswatch
type InotifyWrapper struct {
	// Vérifié indique si inotify/fswatch est disponible
	Verified bool
	// UseInotify indique si on utilise inotify (true) ou fswatch (false)
	UseInotify bool
}

// NewInotifyWrapper crée une nouvelle instance de InotifyWrapper
func NewInotifyWrapper() (*InotifyWrapper, error) {
	iw := &InotifyWrapper{
		Verified:   false,
		UseInotify: false,
	}

	// Vérifier si inotifywait est disponible
	if common.IsCommandAvailable("inotifywait") {
		iw.Verified = true
		iw.UseInotify = true
		return iw, nil
	}

	// Sinon, vérifier si fswatch est disponible
	if common.IsCommandAvailable("fswatch") {
		iw.Verified = true
		iw.UseInotify = false
		return iw, nil
	}

	return iw, fmt.Errorf("ni inotifywait ni fswatch ne sont installés")
}

// EnsureAvailable vérifie que inotify/fswatch est disponible, et tente de l'installer si ce n'est pas le cas
func (iw *InotifyWrapper) EnsureAvailable() error {
	if iw.Verified {
		return nil
	}

	// Essayer d'installer inotify-tools d'abord (plus léger et plus commun)
	if common.EnsureDependency("inotifywait", "inotify-tools") == nil {
		iw.Verified = true
		iw.UseInotify = true
		return nil
	}

	// Si inotify n'a pas pu être installé, essayer fswatch
	if common.EnsureDependency("fswatch", "fswatch") == nil {
		iw.Verified = true
		iw.UseInotify = false
		return nil
	}

	return fmt.Errorf("impossible d'installer inotify-tools ou fswatch")
}

// WatchDirectory surveille un répertoire pour les changements et appelle callback pour chaque événement
func (iw *InotifyWrapper) WatchDirectory(ctx context.Context, directory string, recursive bool, callback WatchCallback) error {
	if err := iw.EnsureAvailable(); err != nil {
		return err
	}

	var cmd *exec.Cmd
	var stdout io.ReadCloser
	var err error

	if iw.UseInotify {
		// Utiliser inotifywait
		args := []string{
			"-m",       // Mode surveillance continue
			"-q",       // Mode silencieux
			"--format", "%w%f %e", // Format de sortie: <chemin> <événement>
			"-e", "create,modify,delete,move", // Types d'événements à surveiller
		}

		if recursive {
			args = append(args, "-r") // Mode récursif
		}

		args = append(args, directory) // Répertoire à surveiller
		cmd = exec.CommandContext(ctx, "inotifywait", args...)
	} else {
		// Utiliser fswatch
		args := []string{
			"-0",       // Terminer chaque événement par un caractère NULL
			"--format", "%p %f", // Format: <chemin> <type>
		}

		if recursive {
			args = append(args, "-r") // Mode récursif
		}

		args = append(args, directory) // Répertoire à surveiller
		cmd = exec.CommandContext(ctx, "fswatch", args...)
	}

	// Obtenir la sortie standard
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("erreur lors de la création du pipe: %w", err)
	}

	// Démarrer la commande
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("erreur lors du démarrage de la surveillance: %w", err)
	}

	// Lancer une goroutine pour lire la sortie
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			event := parseWatchEvent(line, iw.UseInotify)
			callback(event)
		}
	}()

	// Attendre que la commande se termine ou que le contexte soit annulé
	go func() {
		<-ctx.Done()
		cmd.Process.Kill()
	}()

	return cmd.Wait()
}

// parseWatchEvent convertit une ligne de sortie d'inotify/fswatch en un WatchEvent
func parseWatchEvent(line string, isInotify bool) WatchEvent {
	var event WatchEvent

	if isInotify {
		// Format inotify: "<chemin> <événements>"
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			event.Path = parts[0]
			event.EventType = parts[1]
			// Vérifier si c'est un répertoire d'après les événements
			event.IsDir = strings.Contains(parts[1], "ISDIR")
		}
	} else {
		// Format fswatch: "<chemin> <type>"
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			event.Path = parts[0]
			// Convertir les types d'événements fswatch en format cohérent
			switch parts[1] {
			case "Created":
				event.EventType = "CREATE"
			case "Updated", "OwnerModified", "AttributeModified":
				event.EventType = "MODIFY"
			case "Removed":
				event.EventType = "DELETE"
			case "Renamed":
				event.EventType = "MOVE"
			}
			// fswatch n'indique pas directement si c'est un répertoire dans la sortie
			// On pourrait implémenter une vérification supplémentaire si nécessaire
		}
	}

	return event
}