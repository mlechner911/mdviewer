#!/bin/sh
# Handtest Ordner-Prüfung: macht diesen Ordner zu einem Git-Repository (macOS, Linux).
cd "$(dirname "$0")" || exit 1
[ -d .git ] && { echo "Schon eingerichtet."; exit 0; }
git init -q
git config user.name "Handtest"
git config user.email "handtest@example.invalid"
git config commit.gpgsign false
git add .
git commit -q -m "Handtest: Ausgangsstand"
echo "Fertig. Jetzt sop.md in MarkSafe öffnen."
