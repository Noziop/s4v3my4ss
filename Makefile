# Variables
APPNAME = saveme
VERSION = 1.0.0
GOFILES = $(shell find . -name "*.go")

# Chemins d'installation
LOCAL_INSTALL_PATH = $(HOME)/.local/bin
GLOBAL_INSTALL_PATH = /usr/local/bin

# Couleurs pour les messages
GREEN = \033[0;32m
YELLOW = \033[0;33m
NC = \033[0m # No Color

.PHONY: all build clean install install-global test help

# Cible par défaut
all: build

# Compilation de l'application
build:
	@echo "$(GREEN)Compilation de $(APPNAME) v$(VERSION)...$(NC)"
	@go build -o bin/$(APPNAME) ./cmd/$(APPNAME)
	@echo "$(GREEN)Compilation terminée. Exécutable disponible dans bin/$(APPNAME)$(NC)"

# Suppression des fichiers générés
clean:
	@echo "$(GREEN)Nettoyage...$(NC)"
	@rm -rf bin/
	@go clean
	@echo "$(GREEN)Nettoyage terminé.$(NC)"

# Installation locale (dans ~/.local/bin)
install: build
	@echo "$(GREEN)Installation de $(APPNAME) dans $(LOCAL_INSTALL_PATH)...$(NC)"
	@mkdir -p $(LOCAL_INSTALL_PATH)
	@cp bin/$(APPNAME) $(LOCAL_INSTALL_PATH)/
	@chmod +x $(LOCAL_INSTALL_PATH)/$(APPNAME)
	@echo "$(GREEN)Installation terminée.$(NC)"
	@echo "$(YELLOW)Assurez-vous que $(LOCAL_INSTALL_PATH) est dans votre PATH.$(NC)"

# Installation globale (nécessite les permissions root)
install-global: build
	@echo "$(GREEN)Installation de $(APPNAME) dans $(GLOBAL_INSTALL_PATH)...$(NC)"
	@mkdir -p $(GLOBAL_INSTALL_PATH)
	@cp bin/$(APPNAME) $(GLOBAL_INSTALL_PATH)/
	@chmod +x $(GLOBAL_INSTALL_PATH)/$(APPNAME)
	@echo "$(GREEN)Installation terminée.$(NC)"

# Exécution des tests
test:
	@echo "$(GREEN)Exécution des tests...$(NC)"
	@go test ./...
	@echo "$(GREEN)Tests terminés.$(NC)"

# Affichage de l'aide
help:
	@echo "$(GREEN)Make targets for $(APPNAME):$(NC)"
	@echo "  make build         - Compile l'application"
	@echo "  make clean         - Supprime les fichiers générés"
	@echo "  make install       - Installe l'application dans ~/.local/bin"
	@echo "  make install-global - Installe l'application dans /usr/local/bin (nécessite sudo)"
	@echo "  make test          - Exécute les tests"
	@echo "  make help          - Affiche cette aide"