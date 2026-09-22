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
}
