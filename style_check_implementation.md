# Style Check - Implementierungsplan

## Übersicht
Das `style_check.md` enthält 4 Hauptverbesserungsvorschläge für die MarkSafe UI:

1. **Kopfbereich & Menüleisten aufräumen**
   - Hamburger-Menü statt klassischer Menüleiste
   - Einheitlicher Sub-Header für Editor/Vorschau

2. **Rahmen reduzieren und Abstände erhöhen**
   - Fehlende Rahmen entfernen
   - Mehr Padding (24-32px) für besseren Lesekomfort

3. **Typografie & Code-Highlighting modernisieren**
   - Bessere Editor-Schriftart (JetBrains Mono, Fira Code)
   - Moderne Vorschau-Schrift (Inter, Merriweather)
   - Dezentere Markdown-Syntax

4. **Statusleiste & Controls auffrischen**
   - Mehr Padding in Statusleiste
   - Kleinere Schrift
   - Verbesserte Controls

## Implementierungsstrategie

### Phase 1: UI-Struktur optimieren (höchste Priorität)
- [ ] Hamburger-Menü implementieren (ersetzt现有 Toolbar)
- [ ] Einheitlichen Header für Editor/Vorschau schaffen
- [ ] Remove redundant headers

### Phase 2: Visual polish (mittlere Priorität)
- [ ] Entferne harte Rahmenlinien
- [ ] Mehr Padding hinzufügen (24-32px)
- [ ] Subtile Hintergrundfarben nutzen

### Phase 3: Typografie verbessern (niedrige Priorität)
- [ ] Editor-Schriftart auf JetBrains Mono ändern
- [ ] Vorschau-Schriftart auf Inter/Merriweather ändern
- [ ] Markdown-Syntax dezent färben

### Phase 4: Statusleiste aktualisieren (niedrige Priorität)
- [ ] Mehr Padding hinzufügen
- [ ] Schriftgröße reduzieren
- [ ] Controls refresh

## Notwendige Änderungen

### Backend/UI-Konfiguration
- `frontend/src/components/Toolbar.svelte` → Ersetzen durch Hamburger-Menu-Komponente
- `frontend/src/App.svelte` → Tab-leiste strukturieren
- `frontend/src/themes.ts` → Typografie-Paare definieren
- `frontend/src/lib/constants.ts` → Neue Padding-Werte, Farben

### Design-Konstancen
- Neue Padding-Werte in constants.ts: `EDITOR_PADDING = 'p-8'`
- Neue typography config: `font-mono-jetbrains`, `font-sans-inter`
- Hamburger-menu component neu erstellen

## Branch-Strategie
**Branch-Name:** `style-check-improvements`

1. Current State: Alle Änderungen in commits (oben: 5 commits ahead)
2. Branch erstellen: `git checkout -b style-check-improvements`
3. Phase 1 implementieren → test → commit
4. Phase 2 implementieren → test → commit
5. Phase 3 implementieren → test → commit
6. Phase 4 implementieren → test → commit
7. Branch mergen oder verwerfen nach Review

## Empfohlene nächste Schritte
1. Branch erstellen: `git checkout -b style-check-improvements`
2. Phase 1 beginnen (Hamburger-Menü)
3. Incremental commits nach Phase
4. User feedback nach Phase 1+2 einholen

---
**Empfehlung:** Mit Phase 1 beginnen und nach jeder Phase testen. Die größten UX-Verbesserungen kommen aus Phase 1 & 2.
