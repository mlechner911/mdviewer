// What an editor shows from a field schema (mlcext.FieldSchema). No rules here:
// whether a name is valid is the Go checker's answer. Kept free of CodeMirror
// so it can be tested and used by a host dialog.

// Maps as the Go bindings type them: nil or with entries missing.
type LocalizedText = { [language: string]: string | undefined };

export interface MlcSchemaField {
  type: string;
  requires?: string;
  description?: LocalizedText | null;
  example?: LocalizedText | null;
  fixed?: boolean;
  right?: string;
}

export interface MlcSchema {
  id?: string;
  title?: LocalizedText | null;
  fields?: { [name: string]: MlcSchemaField | undefined } | null;
  legacy?: { [name: string]: string | undefined } | null;
}

/** Text in the locale, else English, German or whatever there is. */
export function localizedText(text: LocalizedText | null | undefined, locale: string): string {
  if (!text) return '';
  return text[locale] || text.en || text.de || Object.values(text)[0] || '';
}

export interface MlcPropertyOption {
  name: string;
  type: string;
  detail: string;
}

/**
 * Properties to offer: those starting with typed, without names the document
 * already uses (a name may appear only once), except the one being typed.
 */
export function propertyOptions(
  schema: MlcSchema | null | undefined,
  used: string[],
  typed: string,
  locale: string,
): MlcPropertyOption[] {
  return Object.entries(schema?.fields ?? {})
    .filter(([name]) => name.startsWith(typed) && (name === typed || !used.includes(name)))
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([name, f]) => ({ name, type: f?.type ?? '', detail: localizedText(f?.description, locale) }));
}

const BODY: Record<string, string> = {
  date: 'TT.MM.JJJJ',
  time: 'HH:MM',
  datetime: 'TT.MM.JJJJ HH:MM',
  number: '__________',
};

/**
 * What completing a name inserts. In a field that already has a body only the
 * name changes; otherwise the body fits the type, and "]" is added unless the
 * editor has already closed the bracket.
 */
export function propertyInsertion(name: string, type: string, afterCursor: string): string {
  const after = afterCursor.trimStart();
  if (after !== '' && !after.startsWith(']')) return name;
  const body = BODY[type] ?? '_'.repeat(30);
  return `${name} ${body}${after === '' ? ']' : ''}`;
}

/** Lines to show for a field, or null when the schema does not know the name. */
export function fieldHelp(schema: MlcSchema | null | undefined, name: string, locale: string): string[] | null {
  if (!schema) return null;
  const replacement = schema.legacy?.[name];
  if (replacement) return [`Alte Form von ${replacement}; bleibt gültig.`];
  const f = schema.fields?.[name];
  if (!f) return null;
  const lines = [localizedText(f.description, locale), `Typ: ${f.type}`];
  const example = localizedText(f.example, locale);
  if (example) lines.push(`Beispiel: ${example}`);
  if (f.requires && f.requires !== 'always') lines.push(`Wert nur mit Bezug: ${f.requires}`);
  if (f.right) lines.push(`Recht nötig: ${f.right}`);
  if (f.fixed) lines.push('Beim Ausfüllen nicht änderbar');
  return lines.filter(Boolean);
}

/** Field name → example value, for a preview; properties without one are left out. */
export function exampleValues(schema: MlcSchema | null | undefined, locale: string): Record<string, string> {
  const out: Record<string, string> = {};
  for (const [name, f] of Object.entries(schema?.fields ?? {})) {
    const value = localizedText(f?.example, locale);
    if (value) out[name] = value;
  }
  return out;
}
