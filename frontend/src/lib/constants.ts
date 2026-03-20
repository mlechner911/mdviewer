

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
  },
  tab: {
    active: {
      dark: 'bg-slate-900 text-white border-t-2 border-t-blue-500',
      light: 'bg-white text-slate-900 border-t-2 border-t-blue-500'
    },
    inactive: {
      dark: 'bg-slate-800 text-slate-400 hover:bg-slate-700 border-t-2 border-t-transparent',
      light: 'bg-slate-200 text-slate-500 hover:bg-slate-100 border-t-2 border-t-transparent'
    }
  }
} as const;

export const DEFAULTS = {
  fontSize: 90
} as const;
