/**
 * Theme defines the structural and visual properties for a preview style.
 * It coordinates styles between the Go backend (Chroma) and the Frontend (Tailwind/Mermaid).
 */
export interface Theme {
  id: string; // Internal identifier
  name: string; // Display name in the UI
  chromaStyle: string; // The specific style name used by the Go Chroma highlighter
  containerClass: string; // Tailwind background and text color classes for the preview pane
  proseClass: string; // Tailwind Typography classes (e.g., prose-invert)
  mermaidTheme: 'dark' | 'default' | 'neutral' | 'forest'; // Built-in Mermaid.js theme identifier
  mermaidVars?: Record<string, string>; // Custom CSS variables for fine-tuning Mermaid diagrams
}

/**
 * themes is the centralized list of available presets for the Markdown preview.
 */
export const themes: Theme[] = [
  {
    id: 'dark',
    name: 'Dark Mode',
    chromaStyle: 'github-dark',
    containerClass: 'bg-slate-900 text-slate-100',
    proseClass: 'prose-invert',
    mermaidTheme: 'dark',
    mermaidVars: {
      lineColor: '#94a3b8',
      textColor: '#f1f5f9'
    }
  },
  {
    id: 'light',
    name: 'Light Mode',
    chromaStyle: 'github',
    containerClass: 'bg-white text-slate-900',
    proseClass: 'prose-slate',
    mermaidTheme: 'default',
    mermaidVars: {
      lineColor: '#334155',
      textColor: '#0f172a'
    }
  },
  {
    id: 'sepia',
    name: 'Sepia Paper',
    chromaStyle: 'monokai-light',
    containerClass: 'bg-[#f4ecd8] text-[#5b4636]',
    proseClass: 'prose-stone',
    mermaidTheme: 'neutral',
    mermaidVars: {
      primaryColor: '#d4c5a9',
      primaryTextColor: '#5b4636',
      primaryBorderColor: '#a89d85',
      lineColor: '#5b4636',
      secondaryColor: '#e4dcc7',
      tertiaryColor: '#f4ecd8',
      edgeLabelBackground: '#f4ecd8',
      nodeBorder: '#a89d85'
    }
  },
  {
    id: 'monochrome',
    name: 'Monochrome',
    chromaStyle: 'bw',
    containerClass: 'bg-white text-black monochrome',
    proseClass: 'prose-slate',
    mermaidTheme: 'default',
    mermaidVars: {
      primaryColor: '#ffffff',
      primaryTextColor: '#000000',
      primaryBorderColor: '#000000',
      lineColor: '#000000',
      secondaryColor: '#eeeeee',
      tertiaryColor: '#ffffff',
      nodeBorder: '#000000'
    }
  }
];
