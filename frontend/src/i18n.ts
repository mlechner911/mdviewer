import { writable, derived } from 'svelte/store';

export const locale = writable('en');

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
    toggleTheme: 'Toggle App Theme'
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
    toggleTheme: 'App-Design umschalten'
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
    toggleTheme: 'Cambiar tema de la aplicación'
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
    toggleTheme: 'Changer le thème de l\'application'
  }
};

export const t = derived(locale, ($locale) => (key: string) => {
  return translations[$locale][key] || key;
});