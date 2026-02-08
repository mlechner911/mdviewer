Projektphasen: Wails Markdown Viewer
Phase 1: Setup & Infrastruktur
Wails-Initialisierung: Erstellen des Projekt-Scaffolds mit Go-Backend und Svelte-TypeScript-Frontend.
Frontend-Grundausstattung: Installation von Tailwind CSS und dem @tailwindcss/typography Plugin für das automatische Styling der HTML-Ausgabe.
Go-Modulverwaltung: Installation der Kern-Bibliotheken für Markdown-Parsing (goldmark), Syntax-Highlighting (chroma) und HTML-Sicherung (bluemonday).

Phase 2: Das Go-Backend (Logik-Layer)
Markdown-Engine: Konfiguration des Parsers für GitHub Flavored Markdown (Tabellen, Checklisten, Emojis, mermaid).
Highlighting-Integration: Einbindung der Syntax-Hervorhebung für Code-Blöcke direkt im Konvertierungsprozess.
Sicherheits-Layer: Implementierung eines HTML-Sanitizers, der XSS verhindert, aber notwendige CSS-Klassen für das Styling beibehält.
Betriebssystem-Schnittstelle: Erstellung von Funktionen für den Zugriff auf das Dateisystem (Öffnen, Lesen und Speichern von .md Dateien).
Phase 3: Das Svelte-Frontend (UI-Layer)
Zwei-Spalten-Layout: Implementierung einer geteilten Ansicht (Links: Editor/Text-Area, Rechts: Vorschau-Fenster).
Echtzeit-Rendering: Programmierung der Schnittstelle, die Textänderungen sofort an das Go-Backend sendet und das Ergebnis im Viewer darstellt.
Native Dialoge: Einbindung von Wails-Laufzeitfunktionen für native "Datei öffnen"- und "Speichern"-Dialoge.
Styling-Optimierung: Konfiguration des Dark-Mode-Verhaltens für den Editor und das Dokument-Layout.
Phase 4: Polishing & Deployment
Performance-Check: Optimierung des Rendering-Prozesses bei sehr großen Markdown-Dateien.
Asset-Handling: Sicherstellung, dass Bilder und lokale Verknüpfungen im Viewer korrekt aufgelöst werden.
Finaler Build: Kompilierung des Projekts in eine einzige, plattformspezifische ausführbare Datei (.exe oder .app).
