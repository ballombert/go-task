# Terminal UI

Cette page decrit une proposition de TUI plus visuelle, avec une palette coloree et une navigation claire. Le but est d'avoir un terminal qui reste lisible mais qui ne ressemble pas a un simple dump texte.

## Direction visuelle

Palette proposee pour le theme terminal :

- fond principal : ardoise sombre
- texte principal : ivoire
- accent navigation : ambre
- action principale : turquoise
- timer actif : vert menthe
- alerte / stop : corail
- information secondaire : bleu ciel
- tache terminee : gris bleute

Roles visuels :

- navigation active : fond ambre, texte sombre
- bouton principal : fond turquoise, texte sombre
- bouton stop : fond corail, texte clair
- tache selectionnee : fond bleu ciel ou bordure accentuee
- tache en cours : accent vert
- tache terminee : contraste affaibli
- feedback succes : vert ou turquoise
- feedback erreur : corail

## Navigation

Navigation cible :

- Focus
- Taches
- Logs

Ecrans cibles a terme :

- Focus
- Pomodoro en cours
- Liste de taches
- Detail tache
- Creation tache
- Logs

## Ecran Focus

Objectif : capturer vite, choisir une tache, lancer ou arreter un timer.

```text
+============================================================================+
| CONDUCTEUR ORCHESTRE                           [ Focus ] [ Taches ] [ Logs ] |
+============================================================================+
| Etat : Pret a lancer une session                                           |
+----------------------------------------------------------------------------+
| Capture rapide                                                             |
| [ Ecrire la prochaine action utile................................. ]       |
| [Important]  [Aujourd'hui]  [Ajouter a l'inbox]                            |
+----------------------------------------------------------------------------+
| Focus du moment                                                            |
| Timer : Pomodoro Court - 25 min focus / 5 min pause                        |
| [Lancer preset 1]   [Lancer preset 2]   [Arreter le timer]                 |
|                                                                            |
| Choix rapides                                                              |
| [> Revoir la story TUI]   [  Corriger la vue logs]                         |
|                                                                            |
| Selection                                                                  |
| Revoir la story TUI - en pause - 15 min                                    |
| Dossier docs / UX                                                          |
| [Mettre en cours]   [Terminer]                                             |
+----------------------------------------------------------------------------+
| Top 3                                                                      |
| 1. Revoir la story TUI - priorite haute - 15 min                           |
| 2. Corriger la vue logs - echeance 2026-04-12                              |
| 3. Verifier le timer - recurrence chaque jour                              |
+----------------------------------------------------------------------------+
| Intervention                                                               |
| Une seule chose a faire maintenant : demarrer la tache selectionnee.       |
+============================================================================+
```

## Ecran Pomodoro En Cours

Objectif : maximiser le contraste sur la tache active et le temps restant.

```text
+============================================================================+
| SESSION ACTIVE                                               [Stop session] |
+============================================================================+
| Tache : Revoir la story TUI                                               |
| Phase : Focus                                                             |
| Temps restant : 24:12                                                     |
| Temps cumule : 40 min                                                     |
|=========================================================================== |
|                                                                           |
|                           24:12                                           |
|                     [##################......]                            |
|                                                                           |
| Prochaine pause : 5 min                                                   |
| Notes : garder le wording simple                                          |
|                                                                           |
+============================================================================+
```

## Ecran Liste De Taches

Objectif : parcourir l'inbox locale avec une vraie selection visuelle.

```text
+============================================================================+
| TACHES                                                  [ Focus ] [ Logs ] |
+============================================================================+
| [Precedente]                                       Page 1 / 2  [Suivante]  |
+----------------------------------------------------------------------------+
| > [>] Revoir la story TUI                               haute      15 min   |
|   [ ] Corriger la vue logs                              moyenne      0 min   |
|   [ ] Ajouter theme ambre                               haute        0 min   |
|   [x] Relire la roadmap                                 terminee    25 min   |
|   [ ] Ecrire note release                               basse        0 min   |
|   [ ] Verifier timer preset                             moyenne      5 min   |
+----------------------------------------------------------------------------+
| Detail selection                                                           |
| Revoir la story TUI                                                       |
| Statut : en cours                                                         |
| Echeance : 2026-04-12 09:00                                               |
| Categorie : docs                                                           |
| Rappel : 1 h avant                                                         |
| Details : rafraichir la maquette et aligner la vraie TUI                   |
| [Mettre en pause]   [Terminer]                                             |
+============================================================================+
```

## Ecran Detail Tache

Objectif : afficher une fiche compacte, lisible, sans formulaire dense.

```text
+============================================================================+
| DETAIL TACHE                                                  [Retour liste]|
+============================================================================+
| Revoir la story TUI                                                       |
|                                                                            |
| Statut        : en cours                                                  |
| Priorite      : haute                                                     |
| Temps cumule  : 40 min                                                    |
| Echeance      : 2026-04-12 09:00                                          |
| Categorie     : docs                                                      |
| Rappel        : 1 h avant                                                 |
| Recurrence    : aucune                                                    |
|                                                                            |
| Details                                                                    |
| Rafraichir la maquette, ajouter les couleurs, brancher la vue liste.      |
|                                                                            |
| [Mettre en pause]   [Terminer]   [Modifier]                               |
+============================================================================+
```

## Ecran Creation Tache

Objectif : un formulaire simple, champs visibles, CTA unique tres net.

```text
+============================================================================+
| NOUVELLE TACHE                                                [Annuler]     |
+============================================================================+
| Description                                                               |
| [ Ajouter une action courte et concrete.............................. ]    |
|                                                                            |
| Details                                                                   |
| [ Optionnel : contexte utile......................................... ]    |
|                                                                            |
| Priorite : [aucune] [basse] [moyenne] [haute] [max]                       |
| Echeance : [2026-04-12]   Heure : [09:00]                                 |
| Rappel   : [1 h avant]                                                    |
|                                                                            |
|                                                     [Enregistrer la tache] |
+============================================================================+
```

## Premiere tranche d'implementation

Pour une premiere iteration dans le code :

- ajouter un theme colore centralise dans la TUI Terminal.Gui ;
- ajouter un onglet `Taches` avec liste locale et detail synchronise ;
- garder `Focus` et `Logs` ;
- conserver les actions existantes : ajouter, demarrer, mettre en pause, terminer, arreter le timer.

Les ecrans `Detail` et `Creation` peuvent d'abord rester une cible de maquette, puis devenir des vues dediees dans une seconde iteration.