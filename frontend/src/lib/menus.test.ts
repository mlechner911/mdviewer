// Toolbar menus close on a click elsewhere, when another opens, and on Escape.
import { describe, expect, it } from 'vitest';
import { closeMenus, closeMenusOnEscape } from './menus';

function menu(open: boolean) {
  const inside = {};
  return { open, inside, contains: (n: unknown) => n === inside };
}

function rootOf(...menus: ReturnType<typeof menu>[]) {
  return {
    querySelectorAll: () => menus.filter((m) => m.open),
    querySelector: () => menus.find((m) => m.open) ?? null,
  } as unknown as ParentNode;
}

describe('closeMenus', () => {
  it('closes an open menu on a click outside it', () => {
    const h = menu(true);
    closeMenus({} as EventTarget, rootOf(h));
    expect(h.open).toBe(false);
  });

  it('keeps the menu the click is in, closes the others', () => {
    const h = menu(true);
    const table = menu(true);
    closeMenus(table.inside as EventTarget, rootOf(h, table));
    expect(h.open).toBe(false);
    expect(table.open).toBe(true);
  });

  it('Escape closes and uses the key only when a menu was open', () => {
    const h = menu(true);
    expect(closeMenusOnEscape({ key: 'Escape' } as KeyboardEvent, rootOf(h))).toBe(true);
    expect(h.open).toBe(false);
    expect(closeMenusOnEscape({ key: 'Escape' } as KeyboardEvent, rootOf(h))).toBe(false);
    const b = menu(true);
    expect(closeMenusOnEscape({ key: 'a' } as KeyboardEvent, rootOf(b))).toBe(false);
    expect(b.open).toBe(true);
  });
});
