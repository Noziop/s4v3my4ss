package backup

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Noziop/s4v3my4ss/internal/ui/display"
	"github.com/Noziop/s4v3my4ss/internal/ui/input"
	"github.com/Noziop/s4v3my4ss/pkg/common"
)

// ManageBackupsInteractive permet de gérer les sauvegardes
func ManageBackupsInteractive() {
	common.LogInfo("Début de la gestion interactive des sauvegardes.")
	for {
		display.ClearScreen()
		fmt.Printf("%sGestion des sauvegardes%s\n\n", display.ColorBold(), display.ColorReset())

		fmt.Printf("  %s1.%s Lister les sauvegardes\n", display.ColorGreen(), display.ColorReset())
		fmt.Printf("  %s2.%s Supprimer une sauvegarde\n", display.ColorGreen(), display.ColorReset())
		fmt.Printf("  %s3.%s Nettoyer les anciennes sauvegardes\n", display.ColorGreen(), display.ColorReset())
		fmt.Printf("  %s0.%s Retour au menu principal\n", display.ColorGreen(), display.ColorReset())

		choice := input.ReadInput("Votre choix: ")

		switch choice {
		case "1":
			ListBackups()
		case "2":
			DeleteBackupInteractive()
		case "3":
			cleanOldBackups()
		case "0":
			common.LogInfo("Retour au menu principal depuis la gestion des sauvegardes.")
			return
		default:
			common.LogWarning("Option de gestion des sauvegardes non valide: %s", choice)
			fmt.Println("Option non valide. Veuillez réessayer.")
		}

		input.ReadInput("Appuyez sur Entrée pour continuer...")
	}
}

// ListBackups affiche la liste des sauvegardes
func ListBackups() {
	common.LogInfo("Liste des sauvegardes demandée.")
	backups, err := common.ListBackups()
	if err != nil {
		input.DisplayMessage(true, "Erreur lors de la récupération de la liste des sauvegardes: %v", err)
		return
	}

	if len(backups) == 0 {
		input.DisplayMessage(false, "Aucune sauvegarde disponible.")
		return	}

	fmt.Printf("%-20s %-30s %-20s %-10s %-8s\n", "NOM", "CHEMIN SOURCE", "DATE", "TAILLE", "TYPE")
	fmt.Println(strings.Repeat("-", 100))

	for _, b := range backups {
		timeStr := b.Time.Format("02/01/2006 15:04")
		sizeStr := display.FormatSize(b.Size)
		typeStr := "Normal"
		if b.IsIncremental {
			typeStr = "Incr."
		}
		if b.Compression {
			typeStr += " (C)"
		}

		fmt.Printf("%-20s %-30s %-20s %-10s %-8s\n",
			display.TruncateString(b.Name, 20),
			display.TruncateString(b.SourcePath, 30),
			timeStr,
			sizeStr,
			typeStr)
	}
	common.LogInfo("Liste des %d sauvegardes affichée.", len(backups))
}

// DeleteBackupInteractive permet de supprimer une sauvegarde
func DeleteBackupInteractive() {
	common.LogInfo("Début de la suppression interactive de sauvegarde.")
	backups, err := common.ListBackups()
	if err != nil {
		input.DisplayMessage(true, "Erreur lors de la récupération des sauvegardes pour suppression: %v", err)
		return
	}

	if len(backups) == 0 {
		input.DisplayMessage(false, "Aucune sauvegarde disponible à supprimer.")
		return
	}

	fmt.Println("Sauvegardes disponibles:")
	for i, b := range backups {
		timeStr := b.Time.Format("02/01/2006 15:04:05")
		fmt.Printf("%d. %s (%s) - %s\n", i+1, b.Name, b.SourcePath, timeStr)
	}

	idxStr := input.ReadInput("Sélectionnez une sauvegarde à supprimer (numéro): ")
	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 1 || idx > len(backups) {
		input.DisplayMessage(true, "Choix invalide.")
		return
	}

	backup := backups[idx-1]

	if !input.ConfirmAction(fmt.Sprintf("Êtes-vous sûr de vouloir supprimer '%s'?", backup.Name)) {
			common.LogInfo("Suppression annulée par l'utilisateur pour la sauvegarde: %s.", backup.Name)
			fmt.Println("Suppression annulée.")
			return
		}

	DeleteBackup(backup.ID)
	common.LogInfo("Demande de suppression de la sauvegarde %s.", backup.ID)
}

// DeleteBackup supprime une sauvegarde
func DeleteBackup(id string) {
	common.LogInfo("Tentative de suppression de la sauvegarde avec ID: %s", id)
	fmt.Printf("Suppression de la sauvegarde %s...\n", id)

	err := common.DeleteBackup(id)
	if err != nil {
		input.DisplayMessage(true, "Erreur lors de la suppression: %v", err)
		return
	}

	input.DisplayMessage(false, "Suppression terminée avec succès.")
}

// cleanOldBackups nettoie les anciennes sauvegardes selon la politique de rétention
func cleanOldBackups() {
	common.LogInfo("Début du nettoyage des anciennes sauvegardes.")
	fmt.Println("Nettoyage des anciennes sauvegardes...")
	// TODO: Implémenter la logique de nettoyage des anciennes sauvegardes ici.
	// Cette fonctionnalité est un placeholder et doit être développée ultérieurement.
	common.LogInfo("Nettoyage des anciennes sauvegardes terminé.")
	input.DisplayMessage(false, "Nettoyage terminé.")
}
