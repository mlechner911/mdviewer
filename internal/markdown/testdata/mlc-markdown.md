---
title: Testdokument MLC-Markdown
version: "1.0"
tags: [sop, test]
---
# Standardarbeitsanweisung (SOP)

**Titel:** Testdokument für den gemeinsamen Dialekt
**Dokumenten-ID:** SOP-TEST-001
**Version:** 1.0

---

## 1. Hinweisboxen {#boxen}

> [!NOTE]
> Hinweis mit **fettem** Text.

> [!TIP]
> Tipp.

> [!IMPORTANT]
> Wichtig.

> [!WARNING]
> Bei mehr als −70 °C Proben umlagern.

> [!CAUTION]
> Vorsicht.

> Normales Zitat.

## 2. Tabellen und Listen

| Maßnahme | Verantwortlich | Frist |
| --- | --- | --- |
| Fühler tauschen | Technik | 16.10.2026 |
| Protokoll prüfen | QMB | 20.10.2026 |

- [x] erledigt
- [ ] offen

1. Erster Schritt
2. Zweiter Schritt[^1]

[^1]: Fußnote.

## 3. Code, Formel, Diagramm

Inline `code` und Formel $E=mc^2$.

```go
func main() { fmt.Println("hallo") }
```

```mermaid
graph LR; A-->B;
```

<div class="page-break"></div>

## 4. Formularfelder

Ort: [@#ort ______________________] <!--* z. B. Kühlraum A -->

Datum: [@#datum TT.MM.JJJJ] Kritisch: [@#kritisch]

Zurück zu [den Boxen](#boxen).
