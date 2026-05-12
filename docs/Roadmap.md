# Roadmap de ConducteurOrchestre

## 1. Objectif de la roadmap

- notifications de rappel pour les tâches à faire :
  - [X] Gestion des toast pour les notifications de rappel et d'évaluation des règles
    - Interface de gestionnaire de notification pour envoi de notifications
    - Implementation de notifications pour toast Windows
  - [X] Mise en place d'un worker d'arrière-plan pour les rappels de tâches et l'évaluation périodique des règles
  
- [X] Intégration de la journalisation locale avec SQLite pour suivre les actions de l'utilisateur et les décisions du système
-[ ] ajouter la posibilité de modifier le status d'une tache (en cours, en pause, terminée) et de suivre le temps passé sur chaque tâche pour une meilleure analyse du focus et de la productivité


- UI :  
  - [X] mise en place d'une interface pomodoro pour aider à rythmer les sessions de travail
    - [X] ajouter une selection de timer preset
    - [X] ajouter une selection de tâche inbox pour le pomodoro
  - [X] mise en place d'une interface minimaliste pour visualiser les tâches à venir et les règles actives
  - [ ] mettre a jour l'interface de capture de tâche pour la rendre plus rapide et intuitive
  
- Focus et anti-distraction :
  - [ ] Détection de la dérive de l'attention via des règles basées sur les actions de l'utilisateur
  - [ ] Recommandations de recentrage courtes et actionnables via notifications
  - [ ] Intégration avec le serveur MCP pour ajuster les recommandations en fonction de l'état du focus
