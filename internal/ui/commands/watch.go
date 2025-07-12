package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Noziop/s4v3my4ss/internal/ui/display"
	"github.com/Noziop/s4v3my4ss/internal/ui/input"
	"github.com/Noziop/s4v3my4ss/internal/watch"
	"github.com/Noziop/s4v3my4ss/pkg/common"
)

// WatchDirectoryInteractive permet de démarrer la surveillance d'un répertoire
func WatchDirectoryInteractive() {
	common.LogInfo("Début de la surveillance interactive.")
	fmt.Printf("%sSurveillance d'un répertoire%s\n\n", display.ColorBold(), display.ColorReset())
	
	// Afficher les configurations disponibles
	configs := common.AppConfig.BackupDirs
	if len(configs) == 0 {
		common.LogWarning("Aucune configuration de sauvegarde disponible pour la surveillance.")
		fmt.Printf("%sAucune configuration de sauvegarde n'est est disponible.%s\n", display.ColorYellow(), display.ColorReset())
		fmt.Println("Veuillez d'abord créer une configuration.")
		return
	}
	
	fmt.Println("Configurations disponibles:")
	for i, cfg := range configs {
		fmt.Printf("%d. %s (%s)\n", i+1, cfg.Name, cfg.SourcePath)
	}
	
	// Demander quelle configuration utiliser
	idxStr := input.ReadInput("Sélectionnez une configuration (numéro): ")
	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 1 || idx > len(configs) {
		common.LogError("Choix de configuration invalide pour la surveillance: %s", idxStr)
		fmt.Printf("%sChoix invalide.%s\n", display.ColorRed(), display.ColorReset())
		return
	}
	
	config := configs[idx-1]
	common.LogInfo("Démarrage de la surveillance pour la configuration: %s.", config.Name)
	fmt.Printf("Démarrage de la surveillance du répertoire: %s\n", config.SourcePath)
	
	// Option pour la durée de surveillance
	durationStr := input.ReadInput("Durée de surveillance en minutes (0 pour mode continu, Ctrl+C pour arrêter): ")
	duration := 0
	if dur, err := strconv.Atoi(durationStr); err == nil && dur > 0 {
		duration = dur
	} else {
		common.LogWarning("Durée de surveillance invalide ou continue sélectionnée: %s", durationStr)
	}
	
	if duration > 0 {
		common.LogInfo("Surveillance pendant %d minutes pour %s.", duration, config.Name)
		fmt.Printf("Surveillance pendant %d minutes...\n", duration)
		
		// Créer un canal pour communiquer avec la goroutine de surveillance
		done := make(chan bool)
		
		// Lancer la surveillance dans une goroutine
		go func() {
			if err := watch.StartWatchWithCallback(config, done); err != nil {
				common.LogError("Erreur lors de la surveillance avec callback pour %s: %v", config.Name, err)
				fmt.Printf("%sErreur lors de la surveillance: %v%s\n", display.ColorRed, err, display.ColorReset)
			}
		}()
		
		// Attendre la durée spécifiée
		time.Sleep(time.Duration(duration) * time.Minute)
		
		// Signaler l'arrêt
		done <- true
		
		common.LogInfo("Surveillance terminée après %d minutes pour %s.", duration, config.Name)
		fmt.Printf("\n%sSurveillance terminée après %d minutes.%s\n", display.ColorGreen(), duration, display.ColorReset())
	} else {
		common.LogInfo("Surveillance en mode continu pour %s.", config.Name)
		// Mode continu avec prompt pour retourner au menu
		fmt.Println("Mode surveillance continue. Appuyez sur Ctrl+C pour arrêter.")
		fmt.Println("Les sauvegardes continueront même si vous quittez ce prompt.")
		
		// Lancer la surveillance en arrière-plan
		go watch.StartWatch(config)
		
		// Attendre que l'utilisateur appuie sur Entrée pour revenir au menu
		input.ReadInput("\nAppuyez sur Entrée pour revenir au menu principal (la surveillance continuera en arrière-plan)... ")
	}
}