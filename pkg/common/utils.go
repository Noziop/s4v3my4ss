package common

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Couleurs pour le terminal
const (
	ColorRed    = "\033[0;31m"
	ColorGreen  = "\033[0;32m"
	ColorYellow = "\033[0;33m"
	ColorBlue   = "\033[0;34m"
	ColorBold   = "\033[1m"
	ColorReset  = "\033[0m"
)

// PrintColored affiche un texte coloré dans le terminal
func PrintColored(color, text string) {
	fmt.Printf("%s%s%s", color, text, ColorReset)
}

// PrintColoredLine affiche une ligne de texte coloré dans le terminal
func PrintColoredLine(color, text string) {
	fmt.Printf("%s%s%s\n", color, text, ColorReset)
}

// IsCommandAvailable vérifie si une commande est disponible dans le système
func IsCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// GetConfigDir renvoie le répertoire de configuration de l'application
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	configDir := filepath.Join(homeDir, ".config", "s4v3my4ss")
	err = os.MkdirAll(configDir, 0755)
	return configDir, err
}

// GetTempDir crée et renvoie un répertoire temporaire pour l'application
func GetTempDir() (string, error) {
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("s4v3my4ss-%d", os.Getpid()))
	err := os.MkdirAll(tempDir, 0755)
	return tempDir, err
}

// DirExists vérifie si un répertoire existe
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// FileExists vérifie si un fichier existe
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// GetPackageManager renvoie le gestionnaire de paquets approprié pour le système
func GetPackageManager() (string, []string, bool) {
	if runtime.GOOS != "linux" {
		return "", nil, false
	}

	// Vérifier les gestionnaires de paquets courants
	if IsCommandAvailable("apt-get") {
		return "apt-get", []string{"install", "-y"}, true
	} else if IsCommandAvailable("dnf") {
		return "dnf", []string{"install", "-y"}, true
	} else if IsCommandAvailable("pacman") {
		return "pacman", []string{"-S", "--noconfirm"}, true
	} else if IsCommandAvailable("zypper") {
		return "zypper", []string{"install", "-y"}, true
	} else if IsCommandAvailable("apk") {
		return "apk", []string{"add"}, true
	}
	
	return "", nil, false
}

// EnsureDependency vérifie si une dépendance est installée et tente de l'installer
func EnsureDependency(command, packageName string) error {
	if IsCommandAvailable(command) {
		return nil
	}

	fmt.Printf("La commande %s n'est pas installée. Tentative d'installation...\n", command)

	mgr, args, ok := GetPackageManager()
	if !ok {
		return fmt.Errorf("impossible de détecter un gestionnaire de paquets compatible")
	}

	fullArgs := append(args, packageName)
	cmd := exec.Command("sudo", append([]string{mgr}, fullArgs...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// PadRight ajoute des espaces à droite d'une chaîne pour atteindre une longueur spécifique
func PadRight(str string, length int) string {
	if len(str) >= length {
		return str
	}
	return str + strings.Repeat(" ", length-len(str))
}

// GenerateRandomString génère une chaîne aléatoire de la longueur spécifiée
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}