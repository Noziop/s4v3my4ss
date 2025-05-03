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
		return cw, fmt.Errorf("tar n'est pas installé")
	}

	// Vérifier si gzip est disponible
	if !common.IsCommandAvailable("gzip") {
		// Si gzip n'est pas disponible mais que zip l'est, utiliser zip comme format par défaut
		if common.IsCommandAvailable("zip") {
			cw.DefaultFormat = FormatZip
		} else {
			return cw, fmt.Errorf("ni gzip ni zip ne sont installés")
		}
	}

	cw.Verified = true
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
		if common.EnsureDependency("zip", "zip") == nil && 
		   common.EnsureDependency("unzip", "unzip") == nil {
			cw.DefaultFormat = FormatZip
		} else {
			return fmt.Errorf("impossible d'installer les outils de compression")
		}
	} else {
		cw.DefaultFormat = FormatTarGz
	}

	cw.Verified = true
	return nil
}

// Compress compresse un répertoire dans un fichier
func (cw *CompressionWrapper) Compress(sourcePath, destPath string, format CompressionFormat) error {
	if err := cw.EnsureAvailable(); err != nil {
		return err
	}

	// Si format n'est pas spécifié, utiliser le format par défaut
	if format == "" {
		format = cw.DefaultFormat
	}

	// Assurer que le répertoire parent de destination existe
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("impossible de créer le répertoire de destination: %w", err)
	}

	// Composer la commande selon le format
	var cmd *exec.Cmd

	switch format {
	case FormatTarGz:
		// Utiliser tar avec gzip
		cmd = exec.Command("tar", "-czf", destPath, "-C", filepath.Dir(sourcePath), filepath.Base(sourcePath))
	case FormatZip:
		// Utiliser zip
		// Se déplacer dans le répertoire parent pour éviter d'inclure le chemin complet
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("impossible d'obtenir le répertoire courant: %w", err)
		}
		defer os.Chdir(currentDir)
		
		if err := os.Chdir(filepath.Dir(sourcePath)); err != nil {
			return fmt.Errorf("impossible de changer de répertoire: %w", err)
		}
		
		cmd = exec.Command("zip", "-r", destPath, filepath.Base(sourcePath))
	default:
		return fmt.Errorf("format de compression non supporté: %s", format)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Decompress décompresse un fichier vers un répertoire
func (cw *CompressionWrapper) Decompress(sourcePath, destPath string) error {
	if err := cw.EnsureAvailable(); err != nil {
		return err
	}

	// Assurer que le répertoire de destination existe
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("impossible de créer le répertoire de destination: %w", err)
	}

	// Déterminer le format d'après l'extension
	var cmd *exec.Cmd

	if strings.HasSuffix(sourcePath, ".tar.gz") || strings.HasSuffix(sourcePath, ".tgz") {
		// Décompresser avec tar
		cmd = exec.Command("tar", "-xzf", sourcePath, "-C", destPath)
	} else if strings.HasSuffix(sourcePath, ".zip") {
		// Décompresser avec unzip
		cmd = exec.Command("unzip", sourcePath, "-d", destPath)
	} else {
		return fmt.Errorf("format de compression non reconnu pour le fichier: %s", sourcePath)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}