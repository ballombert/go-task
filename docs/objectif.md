# Objectif produit - ConducteurOrchestre

> Ce document définit l'objectif produit de référence de l'application. Copilot et les contributeurs doivent s'y conformer pour toute proposition de code, d'architecture, de fonctionnalité ou de documentation. En cas de doute, la réduction de la charge cognitive prime sur la sophistication technique.

## 1. Mission

ConducteurOrchestre est une application Windows/.NET locale conçue pour aider une personne TDAH à externaliser sa mémoire de travail et à soutenir ses fonctions exécutives.

L'application agit comme un chef d'orchestre cognitif :

- elle capte l'information importante dans un support externe fiable ;
- elle rend visible la prochaine action utile ;
- elle cadence le travail avec des minuteries adaptées ;
- elle déclenche des rappels concrets au bon moment ;
- elle limite volontairement la complexité pour ne pas ajouter de surcharge.

## 2. Problème utilisateur

Le produit part du constat suivant : pour une personne TDAH, une partie de l'organisation ne peut pas reposer uniquement sur la mémoire interne. Les informations, rappels et priorités doivent être déplacés dans l'environnement.

Les difficultés principales à compenser sont :

- mémoriser plusieurs informations en même temps ;
- prioriser sans se noyer dans trop d'options ;
- commencer une tâche ;
- rester focalisé ;
- changer de tâche sans perdre le fil ;
- revenir à une intention claire après une distraction.

## 3. Utilisateur cible

L'utilisateur cible est une personne TDAH qui :

- travaille sur Windows ;
- a besoin d'un système simple, visible et local ;
- utilise un fichier `inbox.md` comme point d'entrée unique des tâches ;
- bénéficie davantage de rappels courts et actionnables que d'interfaces riches ou abstraites.

## 4. Promesse produit

ConducteurOrchestre ne cherche pas à être un gestionnaire de projet généraliste. Sa promesse est plus précise :

1. externaliser la mémoire de travail ;
2. présenter la prochaine action la plus simple possible ;
3. n'activer qu'une intervention utile à la fois ;
4. garder le système cohérent, explicable et faible en charge cognitive.

## 5. Non-objectifs

Le produit ne doit pas devenir :

- une suite de productivité généraliste ;
- un outil multi-utilisateur ou collaboratif ;
- une interface surchargée avec trop de panneaux, d'états ou d'options ;
- un agent opaque qui décide seul sans logique explicable ;
- un système qui demande à l'utilisateur de retenir ce que l'application devrait mémoriser à sa place.

## 6. Principes directeurs obligatoires

Les invariants suivants doivent guider toute évolution du produit :

1. **Une seule intervention active à la fois**  
   Le système ne doit pas pousser plusieurs sollicitations concurrentes.

2. **Jamais plus de 3 informations principales affichées**  
   Exemple : top 3 des tâches prioritaires.

3. **Jamais plus de 2 choix proposés simultanément**  
   Le système doit réduire l'effort de décision.

4. **Toujours préférer l'action la plus simple**  
   Si deux solutions sont possibles, choisir celle qui demande le moins d'effort cognitif.

5. **Persistance visible avant intelligence complexe**  
   Une information importante doit être stockée dans un support local, lisible et modifiable.

6. **Règles déterministes avant automatisation opaque**  
   Les interventions doivent être explicables, auditables et prévisibles.

7. **Chaque fonctionnalité doit réduire la charge cognitive nette**  
   Une fonctionnalité élégante mais mentalement coûteuse est hors cible.

## 7. Responsabilités du Chef d'Orchestre

| Responsabilité | Attendu produit |
| --- | --- |
| Détecter l'état cognitif utile | Déduire des signaux simples : surcharge, absence de priorisation, timer terminé, absence de rythme, besoin de relance, besoin de transition. |
| Activer le bon soutien au bon moment | Choisir une seule intervention ou un seul sous-agent pertinent selon le contexte. |
| Limiter la charge cognitive | Appliquer strictement les règles de simplicité : 1 intervention, 3 informations max, 2 choix max. |
| Maintenir la cohérence | Éviter contradictions, doublons, escalades de complexité et empilement de consignes. |

## 8. Portée fonctionnelle actuelle

Le MVP actuel couvre les capacités suivantes :

| Domaine | Objectif | État actuel |
| --- | --- | --- |
| Inbox | Centraliser les tâches dans `inbox.md` | Implémenté |
| Priorisation | Limiter l'affichage aux 3 tâches les plus utiles | Implémenté |
| Timer | Cadrer le travail avec un seul timer actif | Implémenté |
| Historique | Journaliser les sessions de timer | Implémenté |
| Notifications | Envoyer des rappels ou encouragements concrets | Implémenté |
| Règles | Déclencher une intervention déterministe selon le contexte | Implémenté |

## 9. Carte cible des sous-agents cognitifs

Ces sous-agents représentent la cible fonctionnelle du produit. Ils n'ont pas tous le même niveau de maturité.

| Sous-agent | Fonction exécutive soutenue | Rôle cible | Couverture actuelle |
| --- | --- | --- | --- |
| Mémoire Externe | Mémoire de travail | Capturer, stocker, rappeler, rendre visible | Partielle via `Inbox` |
| Priorisation | Hiérarchisation | Trier, réduire, choisir la prochaine action | Partielle via `Top 3` et `Rules` |
| Planification | Séquencement | Transformer une priorité en étapes courtes | Roadmap |
| Temps & Rythme | Gestion du temps | Démarrer, minuter, rythmer, journaliser | Partielle via `Timer` |
| Anti-Distraction | Attention | Détecter la dérive et aider au retour au focus | Roadmap |
| Transition | Initiation et changement de tâche | Aider à commencer, terminer, basculer | Roadmap |

## 10. Modèle de tâche de référence

La source de vérité des tâches est un fichier Markdown local `inbox.md`, compatible avec le format Obsidian Tasks.

Format attendu :

```md
- [ ] Description courte et actionnable 📅 2026-04-07 ⏫
  - [ ] Sous-action courte et actionnable
- [>] Description courte et actionnable ⏱ 25m 🔺
- [x] Description courte et actionnable ⏱ 50m ✅ 2026-04-07
```

Règles de modélisation :

- `[ ]` = tâche en pause ;
- `[>]` = tâche en cours ;
- `[x]` = tâche terminée ;
- une sous-tâche se modélise par indentation Markdown sous sa tâche parente ;
- `⏱ Xm` = temps cumulé local suivi sur la tâche ;
- une seule tâche peut être `en cours` à la fois ;
- l'interface principale montre d'abord les tâches racines ; les sous-tâches se consultent dans le détail ;
- une tâche doit décrire une action concrète ;
- la description doit rester courte et claire ;
- la priorité doit aider au tri, pas complexifier la lecture ;
- les dates et le temps servent à rendre l'action visible dans le temps, pas à multiplier les champs.

Mapping des priorités :

| Priorité | Emoji |
| --- | --- |
| Highest | `🔺` |
| High | `⏫` |
| Medium | `🔼` |
| Low | `🔽` |
| Lowest | `⏬` |

## 11. Orientation technique de référence

Tant que la cible produit ne change pas, Copilot doit privilégier les choix suivants :

- architecture locale, simple et explicable ;
- stockage local et lisible (`inbox.md`, SQLite) ;
- logique de règles déterministes ;
- notifications courtes et concrètes ;
- compatibilité Windows ;
- évolution incrémentalement utile plutôt que refonte ambitieuse.

Technologies actuellement cohérentes avec cet objectif :

- .NET 10 ;
- Minimal API ;
- Vertical Slice Architecture ;
- SQLite pour la journalisation locale ;
- Markdown / Obsidian Tasks pour l'inbox ;
- notifications système Windows ;
- workers d'arrière-plan pour rappels et évaluation des règles.

## 12. Règles de décision pour Copilot

Quand Copilot propose ou modifie du code dans ce dépôt, il doit appliquer les règles suivantes :

1. vérifier qu'une fonctionnalité soutient bien une fonction exécutive utile ;
2. refuser toute complexité qui augmente la charge cognitive sans gain net ;
3. préférer une logique simple, déterministe et visible ;
4. conserver l'idée d'un point d'entrée unique pour les tâches ;
5. faire respecter les limites de simplicité du produit ;
6. considérer `docs\objectif.md` comme la source de vérité produit en cas d'ambiguïté.

## 13. Critère de réussite produit

Une évolution est considérée comme cohérente avec l'objectif si elle permet au système de mieux :

- capturer rapidement une tâche ;
- montrer la prochaine action utile ;
- limiter le nombre de décisions à prendre ;
- soutenir le focus avec un seul rythme actif ;
- déclencher des rappels concrets, courts et explicables ;
- rester simple à comprendre, modifier et maintenir.
