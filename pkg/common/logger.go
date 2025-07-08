package common

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	logFile *os.File
	logger  *log.Logger
	logMutex sync.Mutex
)

// InitLogger initialise le système de journalisation.
// Il crée ou ouvre un fichier de log et configure le logger.
func InitLogger() error {
	logDir, err := GetConfigDir()
	if err != nil {
		return fmt.Errorf("impossible d'obtenir le répertoire de configuration pour les logs: %w", err)
	}

	logFilePath := filepath.Join(logDir, "s4v3my4ss.log")

	// Ouvre le fichier de log en mode append, le crée s'il n'existe pas.
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("impossible d'ouvrir le fichier de log: %w", err)
	}

	logFile = file
	logger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)

	LogInfo("Logger initialisé.")
	return nil
}

// CloseLogger ferme le fichier de log.
func CloseLogger() {
	if logFile != nil {
		LogInfo("Logger fermé.")
		logFile.Close()
	}
}

// LogInfo enregistre un message d'information.
func LogInfo(message string, args ...interface{}) {
	logMutex.Lock()
	defer logMutex.Unlock()
	if logger != nil {
		logger.Printf("INFO: "+message+"\n", args...)
	}
}

// LogWarning enregistre un message d'avertissement.
func LogWarning(message string, args ...interface{}) {
	logMutex.Lock()
	defer logMutex.Unlock()
	if logger != nil {
		logger.Printf("WARNING: "+message+"\n", args...)
	}
}

// LogError enregistre un message d'erreur.
func LogError(message string, args ...interface{}) {
	logMutex.Lock()
	defer logMutex.Unlock()
	if logger != nil {
		logger.Printf("ERROR: "+message+"\n", args...)
	}
}

// LogSecurity enregistre un événement de sécurité critique.
func LogSecurity(message string, args ...interface{}) {
	logMutex.Lock()
	defer logMutex.Unlock()
	if logger != nil {
		logger.Printf("SECURITY: "+message+"\n", args...)
	}
}
