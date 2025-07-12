package backup

import (
	"time"

	"github.com/Noziop/s4v3my4ss/pkg/common"
)

// cleanupOldBackups nettoie les anciennes sauvegardes selon la politique de rétention.
// Cette fonction est exécutée en arrière-plan après chaque sauvegarde.
func cleanupOldBackups(name string) {
	common.LogInfo("Démarrage du nettoyage des anciennes sauvegardes pour '%s'...", name)
	defer common.LogInfo("Nettoyage des anciennes sauvegardes pour '%s' terminé.", name)

	allBackups, err := common.ListBackups()
	if err != nil {
		common.LogError("cleanupOldBackups: impossible de lister les sauvegardes: %v", err)
		return
	}

	var relevantBackups []common.BackupInfo
	for _, b := range allBackups {
		if b.Name == name {
			relevantBackups = append(relevantBackups, b)
		}
	}

	if len(relevantBackups) == 0 {
		common.LogInfo("Aucune sauvegarde trouvée pour '%s'.", name)
		return
	}

	// Trier les sauvegardes par date, les plus anciennes en premier
	// Sort by time, oldest first
	for i := 0; i < len(relevantBackups); i++ {
		for j := i + 1; j < len(relevantBackups); j++ {
			if relevantBackups[i].Time.After(relevantBackups[j].Time) {
				relevantBackups[i], relevantBackups[j] = relevantBackups[j], relevantBackups[i]
			}
		}
	}

	policy := common.AppConfig.RetentionPolicy

	// Nettoyage quotidien
	dailyKept := cleanupByInterval(relevantBackups, policy.KeepDaily, common.Daily)
	// Nettoyage hebdomadaire
	weeklyKept := cleanupByInterval(dailyKept, policy.KeepWeekly, common.Weekly)
	// Nettoyage mensuel
	monthlyKept := cleanupByInterval(weeklyKept, policy.KeepMonthly, common.Monthly)

	// Supprimer les sauvegardes qui ne sont pas conservées par la politique
	toDelete := make(map[string]struct{})
	for _, b := range relevantBackups {
		toDelete[b.ID] = struct{}{}
	}

	for _, b := range dailyKept {
		delete(toDelete, b.ID)
	}
	for _, b := range weeklyKept {
		delete(toDelete, b.ID)
	}
	for _, b := range monthlyKept {
		delete(toDelete, b.ID)
	}

	for id := range toDelete {
		if err := common.DeleteBackup(id); err != nil {
			common.LogError("cleanupOldBackups: impossible de supprimer la sauvegarde %s: %v", id, err)
		} else {
			common.LogSecurity("Sauvegarde %s supprimée par la politique de rétention.", id)
		}
	}
}

// cleanupByInterval filtre les sauvegardes pour ne garder que celles qui respectent la politique de rétention pour un intervalle donné
func cleanupByInterval(backups []common.BackupInfo, keep int, interval common.RetentionInterval) []common.BackupInfo {
	if keep <= 0 {
		return []common.BackupInfo{} // Ne rien garder si la politique est 0 ou moins
	}

	var keptBackups []common.BackupInfo
	lastKeptTime := time.Time{} // Initialiser à l'heure zéro

	for _, b := range backups {
		shouldKeep := false
		switch interval {
		case common.Daily:
			// Garder si c'est la première sauvegarde du jour ou si le jour est différent
			if lastKeptTime.IsZero() || b.Time.YearDay() != lastKeptTime.YearDay() || b.Time.Year() != lastKeptTime.Year() {
				shouldKeep = true
			}
		case common.Weekly:
			// Garder si c'est la première sauvegarde de la semaine ou si la semaine est différente
			year1, week1 := b.Time.ISOWeek()
			year2, week2 := lastKeptTime.ISOWeek()
			if lastKeptTime.IsZero() || week1 != week2 || year1 != year2 {
				shouldKeep = true
			}
		case common.Monthly:
			// Garder si c'est la première sauvegarde du mois ou si le mois est différent
			if lastKeptTime.IsZero() || b.Time.Month() != lastKeptTime.Month() || b.Time.Year() != lastKeptTime.Year() {
				shouldKeep = true
			}
		}

		if shouldKeep {
			keptBackups = append(keptBackups, b)
			lastKeptTime = b.Time
		}
	}

	// Si on a plus de sauvegardes que la politique "keep", on supprime les plus anciennes
	if len(keptBackups) > keep {
		common.LogInfo("Suppression de %d sauvegardes excédentaires pour l'intervalle %s.", len(keptBackups)-keep, interval)
		return keptBackups[len(keptBackups)-keep:]
	}

	return keptBackups
}
