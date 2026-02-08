# Theming Architecture (v0.2.0)

MD Viewer uses a modular JSON-based theming system.

## 1. JSON Configuration

The themes are now defined in external JSON files located in `frontend/src/themes/`.

- **`base.json`**: Contains the default settings shared by all themes (default Mermaid variables, base CSS, etc.).
- **`presets/*.json`**: Individual theme definitions (Dark, Light, Sepia, Monochrome) that override the base settings.

### Merging Logic
Themes are automatically merged in `themes.ts`:
1. Start with `base.json`.
2. Apply properties from the selected `preset.json`.
3. Deep-merge complex objects like `mermaidVars` and `customCSS`.

## 2. Structure of a Theme JSON

```json
{
  "id": "theme-id",
  "name": "Display Name",
  "chromaStyle": "syntax-style-name",
  "containerClass": "tailwind-background-classes",
  "proseClass": "tailwind-typography-classes",
  "mermaidTheme": "mermaid-id",
  "mermaidVars": {
    "lineColor": "#hex",
    "textColor": "#hex"
  }
}
```

## 3. Adding a New Theme

To add a new theme:
1. Create a new JSON file in `frontend/src/themes/presets/`.
2. Open `frontend/src/themes.ts`.
3. Import your new JSON file.
4. Add it to the `themes` array using the `createTheme()` helper.

## 4. Future Roadmap
The move to JSON allows for a future **Theme Editor** where users can modify these values directly in the app and save them to their local configuration.
