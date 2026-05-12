# CLI ConducteurOrchestre

Le build Windows fournit un binaire unique `conducteur-orchestre.exe` qui sait :

1. démarrer le serveur local ;
2. interroger l'état de focus ;
3. piloter l'inbox et le timer en ligne de commande ;
4. ouvrir une interface `Terminal.Gui` locale dans le terminal.

Le CLI reste volontairement court, local et déterministe.

## Préparer le build

```powershell
powershell -ExecutionPolicy Bypass -File build\publish.ps1
```

Le binaire publié est créé dans `build\dist\win-x64\conducteur-orchestre.exe`.


## Commandes disponibles

| Commande | Effet |
| --- | --- |
| `conducteur-orchestre focus snapshot` | Affiche timer, intervention éventuelle et top 3 tâches |
| `conducteur-orchestre` | Ouvre l'interface `Terminal.Gui` locale et démarre le serveur si l'URL locale n'est pas joignable |
| `conducteur-orchestre timer presets` | Liste au maximum 2 presets de timer |
| `conducteur-orchestre timer start --preset "Pomodoro Court" --line 2` | Lance un timer sur le serveur local et associe la tâche si la ligne est fournie |
| `conducteur-orchestre timer stop` | Arrête le timer actif |
| `conducteur-orchestre timer stop --aborted` | Arrête le timer sans le marquer comme complété |
| `conducteur-orchestre timer status` | Affiche le statut du timer actif |
| `conducteur-orchestre inbox top` | Affiche les 3 tâches prioritaires avec leur numéro de ligne |
| `conducteur-orchestre inbox add --description "Action concrète"` | Ajoute une tâche dans `inbox.md` |
| `conducteur-orchestre inbox start --line 2` | Passe une tâche en cours |
| `conducteur-orchestre inbox pause --line 2` | Met une tâche en pause |
| `conducteur-orchestre inbox complete --line 2` | Marque une tâche terminée |

## Exemples

### Lire le focus courant

```powershell
build\dist\win-x64\conducteur-orchestre focus snapshot
```

### Ouvrir l'interface terminal locale

```powershell
build\dist\win-x64\conducteur-orchestre
```

La commande ouvre une interface `Terminal.Gui` branchée sur `/api/ui/snapshot`. Si l'URL cible est locale ou loopback et qu'aucun serveur ne répond, le CLI démarre automatiquement le serveur avant d'afficher l'interface.

L'interface garde les contraintes produit : un seul point d'attention principal, navigation clavier explicite et stockage local dans `inbox.md`.

La TUI propose 3 vues :

1. `Taches` par défaut, pour afficher les tâches racines non terminées, ouvrir l'édition, réordonner l'inbox et consulter les sous-tâches dans le détail ;
2. `Focus` pour capturer, choisir une tâche et piloter le timer ;
3. `Logs` pour suivre en direct la fin du fichier de log serveur, collée en permanence à la dernière ligne.

Quand un timer tourne, la vue `Focus` affiche une barre verte inversée représentant le temps restant.

Par défaut, le serveur écrit dans `logs\conducteur-orchestre.log`. La vue `Logs` lit ce tail via `GET /api/logs/tail`.

### Raccourcis TUI

| Contexte | Touche | Action |
| --- | --- | --- |
| Global | `T` | Aller à `Taches` |
| Global | `F` | Aller à `Focus` |
| Global | `L` | Aller à `Logs` |
| Global | `←` / `→` | Changer d'écran |
| Taches | `J` / `K` | Descendre / monter dans la liste active |
| Taches | `Tab` / `Shift+Tab` | Basculer entre la tâche racine et les sous-tâches du détail |
| Taches | `1` / `2` | Lancer le preset Pomodoro 1 ou 2 sur la tâche ou sous-tâche sélectionnée |
| Taches | `Maj+J` / `Maj+K` | Entrer en mode déplacement |
| Taches | `Entrée` | Ouvrir l'édition de la tâche sélectionnée |
| Taches | `N` | Créer une nouvelle tâche |
| Déplacement | `J` / `K` | Déplacer la tâche dans l'ordre visible |
| Déplacement | `Entrée` | Confirmer le nouvel ordre |
| Déplacement | `Echap` | Annuler le déplacement |
| Focus | `1` / `2` | Lancer le preset Pomodoro 1 ou 2 sur la tâche focus choisie |
| Edition | `Tab` / `Shift+Tab` | Passer d'un champ à l'autre |
| Edition | `J` / `K` | Changer la valeur d'un sélecteur |
| Edition | `Entrée` sur `Recurrence` | Ouvrir l'écran de récurrence |
| Edition | `Echap` | Fermer l'édition |
| Récurrence | `J` / `K` | Changer d'option |
| Récurrence | `Entrée` | Valider la règle |
| Récurrence | `Echap` | Revenir à l'édition |

### Lancer un focus

```powershell
build\dist\win-x64\conducteur-orchestre.exe timer presets
build\dist\win-x64\conducteur-orchestre.exe timer start --preset "Pomodoro Court" --line 2
```

### Ajouter une tâche dans l'inbox

```powershell
build\dist\win-x64\conducteur-orchestre.exe inbox add --description "Préparer le rendez-vous" --priority High --due 2026-04-08
```

Version enrichie pour une tâche quotidienne :

```powershell
build\dist\win-x64\conducteur-orchestre.exe inbox add --description "Préparer la facture" --details "Envoyer aussi le justificatif" --priority High --due 2026-04-12 --time 09:30 --reminder 1h --category "Admin" --repeat every:2:weeks
```

### Changer le statut d'une tâche

```powershell
build\dist\win-x64\conducteur-orchestre.exe inbox top
build\dist\win-x64\conducteur-orchestre.exe inbox start --line 2
build\dist\win-x64\conducteur-orchestre.exe inbox pause --line 2
build\dist\win-x64\conducteur-orchestre.exe inbox complete --line 2
```

## Règles de sortie

- les erreurs restent explicites ;
- les commandes lisibles retournent des sorties courtes ;
- le timer reste porté par le serveur local pour conserver un seul état actif ;
- la TUI réutilise les endpoints existants (`/api/ui/snapshot`, `/api/inbox`, `/api/timer`, `/api/logs/tail`) pour garder un état unique ;
- la liste principale `Taches` montre les tâches racines ouvertes et le détail expose leurs sous-tâches ;
- `inbox.md` reste lisible : `[>]` = en cours, `[ ]` = en pause, `[x]` = terminée, `⏱ Xm` = temps cumulé, `📝 [...]` = détail court, `🕒 HH:mm` = heure, `🔔 10m|1h|1d` = avance de rappel, `🏷 [...]` = catégorie/projet, `🔁 [...]` = répétition.
