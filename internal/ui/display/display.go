package display

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/Noziop/s4v3my4ss/pkg/common"
)

// ClearScreen efface le contenu du terminal.
func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

// DisplayHeader affiche l'en-tête de l'application.
func DisplayHeader() {
	// Couleurs du rainbow flag pour la pride (6 couleurs exactement)
	prideColors := []string{
		"\033[38;5;196m", // Rouge (en haut)
		"\033[38;5;208m", // Orange
		"\033[38;5;226m", // Jaune
		"\033[38;5;46m",  // Vert
		"\033[38;5;27m",  // Bleu
		"\033[38;5;129m", // Violet (en bas)
	}

	// Essayer de lire le fichier banner.txt
	bannerPath := "banner.txt"
	if _, err := os.Stat(bannerPath); os.IsNotExist(err) {
		// Essayer dans le répertoire de l'exécutable
		execPath, err := os.Executable()
		if err == nil {
			bannerPath = filepath.Join(filepath.Dir(execPath), "..", "banner.txt")
		}
	}

	// Lire le fichier banner s'il existe
	if bannerContent, err := os.ReadFile(bannerPath); err == nil {
		common.LogInfo("Fichier banner.txt lu avec succès depuis %s.", bannerPath)
		// Convertir le contenu en string et le diviser en lignes
		bannerLines := strings.Split(string(bannerContent), "\n")

		// Affichage avec couleurs du drapeau pride (1 couleur pour 2 lignes)
		fmt.Print("\033[1m") // Texte en gras
		for i, line := range bannerLines {
			// Déterminer l'index de couleur (une couleur pour deux lignes)
			colorIndex := (i / 2) % len(prideColors)

			// Afficher la ligne avec la couleur correspondante
			fmt.Print(prideColors[colorIndex], line, "\n")
		}
		fmt.Print("\033[0m") // Réinitialiser la couleur
	} else {
		common.LogError("Impossible de lire le fichier banner.txt: %v. Utilisation de la bannière de secours.", err)
		// Bannière de secours si le fichier n'est pas trouvé
		fmt.Print("\033[1m")
		fmt.Println("  ___ _ _  _               __  __  ___ ")
		fmt.Println(" / __| | \\| |  ___  /\\ /\\ /__\\/__\\|_  )")
		fmt.Println(" \\__ \\ | .` | / -_)/ _  //_\\ / \\/ / / / ")
		fmt.Println(" |___/_|_|\\|_| \\___|\\__,_/\\__/\\__//___| ")
		fmt.Print("\033[0m")
	}

	
}

// DisplayMainMenu affiche le menu principal.
func DisplayMainMenu() {
	fmt.Printf("%sMENU PRINCIPAL%s\n", ColorBold(), ColorReset())
	fmt.Printf("  %s1.%s Configurer une nouvelle sauvegarde\n", ColorGreen(), ColorReset())
	fmt.Printf("  %s2.%s Démarrer la surveillance d'un répertoire\n", ColorGreen(), ColorReset())
	fmt.Printf("  %s3.%s Restaurer une sauvegarde\n", ColorGreen(), ColorReset())
	fmt.Printf("  %s4.%s Gérer les sauvegardes existantes\n", ColorGreen(), ColorReset())
	fmt.Printf("  %s5.%s Vérifier/installer les dépendances\n", ColorGreen(), ColorReset())
	fmt.Printf("  %s6.%s Rechercher des serveurs rsync sur le réseau\n", ColorGreen(), ColorReset())
	fmt.Printf("  %s7.%s Gérer la configuration\n", ColorGreen(), ColorReset())
	fmt.Printf("  %s0.%s Quitter\n", ColorGreen(), ColorReset())
	fmt.Println()
}

// TruncateString tronque une chaîne si elle dépasse une certaine longueur.
func TruncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}

// FormatSize formate la taille en une chaîne lisible par l'homme.
func FormatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.1f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.1f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.1f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func ColorBold() string {
	return "\033[1m"
}
func ColorReset() string {
	return "\033[0m"
}
func ColorGreen() string {
	return "\033[0;32m"
}
func ColorRed() string {
	return "\033[0;31m"
}
func ColorYellow() string {
	return "\033[0;33m"
}
func ColorBlue() string {
	return "\033[0;34m"
}
func ColorMagenta() string {
	return "\033[0;35m"
}
func ColorCyan() string {
	return "\033[0;36m"
}
func ColorWhite() string {
	return "\033[0;37m"
}

// DisplayConfigList affiche une liste d'éléments de configuration de manière générique.
// items est la slice d'éléments à afficher.
// header est le titre de la liste.
// itemDisplayFunc est une fonction de rappel qui prend l'index et l'élément, et retourne la chaîne à afficher pour cet élément.
func DisplayConfigList(items interface{}, header string, itemDisplayFunc func(index int, item interface{}) string) {
	val := reflect.ValueOf(items)
	if val.Kind() != reflect.Slice {
		common.LogError("DisplayConfigList attend une slice, a reçu %s", val.Kind())
		return
	}

	fmt.Printf("%s%s:%s\n", ColorBold(), header, ColorReset())

	if val.Len() == 0 {
		fmt.Printf("%sAucun élément configuré.%s\n\n", ColorYellow(), ColorReset())
		return
	}

	for i := 0; i < val.Len(); i++ {
		fmt.Println(itemDisplayFunc(i, val.Index(i).Interface()))
	}
	fmt.Println()
}
