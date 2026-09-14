// Thin adapter over @wailsio/runtime (Wails v3) that keeps the v2-style helper
// names the components were written against.
//
// v3 delivers one payload per event as `event.data`; the Go side only ever
// emits a single value, so callbacks receive that value directly.
import { Events, Browser } from '@wailsio/runtime';

const unsubscribers = new Map<string, Array<() => void>>();

export function EventsOn(eventName: string, callback: (data: any) => void): () => void {
  const off = Events.On(eventName, (e: { data?: unknown }) => callback(e?.data));
  const list = unsubscribers.get(eventName) ?? [];
  list.push(off);
  unsubscribers.set(eventName, list);
  return off;
}

export function EventsOff(...eventNames: string[]): void {
  for (const name of eventNames) {
    unsubscribers.get(name)?.forEach((off) => off());
    unsubscribers.delete(name);
  }
}

// Files dropped onto the window. In v3 the drop is resolved natively and
// forwarded by main.go as the "files-dropped" event; only elements carrying
// `data-file-drop-target` accept a drop.
const FILES_DROPPED = 'files-dropped';

export function OnFileDrop(callback: (x: number, y: number, paths: string[]) => void, _useDropTarget = false): void {
  EventsOn(FILES_DROPPED, (paths: string[]) => callback(0, 0, paths ?? []));
}

export function OnFileDropOff(): void {
  EventsOff(FILES_DROPPED);
}

export function BrowserOpenURL(url: string): void {
  void Browser.OpenURL(url);
}
