import { describe, expect, it } from 'vitest';
import { exampleValues, fieldHelp, localizedText, propertyInsertion, propertyOptions, type MlcSchema } from '../src/schemaHelp';

const schema: MlcSchema = {
  fields: {
    'probe.name': { type: 'text', description: { de: 'Name', en: 'Name EN' } },
    'probe.calibration.last': { type: 'date', requires: 'probe', description: { de: 'Letzte' } },
    'event.no': { type: 'text', requires: 'always', fixed: true, right: 'R' },
  },
  legacy: { ort: 'event.no' },
};

describe('localizedText', () => {
  it('falls back to English, German, then anything', () => {
    expect(localizedText({ fr: 'F', en: 'E', de: 'D' }, 'fr')).toBe('F');
    expect(localizedText({ en: 'E', de: 'D' }, 'fr')).toBe('E');
    expect(localizedText({ de: 'D' }, 'fr')).toBe('D');
    expect(localizedText({ it: 'I' }, 'fr')).toBe('I');
    expect(localizedText(undefined, 'fr')).toBe('');
  });
});

describe('propertyOptions', () => {
  it('filters by prefix, sorted, in the locale', () => {
    expect(propertyOptions(schema, [], 'probe.', 'en').map((o) => [o.name, o.detail])).toEqual([
      ['probe.calibration.last', 'Letzte'],
      ['probe.name', 'Name EN'],
    ]);
  });
  it('leaves out used names, except the one being typed', () => {
    expect(propertyOptions(schema, ['probe.name'], 'probe.', 'de').map((o) => o.name)).toEqual(['probe.calibration.last']);
    expect(propertyOptions(schema, ['probe.name'], 'probe.name', 'de').map((o) => o.name)).toEqual(['probe.name']);
  });
  it('is empty without a schema', () => {
    expect(propertyOptions(null, [], '', 'de')).toEqual([]);
  });
});

describe('propertyInsertion', () => {
  it('writes a whole field at the end of a line', () => {
    expect(propertyInsertion('a.b', 'date', '')).toBe('a.b TT.MM.JJJJ]');
    expect(propertyInsertion('a.b', 'time', '  ')).toBe('a.b HH:MM]');
    expect(propertyInsertion('a.b', 'text', '')).toBe(`a.b ${'_'.repeat(30)}]`);
  });
  it('keeps a bracket the editor has closed', () => {
    expect(propertyInsertion('a.b', 'date', ']')).toBe('a.b TT.MM.JJJJ');
  });
  it('changes only the name in a field with a body', () => {
    expect(propertyInsertion('a.b', 'date', ' ____]')).toBe('a.b');
  });
});

describe('fieldHelp', () => {
  it('describes a property', () => {
    expect(fieldHelp(schema, 'probe.calibration.last', 'de')).toEqual(['Letzte', 'Typ: date', 'Wert nur mit Bezug: probe']);
    expect(fieldHelp(schema, 'event.no', 'de')).toEqual(['Typ: text', 'Recht nötig: R', 'Beim Ausfüllen nicht änderbar']);
  });
  it('points an old name to its property', () => {
    expect(fieldHelp(schema, 'ort', 'de')?.[0]).toContain('event.no');
  });
  it('knows nothing else', () => {
    expect(fieldHelp(schema, 'eigen', 'de')).toBeNull();
    expect(fieldHelp(null, 'probe.name', 'de')).toBeNull();
  });
});

describe('examples', () => {
  const withExamples: MlcSchema = {
    fields: {
      'a.b': { type: 'text', example: { de: 'Kühlraum A', en: 'Cold room A' } },
      'a.c': { type: 'text' },
    },
  };
  it('collects the values in the locale', () => {
    expect(exampleValues(withExamples, 'en')).toEqual({ 'a.b': 'Cold room A' });
    expect(exampleValues(null, 'en')).toEqual({});
  });
  it('shows up in the help', () => {
    expect(fieldHelp(withExamples, 'a.b', 'de')).toContain('Beispiel: Kühlraum A');
  });
});
