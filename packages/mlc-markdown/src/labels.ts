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
  /** Document header: "Version", "signiert von {by} am {at}", kind names. */
  version: string;
  signedBy: string;
  signed: string;
  kinds: { sop: string; va: string; form: string; info: string };
}

export const LABELS: Record<Locale, Labels> = {
  de: {
    note: 'Hinweis', tip: 'Tipp', important: 'Wichtig', warning: 'Warnung', caution: 'Vorsicht', fillTip: 'Ausfülltipp', pageBreak: 'Seitenumbruch', prefill: 'wird aus einem Ereignis vorausgefüllt',
    version: 'Version', signedBy: 'Signiert von {by} am {at}', signed: 'Signiert am {at}',
    kinds: { sop: 'Standardarbeitsanweisung', va: 'Verfahrensanweisung', form: 'Formblatt', info: 'Information' },
  },
  en: {
    note: 'Note', tip: 'Tip', important: 'Important', warning: 'Warning', caution: 'Caution', fillTip: 'Filling tip', pageBreak: 'Page break', prefill: 'filled from an event',
    version: 'Version', signedBy: 'Signed by {by} on {at}', signed: 'Signed on {at}',
    kinds: { sop: 'Standard operating procedure', va: 'Procedure', form: 'Form', info: 'Information' },
  },
  es: {
    note: 'Nota', tip: 'Consejo', important: 'Importante', warning: 'Advertencia', caution: 'Precaución', fillTip: 'Consejo de relleno', pageBreak: 'Salto de página', prefill: 'se rellena desde un evento',
    version: 'Versión', signedBy: 'Firmado por {by} el {at}', signed: 'Firmado el {at}',
    kinds: { sop: 'Procedimiento normalizado de trabajo', va: 'Procedimiento', form: 'Formulario', info: 'Información' },
  },
  fr: {
    note: 'Remarque', tip: 'Astuce', important: 'Important', warning: 'Avertissement', caution: 'Attention', fillTip: 'Conseil de saisie', pageBreak: 'Saut de page', prefill: 'rempli à partir d’un événement',
    version: 'Version', signedBy: 'Signé par {by} le {at}', signed: 'Signé le {at}',
    kinds: { sop: 'Procédure opératoire normalisée', va: 'Procédure', form: 'Formulaire', info: 'Information' },
  },
};

export function labelsFor(locale?: string): Labels {
  const l = (locale ?? 'de').slice(0, 2).toLowerCase() as Locale;
  return LABELS[l] ?? LABELS.en;
}
