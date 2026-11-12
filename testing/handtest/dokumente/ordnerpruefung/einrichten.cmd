@echo off
rem Handtest Ordner-Prüfung: macht diesen Ordner zu einem Git-Repository.
rem Braucht Git für Windows. Die Identität gilt nur für dieses Repository.
cd /d "%~dp0"
if exist .git (echo Schon eingerichtet. & pause & exit /b 0)
git init -q || (echo Git fehlt: https://git-scm.com & pause & exit /b 1)
git config user.name "Handtest"
git config user.email "handtest@example.invalid"
git config commit.gpgsign false
git add .
git commit -q -m "Handtest: Ausgangsstand"
echo Fertig. Jetzt sop.md in MarkSafe oeffnen.
pause
