import baseTheme from './themes/base.json';
import darkPreset from './themes/presets/dark.json';
import lightPreset from './themes/presets/light.json';
import sepiaPreset from './themes/presets/sepia.json';
import monochromePreset from './themes/presets/monochrome.json';

/**
 * Theme defines the structural and visual properties for a preview style.
 */
export interface Theme {
  id: string;
  name: string;
  chromaStyle: string;
  containerClass: string;
  proseClass: string;
  mermaidTheme: 'dark' | 'default' | 'neutral' | 'forest';
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

export const themes: Theme[] = [
  createTheme(darkPreset),
  createTheme(lightPreset),
  createTheme(sepiaPreset),
  createTheme(monochromePreset)
];
