# Handtest unter Windows 11

Vor jeder Freigabe testen wir MarkSafe von Hand unter Windows 11. Die
automatischen Tests (`task check`, `task mlcdoc:test`) prüfen Code und PDF,
aber nicht Installer, Fenster, Dialoge und Drucken.

Dieser Ordner enthält die Liste der Prüfungen und die Testdokumente dazu. Den
Teil für MarkSafePlus beschreibt das Modul selbst in
`frontend/src/sensormonitor/docs/HANDTEST.md`.

## Ablauf

1. **Vorbereiten** (Linux):

   ```sh
   task handtest:prepare            # Standard-Build
   task handtest:prepare SENSOR=1   # zusätzlich MarkSafePlus
   ```

   Die Task baut für Windows und legt in `/mnt/data2tb/dropzone/marksafe-test/`
   ab:
   - `test-marksafe.exe` und `test-marksafe-setup.exe`
   - mit `SENSOR=1` zusätzlich `test-marksafeplus.exe` und
     `test-marksafeplus-setup.exe`
   - `handtest/`: die Testdokumente aus `dokumente/` und diese Liste
   - `handtest/STAND.txt`: Version und Commit, aus denen der Build stammt

2. **Prüfen** (Windows 11): die Liste unten Punkt für Punkt durchgehen.
   Abweichungen gleich als Ticket festhalten, mit Dokument und Schritt.

3. **Melden** (Linux):

   ```sh
   task handtest:report -- standard passed "alles wie erwartet"
   task handtest:report -- plus failed "Schritt P3: Entwurf nicht gespeichert"
   ```

   Das meldet das Ergebnis an den Doc Hub (Tests `other-win11-standard` und
   `other-win11-plus`), mit Version und Commit aus `STAND.txt` vor der Notiz.
   Beim Anlegen der Release-Notes warnt der Doc Hub, wenn ein Handtest fehlt
   oder fehlgeschlagen ist.

Ein **Teiltest** zu einer einzelnen Änderung braucht nicht die ganze Liste.
Gemeldet wird dann mit `passed` und einer Notiz, was geprüft wurde, zum
Beispiel „nur Feldschema-Cache, S5“. Vor der Freigabe läuft die ganze Liste.

## Liste: MarkSafe (Standard)

Die Dokumente liegen in `handtest/` in der dropzone.

### Installation und Start

| # | Schritt | Erwartet |
|---|---|---|
| I1 | `test-marksafe-setup.exe` ausführen | Installer läuft durch; die letzte Seite bietet „MarkSafe starten“ an |
| I2 | MarkSafe über die letzte Seite starten | Startseite mit dem, was MarkSafe kann, und Logo |
| I3 | Über… im Hilfe-Menü | Versionsnummer wie in `STAND.txt` |
| I4 | `test-marksafe.exe` ohne Installation starten | Startet; PDF-Vorschau nur, wenn mlcdoc installiert ist oder neben der exe liegt |

### Bearbeiten

| # | Schritt | Erwartet |
|---|---|---|
| B1 | Ein Dokument öffnen, ändern, speichern; Speichern unter | Titel zeigt ungespeichert / gespeichert; Datei auf der Platte geändert |
| B2 | Öffnen-Dialog abbrechen | Keine Fehlermeldung |
| B3 | Vorschau, Synchron-Scrollen, Fokus-Ansicht | Vorschau folgt dem Editor |
| B4 | **H ▾** in der Werkzeugleiste | H1 bis H3 setzen die Überschrift |
| B5 | Statusleiste | Wörter, Zeichen, Lesezeit, Pfad; bei MLC-Dokumenten die Prüfung |

### Formularfelder und Schema (`schema/`)

Vorher `beispiel-schema.yaml` in den Ordner kopieren, den **Datei › Ordner
für Feld-Schemas öffnen** zeigt.

| # | Schritt | Erwartet |
|---|---|---|
| S1 | `1-ueber-adresse.md` öffnen | `patient.name` ohne Meldung; `patient.nmae` ist ein Fehler (unbekannte Eigenschaft); `name` ist ein Hinweis: alte Form von `patient.name`, bleibt gültig; `eigenes_feld` ohne Meldung |
| S2 | dasselbe ohne `beispiel-schema.yaml` im Ordner (Datei kurz umbenennen) | Hinweis: Schema ist hier nicht bekannt, Felder mit Punkt werden nicht geprüft |
| S3 | `2-im-dokument.md` | Schema aus dem Front Matter: `patient.nmae` ist ein Fehler |
| S4 | `3-ohne-schema.md` | `probe.name` nur Hinweis (kein Schema genannt); `#a` neben `#a.b` ist ein Fehler |
| S5 | `4-editor.md`: über `patient.visit_time` fahren; unter „Hier tippen:“ `[@#patient.` tippen | Beschreibung aus dem Schema; Vervollständigung bietet die Eigenschaften an; `patient.name` mit `TT.MM.JJJJ` ist eine Warnung |
| S6 | Formularfeld-Dialog (Werkzeugleiste oder `Strg+Umschalt+F`), Doppelklick auf ein Feld | Feld einfügen und bearbeiten; das Namensfeld bietet die Eigenschaften des Schemas an |
| S7 | `5-pdf-ebenen.md` als PDF exportieren, in Acrobat öffnen | Felder heißen `patient.name` usw., als Ebenen `patient` › `name`; Ankreuzfeld `kontrolle.erledigt` |

### Ausfüllen

| # | Schritt | Erwartet |
|---|---|---|
| A1 | `5-pdf-ebenen.md` › Ausfüllen | Maske mit Text, Bildern und Links zwischen den Feldern; Feldnamen lassen sich ausblenden |
| A2 | Werte eintragen, als Nachweis speichern | Neue Datei `<doc_id>_<datum>.md` mit `meta:`-Block; die Vorlage bleibt unverändert |
| A3 | Während des Ausfüllens im Markdown ändern | Die Maske übernimmt die Änderung |

### Bilder (`bild/`)

| # | Schritt | Erwartet |
|---|---|---|
| G1 | `formular-mit-bild.md` öffnen | `foto.jpg` in Vorschau und PDF; `grafik.webp` in der Vorschau, im PDF übersprungen |
| G2 | 🖼 in der Werkzeugleiste: Datei, Bild aus der Zwischenablage, gezogene Datei | Wird als Verweis oder eingebettet eingefügt; eingebettet zeigt der Editor `🖼 name ‹Größe›`, Klick zeigt die Daten, ▲ oder Esc kürzt wieder |
| G3 | Ein Bild von außerhalb des Dokumentordners einfügen (z. B. `fremd/aussen.jpg` aus einem Dokument eine Ebene höher, oder ein Bild vom Desktop) und „Verweis“ wählen | Meldung: liegt außerhalb des Dokumentordners, kann nur eingebettet werden; eine WebP-Datei lässt sich gar nicht einbetten |
| G4 | **Datei › Bilder einbetten…**, danach Speichern unter | Bilder stehen im Dokument; ein signiertes Dokument wird nicht verändert |
| G5 | `bildgroesse.md` in der Vorschau | 120 px breit; in 300 × 80 eingepasst, nicht verzerrt; halbe Textbreite; eigene Größe. Bei `\|0` warnt der Editor, das Bild erscheint in eigener Größe. `A\|B` bleibt Alt-Text |
| G6 | **Ansicht › Bilder zentrieren** ein und aus, App neu starten, HTML-Export | Bilder allein im Absatz stehen mittig; die Einstellung bleibt nach dem Neustart; der HTML-Export zeigt sie ebenso |
| G7 | 🖼 einfügen mit Größe `200`, dann `abc` und `150%` | `![…\|200](…)`; bei ungültiger Größe rotes Feld und „Einfügen“ gesperrt |
| G8 | Cursor in ein Bild: „✎ Bild“, Doppelklick auf `](foto.jpg)`, Rechtsklick › Bild bearbeiten… | Dialog mit Text und Größe vorbelegt; Übernehmen ändert nur `[…]`, Pfad und `"Titel bleibt"` bleiben. „Anderes Bild…“ tauscht die Datei. Bei einem eingebetteten Bild zeigt der Dialog eine kleine Vorschau |
| G9 | PDF-Vorschau von `bildgroesse.md` | Bilder noch in eigener Größe (mlcdoc kennt die Größe noch nicht); der Editor weist darauf hin, es gibt keinen Fehler |

### HTML und PDF (`html/`)

| # | Schritt | Erwartet |
|---|---|---|
| H1 | `html-test.md` in der Vorschau | Wie auf GitHub |
| H2 | PDF-Vorschau und Export | A1–A4 fett und kursiv, B2/B3 durchgestrichen, B8 als Code; andere Inline-Tags nur mit Text; `<div>`, `<details>` usw. fehlen; der Editor markiert sie, der Export sagt es |
| H3 | Drucken | Druckdialog, Ausdruck wie die PDF-Vorschau |

### Ordner-Prüfung und Commit (`ordnerpruefung/`)

Beschreibung: `docs/FOLDER-CHECK.md`. Der Ordner hat eine `.marksafe.yaml`, die
nach jedem Speichern `bin/pruefen` startet (Zeitlimit 10 s). Das Prüfprogramm
reagiert auf Zeilen, die mit `FEHLER`, `WARNUNG`, `ABSTURZ`, `LANGSAM` oder
`KEIN-VORLAGENORDNER` beginnen, und vergleicht die Version von `sop.md` und
`sop.en.md`.

Vorher den Ordner `ordnerpruefung` an einen eigenen Platz kopieren (z. B. auf
den Desktop) und dort `einrichten.cmd` ausführen (macOS: `einrichten.sh`).
Das braucht Git und legt ein Repository mit eigener Test-Identität an. Danach
`sop.md` in MarkSafe öffnen. Den Stand prüft `git log --stat` im Ordner.

| # | Schritt | Erwartet |
|---|---|---|
| O1 | `sop.md` ändern, speichern | Dialog „Ordner-Prüfung erlauben?“ mit Befehl `bin/pruefen .` und Arbeitsordner |
| O2 | „Jetzt nicht“; noch einmal ändern und speichern | Statusleiste „Ordner-Prüfung nicht erlaubt“ und „Gespeichert, nicht committet“; kein zweiter Dialog |
| O3 | Klick auf „Ordner-Prüfung nicht erlaubt“ › „Erneut prüfen“ › „Erlauben“ | Prüfung läuft; „keine Probleme“; der Commit-Dialog öffnet sich mit `SOP-DEMO-001 v1: ` |
| O4 | Nachricht ergänzen, „Committen“ (oder Strg+Enter) | Meldung „Committet“; `git log --stat`: neuer Commit nur mit `sop.md`, Autor „Handtest“ |
| O5 | MarkSafe neu starten, `sop.md` ändern, speichern | Kein Erlaubnis-Dialog mehr; Prüfung und Commit-Dialog wie in O3 |
| O6 | Zeile `FEHLER hier` und Zeile `WARNUNG da` schreiben, speichern | Im Editor rot und gelb an den Zeilen, Quelle „Ordner-Prüfung“; Statusleiste „1 Fehler, 1 Warnung(en)“ und „Gespeichert, nicht committet“; kein Commit-Dialog |
| O7 | Weitertippen, ohne zu speichern | Die Markierungen der Ordner-Prüfung verschwinden aus dem Text (die der MLC-Erweiterung bleiben) |
| O8 | Klick auf die Statusleiste | Liste unter `sop.md` mit beiden Befunden und Zeile; Klick springt zur Zeile |
| O9 | `FEHLER`-Zeile löschen, speichern | Nur die Warnung; Commit-Dialog öffnet sich; „Später“ › Statusleiste „Gespeichert, nicht committet“; Klick darauf öffnet den Commit-Dialog wieder |
| O10 | In `sop.md` `version: 2` setzen, speichern | Fehler ohne Zeile unter `sop.en.md` („Übersetzung hat Version 1 …“); Klick öffnet `sop.en.md`; `sop.en.md` auf 2 setzen und speichern › keine Probleme, Commit-Dialog für `sop.en.md`. Danach im Tab `sop.md` auf „Gespeichert, nicht committet“ klicken › Commit von `sop.md` geht ebenfalls |
| O11 | Zeile `ABSTURZ` schreiben, speichern | „Ordner-Prüfung fehlgeschlagen“; die Liste zeigt „exit code 3“ und „panic: Absturz zum Test“; kein Commit |
| O12 | Zeile `LANGSAM` statt `ABSTURZ`, speichern | Nach etwa 10 s „fehlgeschlagen“ mit „timeout after 10s“; MarkSafe bleibt bedienbar |
| O13 | Zeile `KEIN-VORLAGENORDNER`, speichern | Gelbe Meldung „Das ist kein Vorlagen-Ordner (Testfall).“; kein Commit |
| O14 | Testzeilen entfernen; in `.marksafe.yaml` den Befehl auf `["bin/pruefen", ".", "-x"]` ändern; `sop.md` speichern | Erlaubnis-Dialog erscheint erneut |
| O15 | In `.marksafe.yaml` `version: 9` setzen, `sop.md` speichern | „.marksafe.yaml ist ungültig“, Tooltip nennt die Version; kein Befehl läuft. Danach zurück auf `version: 1` |
| O16 | Im Ordner `git config --unset user.name`, ändern, speichern, committen | Commit-Dialog meldet: Git kennt dich nicht …; danach `git config user.name Handtest` |
| O17 | Ein Dokument außerhalb von `ordnerpruefung` speichern | Keine Ordner-Prüfung in der Statusleiste, kein Dialog |

### Neu in dieser Version

Den Abschnitt der Version im `CHANGELOG.md` durchgehen und jede Zeile einmal
ausprobieren, die hier noch nicht vorkommt. Was sich bewährt, kommt danach in
diese Liste.
