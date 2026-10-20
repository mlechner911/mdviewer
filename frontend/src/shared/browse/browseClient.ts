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

/** A document as Markdown source; path is the file read (a directory: its index.md). */
export interface DocSource {
  path: string;
  title: string;
  source: string;
  size?: number;
  modified?: string;
}

/** How a link or image target is treated (decided by the host, see mdbrowse.LinkTarget). */
export interface LinkTarget {
  kind: 'md' | 'file' | 'broken' | 'keep';
  path?: string;
  image?: boolean;
  reason?: string;
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
  /** File name is not the RefNo (a hint, nothing is renamed). */
  nameMismatch?: boolean;
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
  /**
   * Markdown source for client-side rendering (<MlcMarkdown>); null for a
   * directory without index.md (renderDoc shows its listing).
   */
  getSource(path: string, signal?: AbortSignal): Promise<DocSource | null>;
  /** Classify the link/image targets of docPath, keyed by href as written. */
  resolveLinks(docPath: string, hrefs: string[]): Promise<Record<string, LinkTarget>>;
  /** URL that serves a root-relative file (images, downloads). */
  rawUrl(path: string): string;
  /** Ranked full-text hits. Rejects on transport errors. */
  search(q: string): Promise<BrowseHit[]>;
  /** MLC documents below the root (never throws: [] on failure). */
  documents(): Promise<MlcDocument[]>;
  /** Syntax-highlighting CSS for the active theme (code colors). */
  chromaCss(theme: BrowseTheme): Promise<string>;
}
