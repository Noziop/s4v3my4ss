# S4v3my4ss

*[🇫🇷 Version française](#français) | 🇬🇧 English version below*

![Version](https://img.shields.io/badge/version-1.0.0-blue)
![Go](https://img.shields.io/badge/go-%3E%3D1.18-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![ASCII Art](https://img.shields.io/badge/ASCII%20Art-patorjk.com-orange)

S4v3my4ss is a robust automatic backup system that monitors and creates backups of your directories in real-time. It offers advanced features like incremental backups, compression, and an easy-to-use interactive interface.

```

_____/\\\\\\\\\\\_______/\\\\\\\\\_____/\\\________/\\\__/\\\\\\\\\\\\\\\____________/\\\\____________/\\\\__/\\\\\\\\\\\\\\\_        
 ___/\\\/////////\\\___/\\\\\\\\\\\\\__\/\\\_______\/\\\_\/\\\///////////____________\/\\\\\\________/\\\\\\_\/\\\///////////__       
  __\//\\\______\///___/\\\/////////\\\_\//\\\______/\\\__\/\\\_______________________\/\\\//\\\____/\\\//\\\_\/\\\_____________      
   ___\////\\\_________\/\\\_______\/\\\__\//\\\____/\\\___\/\\\\\\\\\\\_______________\/\\\\///\\\/\\\/_\/\\\_\/\\\\\\\\\\\_____     
    ______\////\\\______\/\\\\\\\\\\\\\\\___\//\\\__/\\\____\/\\\///////________________\/\\\__\///\\\/___\/\\\_\/\\\///////______    
     _________\////\\\___\/\\\/////////\\\____\//\\\/\\\_____\/\\\_______________________\/\\\____\///_____\/\\\_\/\\\_____________   
      __/\\\______\//\\\__\/\\\_______\/\\\_____\//\\\\\______\/\\\_______________________\/\\\_____________\/\\\_\/\\\_____________  
       _\///\\\\\\\\\\\/___\/\\\_______\/\\\______\//\\\_______\/\\\\\\\\\\\\\\\___________\/\\\_____________\/\\\_\/\\\\\\\\\\\\\\\_ 
        ___\///////////_____\///________\///________\///________\///////////////____________\///______________\///__\///////////////__

  
  Protect your important files with S4v3my4ss!
```

## Features

- **Real-time monitoring** of file changes (via inotify/fswatch)
- **Incremental backups** to save disk space
- **Automatic compression** of backups (tar.gz or zip)
- **Interactive CLI interface** for easy usage
- **Non-interactive mode** for script integration
- **Smart dependency management** that adapts to available tools
- **Cross-platform** : works on most Linux distributions and WSL

## Installation

### Prerequisites

S4v3my4ss is written in Go and requires Go 1.18 or higher. External dependencies are automatically installed if needed (with your permission).

Main dependencies:
- `rsync` for backups
- `inotify-tools` or `fswatch` for monitoring
- `tar` and `gzip` or `zip` for compression

### Installation from source

```
# Clone the repository
git clone https://github.com/Noziop/s4v3my4ss.git
cd s4v3my4ss

# Compile and install to ~/.local/bin
make install

# Or install globally (requires root permissions)
sudo make install-global
```

### Installation via Go

```
go install github.com/Noziop/s4v3my4ss/cmd/saveme@latest
```


### Go install troubleshooting

After `go install`, if you get "command not found":

```
# The binary is installed but not in PATH
# Add Go bin directory to your PATH:
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc

# Or for zsh users:
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc

# Test the installation:
saveme --help
```

**Alternative**: Use the full path directly:
```
$(go env GOPATH)/bin/saveme --help
```


## Usage

### Interactive mode

Launch S4v3my4ss without arguments to enter interactive mode:

```
saveme
```

You'll see an interface like this:

```
_____/\\\\\\\\\\\_______/\\\\\\\\\_____/\\\________/\\\__/\\\\\\\\\\\\\\\____________/\\\\____________/\\\\__/\\\\\\\\\\\\\\\_        
 ___/\\\/////////\\\___/\\\\\\\\\\\\\__\/\\\_______\/\\\_\/\\\///////////____________\/\\\\\\________/\\\\\\_\/\\\///////////__       
  __\//\\\______\///___/\\\/////////\\\_\//\\\______/\\\__\/\\\_______________________\/\\\//\\\____/\\\//\\\_\/\\\_____________      
   ___\____\\\_________\/\\\_______\/\\\__\//\\\____/\\\___\/\\\\\\\\\\\_______________\/\\\\///\\\/\\\/_\/\\\_\/\\\\\\\\\\\_____     
    ______\////\\\______\/\\\\\\\\\\\\\\\___\//\\\__/\\\____\/\\\///////________________\/\\\__\///\\\/___\/\\\_\/\\\///////______    
     _________\////\\\___\/\\\/////////\\\____\//\\\/\\\_____\/\\\_______________________\/\\\____\///_____\/\\\_\/\\\_____________   
      __/\\\______\//\\\__\/\\\_______\/\\\_____\//\\\\\______\/\\\_______________________\/\\\_____________\/\\\_\/\\\_____________  
       _\///\\\\\\\\\\\/___\/\\\_______\/\\\______\//\\\_______\/\\\\\\\\\\\\\\\___________\/\\\_____________\/\\\_\/\\\\\\\\\\\\\\\_ 
        ___\///////////_____\///________\///________\///________\///////////////____________\///______________\///__\///////////////__

Automatic Backup and Restore System

MAIN MENU
  1. Configure a new backup
  2. Start monitoring a directory
  3. Restore a backup
  4. Manage existing backups
  5. Check/install dependencies
  0. Exit
```

#### Interactive mode guide

1. **Configure a new backup** 
   - Choose option 1 to create a new backup configuration
   - You'll need to provide:
     - Backup name (unique identifier)
     - Path of directory to monitor/backup
     - Enable compression (recommended to save space)
     - Directories to exclude (comma-separated)
     - Files to exclude (you can use wildcards like *.tmp)
     - Interval in minutes (0 to disable periodic monitoring)

2. **Start monitoring a directory**
   - Choose option 2 to start monitoring an already configured directory
   - Select the configuration to use from the displayed list
   - Monitoring will start and display detected events
   - Backups will be triggered automatically when changes are detected
   - Press Ctrl+C to stop monitoring

3. **Restore a backup**
   - Choose option 3 to restore a previous backup
   - Select the backup to restore from the list
   - Enter destination path (or leave empty to restore to original location)
   - Confirm operation if destination directory already exists

4. **Manage existing backups**
   - Choose option 4 to access the management submenu
   - You can:
     - List all backups with their details
     - Delete a specific backup
     - Clean old backups according to configured retention policy

5. **Check/install dependencies**
   - Choose option 5 to check if all required dependencies are installed
   - The tool can automatically install missing dependencies with your permission

### Command line mode

You can also use S4v3my4ss from command line:

```
# Monitor a directory with existing configuration
saveme watch 

# Restore a backup
saveme restore  [destination_path]

# Manage backups
saveme manage list              # List backups
saveme manage delete        # Delete a backup
saveme manage clean             # Clean according to retention policy

# Show help
saveme --help
```

## Configuration

S4v3my4ss stores its configuration in `~/.config/s4v3my4ss/config.json`. You can edit it directly or via the interactive interface.

Configuration example:

```
{
  "backupDirectories": [
    {
      "sourcePath": "/path/to/my/documents",
      "name": "Documents",
      "compression": true,
      "excludeDirs": ["tmp", "cache"],
      "excludeFiles": ["*.tmp", "*.log"],
      "interval": 0
    }
  ],
  "backupDestination": "/path/to/backups",
  "retentionPolicy": {
    "keepDaily": 7,
    "keepWeekly": 4,
    "keepMonthly": 3
  }
}
```

### Backup structure

Backups are stored in the directory defined by `backupDestination` with the following structure:
- Each backup has a unique ID based on name, date and hash
- Metadata is stored in `~/.config/s4v3my4ss/backups/[ID].json`
- Incremental backups use hard links to save space
- Compressed backups are stored in .tar.gz format

## Troubleshooting

### Common issues

1. **"Command not found" error**
   - Make sure the installation directory is in your PATH
   - Try using the full path: ~/.local/bin/saveme

2. **Changes are not detected**
   - Check that inotify-tools or fswatch is properly installed
   - Increase system monitoring limit: `echo fs.inotify.max_user_watches=524288 | sudo tee -a /etc/sysctl.conf && sudo sysctl -p`

3. **Restore failed**
   - Check destination directory permissions
   - If backup is compressed, make sure you have enough temporary disk space

4. **Slow performance**
   - Avoid monitoring directories containing many volatile files
   - Use exclusions to ignore temporary files and caches
   - Adjust minimum interval between backups in source code

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](../CONTRIBUTING.md) for contribution guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.

## Acknowledgments

- Inspired by tools like rsnapshot, timeshift and restic
- ASCII art logo created specially for this project, thanks to: [patorjk.com Text to ASCII generator](http://patorjk.com/software/taag/)

---

## Français

*[🇬🇧 English version](#s4v3my4ss) | 🇫🇷 Version française ci-dessous*

S4v3my4ss est un système de sauvegarde automatique robuste qui surveille et crée des sauvegardes de vos répertoires en temps réel. Il offre des fonctionnalités avancées comme les sauvegardes incrémentielles, la compression, et une interface interactive facile à utiliser.


  Protégez vos fichiers importants avec S4v3my4ss!


## Fonctionnalités

- **Surveillance en temps réel** des modifications de fichiers (via inotify/fswatch)
- **Sauvegardes incrémentielles** pour économiser de l'espace disque
- **Compression automatique** des sauvegardes (tar.gz ou zip)
- **Interface CLI interactive** pour une utilisation facile
- **Mode non-interactif** pour l'intégration dans des scripts
- **Gestion des dépendances intelligente** qui s'adapte aux outils disponibles
- **Multi-plateforme** : fonctionne sur la plupart des distributions Linux et WSL

## Installation

### Prérequis

S4v3my4ss est écrit en Go et nécessite Go 1.18 ou supérieur. Les dépendances externes sont installées automatiquement si nécessaire (avec votre permission).

Dépendances principales :
- `rsync` pour les sauvegardes
- `inotify-tools` ou `fswatch` pour la surveillance
- `tar` et `gzip` ou `zip` pour la compression

### Installation depuis les sources

```
# Cloner le dépôt
git clone https://github.com/Noziop/s4v3my4ss.git
cd s4v3my4ss

# Compiler et installer dans ~/.local/bin
make install

# Ou pour installer globalement (nécessite les permissions root)
sudo make install-global
```

### Installation via Go

```
go install github.com/Noziop/s4v3my4ss/cmd/saveme@latest
```

### Post-installation 'go install'

Après installation, si vous lancez la cmd `saveme`, vous tomberez sur le message  "command not found"

```
# Le binaire est bien installé, mais pas dans PATH
# Ajoutez Go bin dans votre PATH:
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc

# ou pour les utilisateurs zsh :
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc

# testez l'installation:
saveme --help
```

**Alternative**: utilisation du chemin absolu:
```
$(go env GOPATH)/bin/saveme --help
```

## Utilisation

### Mode interactif

Lancez S4v3my4ss sans arguments pour entrer dans le mode interactif :

```
saveme
```

Vous verrez une interface comme celle-ci :

```
Système de Sauvegarde et Restauration Automatique

MENU PRINCIPAL
  1. Configurer une nouvelle sauvegarde
  2. Démarrer la surveillance d'un répertoire
  3. Restaurer une sauvegarde
  4. Gérer les sauvegardes existantes
  5. Vérifier/installer les dépendances
  0. Quitter
```

#### Guide du mode interactif

1. **Configurer une nouvelle sauvegarde** 
   - Choisissez l'option 1 pour créer une nouvelle configuration de sauvegarde
   - Vous devrez fournir :
     - Nom de la sauvegarde (identifiant unique)
     - Chemin du répertoire à surveiller/sauvegarder
     - Activation de la compression (recommandée pour économiser de l'espace)
     - Répertoires à exclure (séparés par des virgules)
     - Fichiers à exclure (vous pouvez utiliser des wildcards comme *.tmp)
     - Intervalle en minutes (0 pour désactiver la surveillance périodique)

2. **Démarrer la surveillance d'un répertoire**
   - Choisissez l'option 2 pour commencer à surveiller un répertoire déjà configuré
   - Sélectionnez la configuration à utiliser parmi la liste affichée
   - La surveillance démarrera et affichera les événements détectés
   - Les sauvegardes seront déclenchées automatiquement lorsque des modifications sont détectées
   - Appuyez sur Ctrl+C pour arrêter la surveillance

3. **Restaurer une sauvegarde**
   - Choisissez l'option 3 pour restaurer une sauvegarde précédente
   - Sélectionnez la sauvegarde à restaurer dans la liste
   - Entrez le chemin de destination (ou laissez vide pour restaurer à l'emplacement d'origine)
   - Confirmez l'opération si le répertoire de destination existe déjà

4. **Gérer les sauvegardes existantes**
   - Choisissez l'option 4 pour accéder au sous-menu de gestion
   - Vous pourrez :
     - Lister toutes les sauvegardes avec leurs détails
     - Supprimer une sauvegarde spécifique
     - Nettoyer les anciennes sauvegardes selon la politique de rétention configurée

5. **Vérifier/installer les dépendances**
   - Choisissez l'option 5 pour vérifier si toutes les dépendances requises sont installées
   - L'outil peut installer automatiquement les dépendances manquantes avec votre permission

### Mode ligne de commande

Vous pouvez également utiliser S4v3my4ss en ligne de commande :

```
# Surveiller un répertoire avec une configuration existante
saveme watch 

# Restaurer une sauvegarde
saveme restore  [chemin_destination]

# Gérer les sauvegardes
saveme manage list              # Lister les sauvegardes
saveme manage delete        # Supprimer une sauvegarde
saveme manage clean             # Nettoyer selon politique de rétention

# Afficher l'aide
saveme --help
```

## Configuration

S4v3my4ss stocke sa configuration dans `~/.config/s4v3my4ss/config.json`. Vous pouvez la modifier directement ou via l'interface interactive.

Exemple de configuration :

```
{
  "backupDirectories": [
    {
      "sourcePath": "/chemin/vers/mes/documents",
      "name": "Documents",
      "compression": true,
      "excludeDirs": ["tmp", "cache"],
      "excludeFiles": ["*.tmp", "*.log"],
      "interval": 0
    }
  ],
  "backupDestination": "/chemin/vers/sauvegardes",
  "retentionPolicy": {
    "keepDaily": 7,
    "keepWeekly": 4,
    "keepMonthly": 3
  }
}
```

### Structure des sauvegardes

Les sauvegardes sont stockées dans le répertoire défini par `backupDestination` avec la structure suivante :
- Chaque sauvegarde a un ID unique basé sur le nom, la date et un hash
- Les métadonnées sont stockées dans `~/.config/s4v3my4ss/backups/[ID].json`
- Les sauvegardes incrémentielles utilisent des liens durs pour économiser de l'espace
- Les sauvegardes compressées sont stockées au format .tar.gz

## Dépannage

### Problèmes courants

1. **Erreur "Commande non trouvée"**
   - Assurez-vous que le répertoire d'installation est dans votre PATH
   - Essayez d'utiliser le chemin complet : ~/.local/bin/saveme

2. **Les modifications ne sont pas détectées**
   - Vérifiez que inotify-tools ou fswatch est correctement installé
   - Augmentez la limite de surveillance système : `echo fs.inotify.max_user_watches=524288 | sudo tee -a /etc/sysctl.conf && sudo sysctl -p`

3. **Restauration échouée**
   - Vérifiez les permissions du répertoire de destination
   - Si la sauvegarde est compressée, assurez-vous d'avoir assez d'espace disque temporaire

4. **Performances lentes**
   - Évitez de surveiller des répertoires contenant de nombreux fichiers volatils
   - Utilisez les exclusions pour ignorer les fichiers temporaires et les caches
   - Ajustez l'intervalle minimal entre les sauvegardes dans le code source

## Contribuer

Les contributions sont les bienvenues ! Consultez [CONTRIBUTING.md](../CONTRIBUTING.md) pour les directives de contribution.

## Licence

Ce projet est sous licence MIT - voir le fichier [LICENSE](../LICENSE) pour plus de détails.

## Remerciements

- Inspiré par des outils comme rsnapshot, timeshift et restic
- Logo ASCII art créé spécialement pour ce projet, grâce à : [patorjk.com Text to ASCII generator](http://patorjk.com/software/taag/)
