import baseTheme from './themes/base.json';
import darkPreset from './themes/presets/dark.json';
import lightPreset from './themes/presets/light.json';

/**
 * Theme defines the structural and visual properties for a preview style.
 */
export interface Theme {
  id: 'dark' | 'light';
  name: string;
  chromaStyle: string;
  containerClass: string;
  proseClass: string;
  mermaidTheme: 'dark' | 'default';
  mermaidVars?: Record<string, string>;
  customCSS?: Record<string, string>;
}

/**
 * Helper to merge a preset with the base theme.
 */
function createTheme(preset: any): Theme {
  return {
    ...baseTheme,
    ...preset,
    mermaidVars: {
      ...baseTheme.mermaidVars,
      ...preset.mermaidVars
    },
    customCSS: {
      ...baseTheme.customCSS,
      ...preset.customCSS
    }
  } as Theme;
}

export const darkTheme: Theme = createTheme(darkPreset);
export const lightTheme: Theme = createTheme(lightPreset);

export const themes: Theme[] = [darkTheme, lightTheme];

export function getTheme(mode: 'dark' | 'light'): Theme {
  return mode === 'dark' ? darkTheme : lightTheme;
}
