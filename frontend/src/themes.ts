export interface Theme {
  id: string;
  name: string;
  chromaStyle: string; // The Go Chroma highlighter style
  containerClass: string; // Background and text colors
  proseClass: string; // Tailwind Typography classes
  mermaidTheme: 'dark' | 'default' | 'neutral' | 'forest';
  mermaidVars?: Record<string, string>;
}

export const themes: Theme[] = [
  {
    id: 'dark',
    name: 'Dark Mode',
    chromaStyle: 'github-dark',
    containerClass: 'bg-slate-900 text-slate-100',
    proseClass: 'prose-invert',
    mermaidTheme: 'dark'
  },
  {
    id: 'light',
    name: 'Light Mode',
    chromaStyle: 'github',
    containerClass: 'bg-white text-slate-900',
    proseClass: 'prose-slate',
    mermaidTheme: 'default'
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
      tertiaryColor: '#f4ecd8'
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
      tertiaryColor: '#ffffff'
    }
  }
];
