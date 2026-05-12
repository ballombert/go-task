# US : Les donnees de taches restent visibles et locales

## Description

En tant qu'utilisateur, je veux que les donnees utiles des taches restent stockees localement dans un fichier texte lisible et editable, afin de pouvoir les consulter et les corriger facilement.

## Decision de cadrage

- `inbox.md` est la source de verite des taches actives.
- SQLite peut rester un support auxiliaire local pour l'historique technique (ex. sessions timer, notifications), mais pas la source principale des taches.
- Le format de reference reste le format Obsidian Tasks sur une seule ligne par tache.
- Cette user story ne rend pas `done.md`, `recurring.md`, `sessions.md` ou `id.txt` obligatoires comme sources actives.

## Format de donnees retenu

Chaque tache reste une ligne markdown compatible Obsidian Tasks, enrichie par des metadonnees inline.

Exemple de tache :

```md
- [ ] Preparer la facture 📝 [Envoyer aussi le justificatif] 📅 2026-04-12 🕒 09:30 🔔 1h 🏷 [Admin] 🔁 [every:2:weeks] 🆔 [#ID-26-15-0001] ⏫
```

Exemple de tache en cours :

```md
- [>] Appeler le garage ⏱ 25m 🆔 [#ID-26-15-0002] 🔺
```

Exemple de tache terminee :

```md
- [x] Sortir la poubelle 📅 2026-04-10 🔁 [daily] 🆔 [#ID-26-15-0003] ⏫ ✅ 2026-04-10
```

## Regles de stockage

- une tache active ou terminee reste stockee dans `inbox.md` ;
- le statut continue d'utiliser `[ ]`, `[>]` et `[x]` ;
- le temps suivi reste stocke inline avec `⏱ Xm` ;
- la recurrence reste stockee inline avec `🔁 [regle]` ;
- l'identifiant visible reste stocke inline avec `🆔 [#ID-YY-WW-####]`.

## Regles d'identifiant

- le format de reference est `#ID-YY-WW-####` ;
- `YY` represente l'annee ISO de creation ;
- `WW` represente la semaine ISO de creation ;
- `####` represente un compteur sur 4 chiffres incremente dans la meme semaine ;
- le compteur est calcule en scannant les identifiants deja presents dans `inbox.md` ;
- aucun fichier `id.txt` n'est requis ;
- si une ligne historique ne contient pas d'identifiant visible, elle reste valide et continue d'etre lisible.

## Regles de recurrence

- une tache recurrente reste dans `inbox.md` avec sa regle `🔁 [regle]` ;
- quand une tache recurrente est terminee, une nouvelle occurrence est creee dans `inbox.md` ;
- la nouvelle occurrence recoit une nouvelle echeance et un nouvel identifiant visible ;
- `inbox.md` reste la seule source de verite des occurrences actives.

## Hors scope de cette iteration

- imposer `done.md` comme archive obligatoire ;
- imposer `recurring.md` comme source separee des taches recurrentes ;
- imposer `sessions.md` comme remplacement de SQLite ;
- introduire un mecanisme de fusion manuelle complexe en cas de conflit ;
- introduire une structure de sous-taches dediee avec numerotation `-01`, `-02`.

## Criteres d'acceptation

- une nouvelle tache creee via l'application est ecrite dans `inbox.md` avec un identifiant visible au format `🆔 [#ID-YY-WW-####]` ;
- une ligne existante sans identifiant visible continue d'etre parsee sans erreur ;
- une tache recurrente terminee cree une nouvelle ligne dans `inbox.md` avec un nouvel identifiant visible ;
- l'ordre et la lisibilite du format Obsidian Tasks restent preserves ;
- `inbox.md` reste la source de verite des taches actives.
