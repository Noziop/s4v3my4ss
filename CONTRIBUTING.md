# Guide de contribution

Merci de votre intérêt pour contribuer à S4v3my4ss ! Voici quelques directives pour vous aider à participer au projet.

## Comment contribuer

1. **Forker** le dépôt sur GitHub
2. **Cloner** votre fork en local
3. **Créer une branche** pour vos modifications
4. **Faire vos modifications** en suivant les conventions de code
5. **Tester** vos modifications
6. **Commiter** vos changements avec un message clair
7. **Pousser** votre branche vers votre fork
8. Créer une **Pull Request** vers notre branche principale

## Conventions de code

- Suivre les conventions de Go (gofmt, golint)
- Utiliser des commentaires explicites pour les fonctions exportées
- Écrire des tests unitaires pour les nouvelles fonctionnalités
- Documenter les changements d'API dans le README ou la documentation

## Structure du projet

```
s4v3my4ss/
├── cmd/               # Points d'entrée de l'application
├── internal/          # Code privé
│   ├── backup/        # Gestion des sauvegardes
│   ├── restore/       # Restauration
│   ├── ui/            # Interface utilisateur
│   ├── watch/         # Surveillance de fichiers
│   └── wrappers/      # Wrappers pour outils externes
├── pkg/               # Packages publics réutilisables
└── docs/              # Documentation
```

## Soumettre une issue

Si vous trouvez un bug ou avez une suggestion d'amélioration :

1. Vérifiez d'abord que l'issue n'existe pas déjà
2. Utilisez le modèle d'issue approprié
3. Soyez aussi précis que possible, incluez :
   - Étapes pour reproduire (pour les bugs)
   - Comportement attendu vs observé
   - Version utilisée
   - Environnement (OS, version de Go)

## Pull Requests

- Liez votre PR à une issue existante quand c'est pertinent
- Décrivez clairement les changements apportés
- Incluez des captures d'écran pour les changements d'interface
- Assurez-vous que tous les tests passent
- Mettez à jour la documentation si nécessaire

## Revue de code

- Soyez respectueux et constructif dans vos commentaires
- Concentrez-vous sur le code, pas sur la personne
- Expliquez vos suggestions et soyez ouvert à la discussion

## Questions?

Si vous avez des questions, n'hésitez pas à ouvrir une issue avec le tag "question" ou à contacter l'équipe de mainteneurs.

Merci pour votre contribution !