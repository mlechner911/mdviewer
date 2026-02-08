

export const APP_THEME = {
  DARK: 'dark',
  LIGHT: 'light',
  AUTO: 'auto'
} as const;

export type AppTheme_t = typeof APP_THEME[keyof typeof APP_THEME];

export const STYLE = {
  toolbar: {
    dark: 'bg-slate-800 border-slate-700 text-slate-100',
    light: 'bg-slate-100 border-slate-300 text-slate-900'
  },
  editor: {
    dark: 'bg-slate-900 text-slate-100 border-slate-700',
    light: 'bg-white text-slate-900 border-slate-300'
  },
  button: {
    dark: 'bg-slate-700 hover:bg-slate-600',
    light: 'bg-slate-200 hover:bg-slate-300'
  },
  divider: {
    dark: 'bg-slate-700',
    light: 'bg-slate-300'
  }
} as const;

export const DEFAULTS = {
  fontSize: 90
} as const;
