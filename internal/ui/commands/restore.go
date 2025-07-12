package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Noziop/s4v3my4ss/internal/restore"
	"github.com/Noziop/s4v3my4ss/internal/ui/display"
	"github.com/Noziop/s4v3my4ss/internal/ui/input"
	"github.com/Noziop/s4v3my4ss/pkg/common"
)

// RestoreBackupInteractive permet de restaurer une sauvegarde
func RestoreBackupInteractive(isCLI bool) {
	common.LogInfo("Début de la restauration interactive.")
	if !isCLI {
		fmt.Printf("%sRestauration d'une sauvegarde%s\n\n", display.ColorBold(), display.ColorReset())
	}

	backups, err := common.ListBackups()
	if err != nil {
		common.LogError("Erreur lors de la récupération des sauvegardes pour restauration: %v", err)
		fmt.Printf("%sErreur lors de la récupération des sauvegardes: %v%s\n", display.ColorRed(), err, display.ColorReset())
		return
	}

	if len(backups) == 0 {
		common.LogWarning("Aucune sauvegarde disponible pour la restauration.")
		fmt.Printf("%sAucune sauvegarde disponible.%s\n", display.ColorYellow(), display.ColorReset())
		return
	}

	fmt.Println("Sauvegardes disponibles:")
	for i, b := range backups {
		timeStr := b.Time.Format("02/01/2006 15:04:05")
		fmt.Printf("%d. %s (%s) - %s\n", i+1, b.Name, b.SourcePath, timeStr)
	}

	idxStr := input.ReadInput("Sélectionnez une sauvegarde (numéro): ")
	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 1 || idx > len(backups) {
		common.LogError("Choix de sauvegarde invalide pour la restauration: %s", idxStr)
		fmt.Printf("%sChoix invalide.%s\n", display.ColorRed, display.ColorReset)
		return
	}

	backup := backups[idx-1]

	targetPath := input.ReadAndValidateInput("Chemin de destination (vide pour restaurer à l'emplacement d'origine): ", common.IsValidPath, "Chemin invalide ou non sécurisé.") // Utilisation de common.IsValidPath
	if targetPath == "" {
		targetPath = backup.SourcePath
	}

	// SECURITY: Vérifier si le chemin de destination est autorisé
	if !common.AppConfig.Security.IsPathAllowed(targetPath) {
		common.LogSecurity("Tentative de restauration vers un chemin non autorisé: %s", targetPath)
		fmt.Printf("%sErreur: Le chemin de destination '%s' n'est pas autorisé dans la configuration de sécurité.%s\n", display.ColorRed, targetPath, display.ColorReset)
		return
	}

	// Vérifier la destination
	if common.DirExists(targetPath) {
		common.LogWarning("Répertoire de destination '%s' existe déjà. Demande de confirmation pour écrasement.", targetPath)
		overwriteStr := input.ReadInput(fmt.Sprintf("Le répertoire '%s' existe déjà. Écraser? (o/n): ", targetPath))
		if strings.ToLower(overwriteStr) != "o" {
			common.LogInfo("Restauration annulée par l'utilisateur.")
			fmt.Println("Restauration annulée.")
			return
		}
	}

	common.LogInfo("Restauration de la sauvegarde '%s' vers '%s'.", backup.Name, targetPath)
	fmt.Printf("Restauration de la sauvegarde '%s' vers '%s'...\n", backup.Name, targetPath)

	if err := restore.RestoreBackup(backup.ID, targetPath); err != nil {
		common.LogError("Erreur lors de la restauration de %s: %v", backup.Name, err)
		fmt.Printf("%sErreur lors de la restauration: %v%s\n", display.ColorRed, err, display.ColorReset)
		return
	}

	common.LogInfo("Restauration terminée avec succès pour %s.", backup.Name)
	fmt.Printf("%sRestauration terminée avec succès.%s\n", display.ColorGreen(), display.ColorReset())
}