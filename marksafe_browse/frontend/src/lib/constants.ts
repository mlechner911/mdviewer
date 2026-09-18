export const APP_THEME = {
  DARK: 'dark',
  LIGHT: 'light',
  AUTO: 'auto',
} as const;

export const DEFAULTS = {
  fontSize: 100,
  splitWidth: 50,
};

export const STYLE = {
  toolbar: { dark: 'bg-slate-800 border-slate-700', light: 'bg-white border-slate-200' },
  editor: { dark: 'bg-slate-900', light: 'bg-white' },
  button: { dark: 'hover:bg-slate-700', light: 'hover:bg-slate-100' },
  divider: { dark: 'border-slate-700', light: 'border-slate-200' },
} as const;
