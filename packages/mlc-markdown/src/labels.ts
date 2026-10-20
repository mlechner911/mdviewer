// Texts the renderer writes itself. Alert titles follow the language of the
// document (like mlcdoc); the rest follows the user interface.

export type Locale = 'de' | 'en' | 'es' | 'fr';

export interface Labels {
  note: string;
  tip: string;
  important: string;
  warning: string;
  caution: string;
  fillTip: string;
  pageBreak: string;
  prefill: string;
}

export const LABELS: Record<Locale, Labels> = {
  de: { note: 'Hinweis', tip: 'Tipp', important: 'Wichtig', warning: 'Warnung', caution: 'Vorsicht', fillTip: 'Ausfülltipp', pageBreak: 'Seitenumbruch', prefill: 'wird aus einem Ereignis vorausgefüllt' },
  en: { note: 'Note', tip: 'Tip', important: 'Important', warning: 'Warning', caution: 'Caution', fillTip: 'Filling tip', pageBreak: 'Page break', prefill: 'filled from an event' },
  es: { note: 'Nota', tip: 'Consejo', important: 'Importante', warning: 'Advertencia', caution: 'Precaución', fillTip: 'Consejo de relleno', pageBreak: 'Salto de página', prefill: 'se rellena desde un evento' },
  fr: { note: 'Remarque', tip: 'Astuce', important: 'Important', warning: 'Avertissement', caution: 'Attention', fillTip: 'Conseil de saisie', pageBreak: 'Saut de page', prefill: 'rempli à partir d’un événement' },
};

export function labelsFor(locale?: string): Labels {
  const l = (locale ?? 'de').slice(0, 2).toLowerCase() as Locale;
  return LABELS[l] ?? LABELS.en;
}
