# ⚡ WF - Minimalist Workflow Automation

**WF** is a stupidly simple procedural language to automate your everyday tasks (build, install, deploy). 
Write your workflows in `.wf` files. No complex syntax, no steep learning curve. Just simple, sequential commands.

🌍 [English](#-english) | 🇫🇷 [Français](#-français)

---

## 🇬🇧 English

### 🚀 Quick Start (with Sub-Workflows)

WF shines when you split your tasks into logical blocks. Calling one workflow from another creates a beautiful, readable execution tree.

1. **Create a `deploy.wf` file:**
```wf
[build] # Sub-workflow: Build the app
echo "Building the application..."
run npm install
run npm run build
sync_time

[permissions] # Sub-workflow: Set correct rights
echo "Setting permissions..."
set_permissions dist/ 0777

[deploy] # Main workflow
echo "Starting deployment pipeline..."
SET ENV=production
wf build
wf permissions
notify_success "Deployment done!"
```

2. **Run the main workflow:**
```bash
wf deploy
```
*WF automatically parses all `.wf` files in the current folder, detects the sections, and executes them top-to-bottom. Sub-workflows (`wf build`) will be visually indented in your terminal like a tree.*

### 🛠️ Built-in Commands vs System Commands

WF comes with **native** cross-platform commands. If a command isn't native, you can simply use `run <command>` to execute it via your operating system's default shell!

**Native Commands:**
- `SET VAR=value` : Define an environment variable for the session
- `wf <section>` : Call another workflow section (creates a visual tree 🌳)
- `copy <src> <dst>` : Copy a file or directory
- `mkdir <dir>` : Create a directory (and its parents)
- `touch <file>` : Create an empty file
- `set_permissions <path> <mode>` : Apply permissions (e.g., `0777`). Applied recursively for folders.
- `sync_time` : Synchronize system clock
- `echo <msg>` : Print a simple message
- `docker_compose <cmd>` : Run Docker Compose (auto-detects `docker compose` or `docker-compose`)
- `notify <msg>` : Standard OS notification
- `notify_success <msg>` : Success OS notification
- `notify_error <msg>` : Error OS notification
- `notify_warning <msg>` : Warning OS notification
- `notify_info <msg>` : Info OS notification
- `exit` : Stop the workflow immediately

**System Commands:**
- `run <cmd>` : Execute ANY system command (e.g., `run composer install`, `run php bin/console`).

### 🎯 Philosophy

- **KISS (Keep It Simple, Stupid)**: Procedural, top-to-bottom execution. No complex logic.
- **Beautiful Output**: Clean, tree-like execution logs for sub-workflows.
- **Maintainable**: Understandable by any developer instantly. 

---

## 🇫🇷 Français

### 🚀 Démarrage Rapide (avec Sous-Workflows)

WF excelle lorsque vous divisez vos tâches en blocs logiques. Appeler un workflow depuis un autre crée un arbre d'exécution propre et lisible.

1. **Créez un fichier `deploy.wf` :**
```wf
[build] # Sous-workflow : Build l'application
echo "Construction de l'application..."
run npm install
run npm run build
sync_time

[permissions] # Sous-workflow : Gère les droits
echo "Configuration des permissions..."
set_permissions dist/ 0777

[deploy] # Workflow principal
echo "Lancement du pipeline de déploiement..."
SET ENV=production
wf build
wf permissions
notify_success "Déploiement terminé !"
```

2. **Exécutez le workflow principal :**
```bash
wf deploy
```
*WF analyse automatiquement tous les fichiers `.wf`, détecte les sections, et les exécute de haut en bas. Les sous-workflows (`wf build`) seront visuellement indentés dans votre terminal sous forme d'arborescence.*

### 🛠️ Commandes Natives vs Commandes Système

WF intègre des commandes **natives** multiplateformes. Si une commande n'est pas native, utilisez simplement `run <commande>` pour l'exécuter via le shell de votre système d'exploitation !

**Commandes Natives :**
- `SET VAR=value` : Définir une variable d'environnement pour la session
- `wf <section>` : Appeler une autre section de workflow (crée un arbre visuel 🌳)
- `copy <src> <dst>` : Copier un fichier ou dossier
- `mkdir <dir>` : Créer un dossier (et ses parents)
- `touch <file>` : Créer un fichier vide
- `set_permissions <path> <mode>` : Appliquer des droits (ex: `0777`). Récursif pour les dossiers.
- `sync_time` : Synchroniser l'horloge système
- `echo <msg>` : Afficher un message simple
- `docker_compose <cmd>` : Lancer Docker Compose (détecte `docker compose` ou `docker-compose`)
- `notify <msg>` : Notification système standard
- `notify_success <msg>` : Notification système de succès
- `notify_error <msg>` : Notification système d'erreur
- `notify_warning <msg>` : Notification système d'avertissement
- `notify_info <msg>` : Notification système d'information
- `exit` : Arrêter le workflow immédiatement

**Commandes Système :**
- `run <cmd>` : Exécuter N'IMPORTE QUELLE commande système (ex: `run composer install`, `run ls -la`).

### 🎯 Philosophie

- **KISS (Keep It Simple, Stupid)** : Procédural, exécution de haut en bas. Aucune logique complexe.
- **Rendu Visuel** : Logs propres sous forme d'arborescence pour les sous-workflows.
- **Maintenable** : Compréhensible par n'importe quel développeur instantanément.

---

## 📄 Licence
**AGPL-3.0**. Free to use, modifications must be open-sourced under the same license.