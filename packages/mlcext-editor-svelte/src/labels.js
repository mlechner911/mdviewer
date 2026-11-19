// Texts of the dialogs, German and English built in (MarkSafe's wording).
// A host with other languages passes `labels` (all or some keys); "%s"
// stands for the inserted value.

/** @type {Record<'de' | 'en', Record<string, string>>} */
export const DIALOG_LABELS = {
  de: {
    cancel: 'Abbrechen',
    fieldApplyEdit: 'Übernehmen',
    fieldApplyInsert: 'Einfügen',
    fieldDialogEdit: 'Formularfeld bearbeiten',
    fieldDialogInsert: 'Formularfeld einfügen',
    fieldKind: 'Art',
    fieldKindText: 'Textfeld',
    fieldKindMultiline: 'Mehrzeilig',
    fieldKindCheck: 'Ankreuzfeld',
    fieldKindDate: 'Datum',
    fieldKindTime: 'Uhrzeit',
    fieldKindPattern: 'Eigenes Muster',
    fieldKindSignature: 'Unterschrift',
    fieldName: 'Name',
    fieldNameHint: 'leer: ohne Namen (nummeriert nach der Überschrift)',
    fieldNameDuplicate: 'Den Namen gibt es schon in diesem Dokument.',
    fieldNamePrefill: 'Wird aus einem Ereignis vorausgefüllt: %s',
    fieldChars: 'Breite (Zeichen)',
    fieldRows: 'Zeilen',
    fieldFillPage: 'bis zum Seitenende',
    fieldPattern: 'Muster',
    fieldRole: 'Wer unterschreibt',
    fieldChecked: 'angekreuzt',
    fieldValue: 'Vorgabe',
    fieldTooltip: 'Tooltip',
    fieldSyntax: 'Schreibweise',
    fieldPreview: 'Vorschau',
    imageInsert: 'Bild einfügen',
    imageEditTitle: 'Bild bearbeiten',
    imageFile: 'Datei',
    imagePasted: 'Aus der Zwischenablage',
    imageEmbedded: 'eingebettet',
    imageChoose: 'Wählen…',
    imageChooseOther: 'Anderes Bild…',
    imageAlt: 'Beschreibung',
    imageSize: 'Größe',
    imageSizeHint: 'Breite in px, Rahmen oder % der Textbreite; leer = eigene Größe',
    imageSizeInvalid: 'z. B. 200, 200x100 oder 50 % (1–100)',
    imageHow: 'Einfügen als',
    imageReference: 'Verweis auf die Datei',
    imageEmbed: 'Eingebettet im Dokument',
    imageEmbedOnly:
      'Das Bild liegt außerhalb des Dokumentordners, oder das Dokument ist noch nicht gespeichert: Das PDF kann es nur eingebettet lesen.',
    imageLarge: 'Großes Bild (%s): Das Dokument wird dadurch langsamer.',
  },
  en: {
    cancel: 'Cancel',
    fieldApplyEdit: 'Apply',
    fieldApplyInsert: 'Insert',
    fieldDialogEdit: 'Edit form field',
    fieldDialogInsert: 'Insert form field',
    fieldKind: 'Kind',
    fieldKindText: 'Text field',
    fieldKindMultiline: 'Several lines',
    fieldKindCheck: 'Check box',
    fieldKindDate: 'Date',
    fieldKindTime: 'Time',
    fieldKindPattern: 'Own pattern',
    fieldKindSignature: 'Signature',
    fieldName: 'Name',
    fieldNameHint: 'empty: no name (numbered after the heading)',
    fieldNameDuplicate: 'This name is already used in the document.',
    fieldNamePrefill: 'Filled in from an event: %s',
    fieldChars: 'Width (characters)',
    fieldRows: 'Lines',
    fieldFillPage: 'to the end of the page',
    fieldPattern: 'Pattern',
    fieldRole: 'Who signs',
    fieldChecked: 'ticked',
    fieldValue: 'Preset',
    fieldTooltip: 'Tooltip',
    fieldSyntax: 'Syntax',
    fieldPreview: 'Preview',
    imageInsert: 'Insert image',
    imageEditTitle: 'Edit image',
    imageFile: 'File',
    imagePasted: 'From the clipboard',
    imageEmbedded: 'embedded',
    imageChoose: 'Choose…',
    imageChooseOther: 'Other image…',
    imageAlt: 'Description',
    imageSize: 'Size',
    imageSizeHint: 'px width, box or % of the text width; empty = own size',
    imageSizeInvalid: 'e.g. 200, 200x100 or 50% (1–100)',
    imageHow: 'Insert as',
    imageReference: 'Reference to the file',
    imageEmbed: 'Embedded in the document',
    imageEmbedOnly: 'The image is outside the document folder or the document is not saved yet: the PDF can only read it embedded.',
    imageLarge: 'Large image (%s): the document gets slower.',
  },
};

/**
 * The texts for locale ("de…" German, else English), with the host's
 * overrides; tr(key, value) fills "%s".
 * @param {string} locale
 * @param {Record<string, string>} [overrides]
 */
export function dialogTexts(locale, overrides = {}) {
  const base = DIALOG_LABELS[locale.toLowerCase().startsWith('de') ? 'de' : 'en'];
  const all = { ...base, ...overrides };
  /** @param {string} key @param {string} [value] */
  return (key, value) => {
    const text = all[key] ?? key;
    return value === undefined ? text : text.replace('%s', value);
  };
}

/**
 * Bytes as "1.2 MB", "340 KB" or "12 B" in the locale's number format.
 * @param {number} n
 * @param {string} locale
 */
export function formatBytes(n, locale) {
  /** @param {number} v */
  const fmt = (v) => v.toLocaleString(locale, { maximumFractionDigits: 1 });
  if (n >= 1024 * 1024) return `${fmt(n / (1024 * 1024))} MB`;
  if (n >= 1024) return `${fmt(n / 1024)} KB`;
  return `${n} B`;
}
