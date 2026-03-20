import { writable, derived } from 'svelte/store';

/**
 * Detect the best matching locale from the browser settings.
 * Falls back to 'en' if no supported match is found.
 */
function getInitialLocale(): string {
  if (typeof navigator === 'undefined') return 'en';
  
  const supported = ['en', 'de', 'es', 'fr'];
  const browserLang = navigator.language.split('-')[0].toLowerCase();
  
  return supported.includes(browserLang) ? browserLang : 'en';
}

export const locale = writable(getInitialLocale());

export const translations: Record<string, any> = {
  en: {
    open: 'Open',
    save: 'Save',
    export: 'Export to HTML',
    app: 'App',
    previewTheme: 'Preview Theme',
    zoom: 'Zoom',
    editor: 'Editor',
    preview: 'Preview',
    placeholder: 'Type markdown here...',
    welcomeTitle: '# Welcome to MD Viewer',
    ready: 'Ready',
    toggleTheme: 'Toggle App Theme',
    untitled: 'Untitled',
    newTab: 'New Tab',
    print: 'Print / PDF',
    hideEditor: 'Hide Editor',
    showEditor: 'Show Editor',
    focusMode: 'Focus Mode',
    exitFocusMode: 'Exit Focus Mode',
    words: 'words',
    characters: 'characters',
    readingTime: 'Approx. %s min read',
    themeLabel: 'Theme:',
    filesLoaded: 'Loaded %s files'
  },
  de: {
    open: 'Öffnen',
    save: 'Speichern',
    export: 'Als HTML exportieren',
    app: 'App',
    previewTheme: 'Vorschau-Design',
    zoom: 'Zoom',
    editor: 'Editor',
    preview: 'Vorschau',
    placeholder: 'Markdown hier eingeben...',
    welcomeTitle: '# Willkommen beim MD Viewer',
    ready: 'Bereit',
    toggleTheme: 'App-Design umschalten',
    untitled: 'Unbenannt',
    newTab: 'Neuer Tab',
    print: 'Drucken / PDF',
    hideEditor: 'Editor ausblenden',
    showEditor: 'Editor einblenden',
    focusMode: 'Fokus-Modus',
    exitFocusMode: 'Fokus-Modus verlassen',
    words: 'Wörter',
    characters: 'Zeichen',
    readingTime: 'Ca. %s Min. Lesezeit',
    themeLabel: 'Design:',
    filesLoaded: '%s Dateien geladen'
  },
  es: {
    open: 'Abrir',
    save: 'Guardar',
    export: 'Exportar a HTML',
    app: 'App',
    previewTheme: 'Tema de Vista Previa',
    zoom: 'Zoom',
    editor: 'Editor',
    preview: 'Vista Previa',
    placeholder: 'Escriba markdown aquí...',
    welcomeTitle: '# Bienvenido a MD Viewer',
    ready: 'Listo',
    toggleTheme: 'Cambiar tema de la aplicación',
    untitled: 'Sin título',
    newTab: 'Nueva pestaña',
    print: 'Imprimir / PDF',
    hideEditor: 'Ocultar editor',
    showEditor: 'Mostrar editor',
    focusMode: 'Modo enfoque',
    exitFocusMode: 'Salir del modo enfoque',
    words: 'palabras',
    characters: 'caracteres',
    readingTime: 'Aprox. %s min de lectura',
    themeLabel: 'Tema:',
    filesLoaded: '%s archivos cargados'
  },
  fr: {
    open: 'Ouvrir',
    save: 'Enregistrer',
    export: 'Exporter en HTML',
    app: 'App',
    previewTheme: 'Thème d\'aperçu',
    zoom: 'Zoom',
    editor: 'Éditeur',
    preview: 'Aperçu',
    placeholder: 'Tapez du markdown ici...',
    welcomeTitle: '# Bienvenue sur MD Viewer',
    ready: 'Prêt',
    toggleTheme: 'Changer le thème de l\'application',
    untitled: 'Sans titre',
    newTab: 'Nouvel onglet',
    print: 'Imprimer / PDF',
    hideEditor: 'Masquer l\'éditeur',
    showEditor: 'Afficher l\'éditeur',
    focusMode: 'Mode focus',
    exitFocusMode: 'Quitter le mode focus',
    words: 'mots',
    characters: 'caractères',
    readingTime: 'Environ %s min de lecture',
    themeLabel: 'Thème :',
    filesLoaded: '%s fichiers chargés'
  }
};

export const t = derived(locale, ($locale) => (key: string, ...args: any[]) => {
  let text = translations[$locale][key] || key;
  args.forEach(arg => {
    text = text.replace('%s', String(arg));
  });
  return text;
});