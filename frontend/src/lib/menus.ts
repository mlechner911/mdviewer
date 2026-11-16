// Toolbar menus are <details class="mlc-menu">. A native <details> closes only
// by a click on its own summary, so it stays open while the user clicks the
// next button. These helpers close it on a click elsewhere, when another menu
// opens, and on Escape.

/** Closes every open menu that does not contain target. */
export function closeMenus(target: EventTarget | null = null, root: ParentNode = document): void {
  for (const d of root.querySelectorAll<HTMLDetailsElement>('details.mlc-menu[open]')) {
    if (!target || !d.contains(target as Node)) d.open = false;
  }
}

/** Closes open menus on Escape; true when one was open (the key is used up). */
export function closeMenusOnEscape(e: KeyboardEvent, root: ParentNode = document): boolean {
  if (e.key !== 'Escape' || !root.querySelector('details.mlc-menu[open]')) return false;
  closeMenus(null, root);
  return true;
}
