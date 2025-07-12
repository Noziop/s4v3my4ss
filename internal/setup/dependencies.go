package setup

import (
	"fmt"
	"strings"

	"github.com/Noziop/s4v3my4ss/internal/ui/display"
	"github.com/Noziop/s4v3my4ss/internal/ui/input"
	"github.com/Noziop/s4v3my4ss/pkg/common"
)

// CheckDependencies vérifie et installe les dépendances nécessaires
func CheckDependencies() {
	common.LogInfo("Vérification des dépendances.")
	fmt.Printf("%sVérification des dépendances%s\n\n", display.ColorBold(), display.ColorReset())

	deps := []struct {
		command string
		pkg     string
		desc    string
	}{
		{"rsync", "rsync", "Outil de synchronisation de fichiers"},
		{"inotifywait", "inotify-tools", "Surveillance des fichiers et répertoires"},
		{"jq", "jq", "Traitement JSON"},
		{"tar", "tar", "Compression et archivage"},
	}

	for _, dep := range deps {
		fmt.Printf("Vérification de %s... ", dep.command)

		if common.IsCommandAvailable(dep.command) {
			common.LogInfo("Dépendance %s: OK.", dep.command)
			fmt.Printf("%sOK%s\n", display.ColorGreen(), display.ColorReset())
			continue
		}

		common.LogWarning("Dépendance %s: Non trouvée.", dep.command)
		fmt.Printf("%sNon trouvé%s\n", display.ColorYellow(), display.ColorReset())
		installStr := input.ReadInput(fmt.Sprintf("Installer %s (%s)? (o/n): ", dep.pkg, dep.desc))

		if strings.ToLower(installStr) != "o" {
			common.LogInfo("Installation de %s ignorée par l'utilisateur.", dep.pkg)
			fmt.Println("Installation ignorée.")
			continue
		}

		common.LogInfo("Installation de %s...", dep.pkg)
		fmt.Printf("Installation de %s...\n", dep.pkg)
		if err := common.EnsureDependency(dep.command, dep.pkg); err != nil {
			common.LogError("Erreur lors de l'installation de %s: %v", dep.pkg, err)
			fmt.Printf("%sErreur lors de l'installation: %v%s\n", display.ColorRed(), err, display.ColorReset())
		} else {
			common.LogInfo("%s installé avec succès.", dep.pkg)
			fmt.Printf("%s%s installé avec succès.%s\n", display.ColorGreen(), dep.pkg, display.ColorReset())
		}
	}
	common.LogInfo("Vérification des dépendances terminée.")
}