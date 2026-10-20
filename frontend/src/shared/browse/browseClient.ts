// Shared browsing contract for every host (standalone web server,
// Wails desktop app, …). Components talk ONLY to this interface —
export interface TocEntry {
  path: string;
  title: string;
  children: TocEntry[] | null;
  isDir: boolean;
}

export interface BrowseDoc {
  title: string;
  html: string;
  size?: number;
  modified?: string;
}

export interface BrowseHit {
  path: string;
  title: string;
  score: number;
  snippet: string;
}

/** A document of the MLC Markdown Extension (non-standard) in the root. */
export interface MlcDocument {
  path: string;
  refNo: string;
  kind: string;
  kindLabel: string;
  version: string;
  title: string;
  fields: number;
  errors: number;
  warnings: number;
}

export type BrowseTheme = 'dark' | 'light';

export interface BrowseClient {
  /** Directory tree (never throws: resolves [] on failure, like before). */
  getTree(): Promise<TocEntry[]>;
  /**
   * Render one document. Rejects on transport/HTTP errors.
   * The optional signal lets hosts cancel superseded loads;
   * implementations that cannot cancel may ignore it.
   */
  renderDoc(path: string, theme: BrowseTheme, signal?: AbortSignal): Promise<BrowseDoc>;
  /** Ranked full-text hits. Rejects on transport errors. */
  search(q: string): Promise<BrowseHit[]>;
  /** MLC documents below the root (never throws: [] on failure). */
  documents(): Promise<MlcDocument[]>;
  /** Syntax-highlighting CSS for the active theme (code colors). */
  chromaCss(theme: BrowseTheme): Promise<string>;
}
