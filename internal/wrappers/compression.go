package wrappers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Noziop/s4v3my4ss/pkg/common"
)

// CompressionFormat représente le format de compression utilisé
type CompressionFormat string

const (
	// FormatTarGz représente le format tar.gz
	FormatTarGz CompressionFormat = "targz"
	// FormatZip représente le format zip
	FormatZip CompressionFormat = "zip"
)

// CompressionWrapper gère la compression et décompression des fichiers
type CompressionWrapper struct {
	// Vérifié indique si les outils de compression sont disponibles
	Verified bool
	// DefaultFormat est le format de compression par défaut
	DefaultFormat CompressionFormat
}

// NewCompressionWrapper crée une nouvelle instance de CompressionWrapper
func NewCompressionWrapper() (*CompressionWrapper, error) {
	cw := &CompressionWrapper{
		Verified:     false,
		DefaultFormat: FormatTarGz,
	}

	// Vérifier si tar est disponible
	if !common.IsCommandAvailable("tar") {
		common.LogError("tar n'est pas installé.")
		return cw, fmt.Errorf("tar n'est pas installé")
	}

	// Vérifier si gzip est disponible
	if !common.IsCommandAvailable("gzip") {
		// Si gzip n'est pas disponible mais que zip l'est, utiliser zip comme format par défaut
		if common.IsCommandAvailable("zip") {
			cw.DefaultFormat = FormatZip
		} else {
			common.LogError("Ni gzip ni zip ne sont installés.")
			return cw, fmt.Errorf("ni gzip ni zip ne sont installés")
		}
	}

	cw.Verified = true
	common.LogInfo("CompressionWrapper créé et outils vérifiés.")
	return cw, nil
}

// EnsureAvailable vérifie que les outils de compression sont disponibles
func (cw *CompressionWrapper) EnsureAvailable() error {
	if cw.Verified {
		return nil
	}

	// Essayer d'installer tar et gzip
	if common.EnsureDependency("tar", "tar") != nil || 
	   common.EnsureDependency("gzip", "gzip") != nil {
		// Si l'installation échoue, essayer zip
		common.LogWarning("Impossible d'installer tar ou gzip. Tentative d'installation de zip.")
		if common.EnsureDependency("zip", "zip") == nil && 
		   common.EnsureDependency("unzip", "unzip") == nil {
			cw.DefaultFormat = FormatZip
			common.LogInfo("zip et unzip installés avec succès. Format par défaut défini sur zip.")
		} else {
			common.LogError("Impossible d'installer les outils de compression.")
			return fmt.Errorf("impossible d'installer les outils de compression")
		}
	} else {
		cw.DefaultFormat = FormatTarGz
		common.LogInfo("tar et gzip installés avec succès. Format par défaut défini sur tar.gz.")
	}

	cw.Verified = true
	return nil
}

// Compress compresse un répertoire dans un fichier de manière sécurisée.
func (cw *CompressionWrapper) Compress(sourcePath, destPath string, format CompressionFormat) error {
	common.LogInfo("Début de la compression: Source=%s, Destination=%s, Format=%s", sourcePath, destPath, format)
	// SECURITY: Valider les chemins et le format avant toute opération.
	if !common.IsValidPath(sourcePath) || !common.IsValidPath(destPath) { // Utilisation de common.IsValidPath
		common.LogError("Chemin source ou destination invalide ou non sécurisé: Source=%s, Dest=%s", sourcePath, destPath)
		return fmt.Errorf("chemin source ou destination invalide ou non sécurisé")
	}
	if !common.IsValidCompressionFormat(string(format)) { // Utilisation de common.IsValidCompressionFormat
		common.LogError("Format de compression non supporté: %s", format)
		return fmt.Errorf("format de compression non supporté: %s", format)
	}

	if err := cw.EnsureAvailable(); err != nil {
		common.LogError("Outils de compression non disponibles: %v", err)
		return err
	}

	if format == "" {
		format = cw.DefaultFormat
	}

	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		common.LogError("Impossible de créer le répertoire de destination %s: %v", destDir, err)
		return fmt.Errorf("impossible de créer le répertoire de destination: %w", err)
	}

	var cmd *exec.Cmd

	switch format {
	case FormatTarGz:
		// SECURITY: Les arguments sont passés séparément pour éviter l'injection de shell.
		cmd = exec.Command("tar", "-czf", destPath, "-C", filepath.Dir(sourcePath), filepath.Base(sourcePath))
	case FormatZip:
		// Le changement de répertoire est une opération sensible. Assurer que les chemins sont propres.
		cleanSourceDir := filepath.Clean(filepath.Dir(sourcePath))
		cmd = exec.Command("zip", "-r", destPath, filepath.Base(sourcePath))
		cmd.Dir = cleanSourceDir // Exécuter la commande dans le répertoire parent de la source
	default:
		common.LogError("Format de compression non supporté: %s", format)
		return fmt.Errorf("format de compression non supporté: %s", format)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	common.LogInfo("Exécution de la commande de compression: %s %s", cmd.Path, strings.Join(cmd.Args, " "))
	err := cmd.Run()
	if err != nil {
		common.LogError("Erreur lors de la compression: %v", err)
		return fmt.Errorf("erreur lors de la compression: %w", err)
	}

	common.LogInfo("Compression terminée avec succès.")
	return nil
}

// Decompress décompresse un fichier vers un répertoire de manière sécurisée.
func (cw *CompressionWrapper) Decompress(sourcePath, destPath string) error {
	common.LogInfo("Début de la décompression: Source=%s, Destination=%s", sourcePath, destPath)
	// SECURITY: Valider les chemins avant la décompression.
	if !common.IsValidPath(sourcePath) || !common.IsValidPath(destPath) { // Utilisation de common.IsValidPath
		common.LogError("Chemin source ou destination invalide ou non sécurisé: Source=%s, Dest=%s", sourcePath, destPath)
		return fmt.Errorf("chemin source ou destination invalide ou non sécurisé")
	}

	if err := cw.EnsureAvailable(); err != nil {
		common.LogError("Outils de décompression non disponibles: %v", err)
		return err
	}

	if err := os.MkdirAll(destPath, 0755); err != nil {
		common.LogError("Impossible de créer le répertoire de destination %s: %v", destPath, err)
		return fmt.Errorf("impossible de créer le répertoire de destination: %w", err)
	}

	var cmd *exec.Cmd

	if strings.HasSuffix(sourcePath, ".tar.gz") || strings.HasSuffix(sourcePath, ".tgz") {
		// SECURITY: Utiliser des arguments séparés pour tar.
		cmd = exec.Command("tar", "-xzf", sourcePath, "-C", destPath)
	} else if strings.HasSuffix(sourcePath, ".zip") {
		// SECURITY: Utiliser des arguments séparés pour unzip.
		cmd = exec.Command("unzip", sourcePath, "-d", destPath)
	} else {
		common.LogError("Format de compression non reconnu pour le fichier: %s", sourcePath)
		return fmt.Errorf("format de compression non reconnu pour le fichier: %s", sourcePath)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	common.LogInfo("Exécution de la commande de décompression: %s %s", cmd.Path, strings.Join(cmd.Args, " "))
	err := cmd.Run()
	if err != nil {
		common.LogError("Erreur lors de la décompression: %v", err)
		return fmt.Errorf("erreur lors de la décompression: %w", err)
	}

	common.LogInfo("Décompression terminée avec succès.")
	return nil
}