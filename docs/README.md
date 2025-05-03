# S4v3my4ss

![Version](https://img.shields.io/badge/version-1.0.0-blue)
![Go](https://img.shields.io/badge/go-%3E%3D1.18-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![ASCII Art](https://img.shields.io/badge/ASCII%20Art-patorjk.com-orange)

S4v3my4ss est un système de sauvegarde automatique robuste qui surveille et crée des sauvegardes de vos répertoires en temps réel. Il offre des fonctionnalités avancées comme les sauvegardes incrémentielles, la compression, et une interface interactive facile à utiliser.

<div align="center">
<pre>
_____/\\\\\\\\\\\_______/\\\\\\\\\_____/\\\________/\\\__/\\\\\\\\\\\\\\\____________/\\\\____________/\\\\__/\\\\\\\\\\\\\\\_        
 ___/\\\/////////\\\___/\\\\\\\\\\\\\__\/\\\_______\/\\\_\/\\\///////////____________\/\\\\\\________/\\\\\\_\/\\\///////////__       
  __\//\\\______\///___/\\\/////////\\\_\//\\\______/\\\__\/\\\_______________________\/\\\//\\\____/\\\//\\\_\/\\\_____________      
   ___\////\\\_________\/\\\_______\/\\\__\//\\\____/\\\___\/\\\\\\\\\\\_______________\/\\\\///\\\/\\\/_\/\\\_\/\\\\\\\\\\\_____     
    ______\////\\\______\/\\\\\\\\\\\\\\\___\//\\\__/\\\____\/\\\///////________________\/\\\__\///\\\/___\/\\\_\/\\\///////______    
     _________\////\\\___\/\\\/////////\\\____\//\\\/\\\_____\/\\\_______________________\/\\\____\///_____\/\\\_\/\\\_____________   
      __/\\\______\//\\\__\/\\\_______\/\\\_____\//\\\\\______\/\\\_______________________\/\\\_____________\/\\\_\/\\\_____________  
       _\///\\\\\\\\\\\/___\/\\\_______\/\\\______\//\\\_______\/\\\\\\\\\\\\\\\___________\/\\\_____________\/\\\_\/\\\\\\\\\\\\\\\_ 
        ___\///////////_____\///________\///________\///________\///////////////____________\///______________\///__\///////////////__
</pre>
  <br>
  <i>Protégez vos fichiers importants avec S4v3my4ss!</i>
</div>

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

```bash
# Cloner le dépôt
git clone https://github.com/Noziop/s4v3my4ss.git
cd s4v3my4ss/Projet\ Go

# Compiler et installer dans ~/.local/bin
make install

# Ou pour installer globalement (nécessite les permissions root)
sudo make install-global
```

### Installation via Go

```bash
go install github.com/Noziop/s4v3my4ss/Projet\ Go/cmd/saveme@latest
```

## Utilisation

### Mode interactif

Lancez S4v3my4ss sans arguments pour entrer dans le mode interactif :

```bash
saveme
```

Vous verrez une interface comme celle-ci :

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

```bash
# Surveiller un répertoire avec une configuration existante
saveme watch <nom_configuration>

# Restaurer une sauvegarde
saveme restore <id_sauvegarde> [chemin_destination]

# Gérer les sauvegardes
saveme manage list              # Lister les sauvegardes
saveme manage delete <id>       # Supprimer une sauvegarde
saveme manage clean             # Nettoyer selon politique de rétention

# Afficher l'aide
saveme --help
```

## Configuration

S4v3my4ss stocke sa configuration dans `~/.config/s4v3my4ss/config.json`. Vous pouvez la modifier directement ou via l'interface interactive.

Exemple de configuration :

```json
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
- Logo ASCII art créé spécialement pour ce projet, grace à : [patorjk.com Text to ASCII generator](http://patorjk.com/software/taag/)