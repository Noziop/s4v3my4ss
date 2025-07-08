package common

import (
	"regexp"
	"strings"
)

// IsValidName valide les noms pour les configurations, sauvegardes, etc.
// Un nom valide ne doit contenir que des caractères alphanumériques, des tirets et des underscores.
func IsValidName(name string) bool {
	// SECURITY: Valider les noms pour éviter l'injection de caractères dangereux.
	if name == "" {
		return true // Autoriser les noms vides si l'utilisateur veut utiliser la valeur par défaut
	}
	re := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return re.MatchString(name)
}

// IsValidPath valide un chemin de fichier ou de répertoire.
// Il vérifie la présence de caractères potentiellement dangereux et de traversées de répertoire.
func IsValidPath(path string) bool {
	// SECURITY: Autoriser les chemins vides pour que l'utilisateur puisse annuler ou utiliser la valeur par défaut.
	if path == "" {
		return true
	}
	// SECURITY: Interdire la traversée de répertoires ("..").
	if strings.Contains(path, "..") {
		// fmt.Printf("%sLa traversée de répertoire ('..') est interdite.%s\n", colorRed, colorReset) // Ne pas afficher ici, c'est une fonction utilitaire
		return false
	}
	// SECURITY: S'assurer que le chemin ne contient pas de caractères d'injection de commande.
	// Autorise les caractères alphanumériques, slashes, points, tirets, underscores, tilde, et deux-points (pour Windows).
	re := regexp.MustCompile(`^[a-zA-Z0-9_/\-.~:]+$`)
	if !re.MatchString(path) {
		// fmt.Printf("%sLe chemin contient des caractères non autorisés.%s\n", colorRed, colorReset) // Ne pas afficher ici
		return false
	}
	return true
}

// IsValidSubnet valide un sous-réseau au format CIDR.
func IsValidSubnet(subnet string) bool {
	if subnet == "" {
		return true // L'utilisateur peut laisser vide pour la valeur par défaut
	}
	// SECURITY: Valider le format CIDR pour éviter les injections dans les outils réseau.
	re := regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}/[0-9]{1,2}$`)
	return re.MatchString(subnet)
}

// IsValidExcludePattern valide un modèle d'exclusion.
func IsValidExcludePattern(pattern string) bool {
	// SECURITY: Empêcher l'injection de commandes via les modèles d'exclusion.
	if strings.ContainsAny(pattern, ";|&`$()<>[]") {
		return false
	}
	return true
}

// IsValidCompressionFormat valide que le format de compression est supporté.
func IsValidCompressionFormat(format string) bool {
	switch format {
	case "targz", "zip": // Utiliser les chaînes littérales car CompressionFormat n'est pas dans ce package
		return true
	default:
		return false
	}
}
