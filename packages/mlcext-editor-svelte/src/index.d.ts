export { default as MlcEditor } from './MlcEditor.svelte';
export { default as MlcFieldDialog, type MlcFieldSpec } from './MlcFieldDialog.svelte';
export { default as MlcImageDialog, type MlcImageRequest } from './MlcImageDialog.svelte';

/** CSS variables of the editor (--mlc-editor-<name>) and their fallbacks per theme. */
export declare const PALETTE: Record<'light' | 'dark', Record<string, string>>;

/** Built-in texts of the dialogs; a host passes `labels` for other languages. */
export declare const DIALOG_LABELS: Record<'de' | 'en', Record<string, string>>;
export declare function dialogTexts(locale: string, overrides?: Record<string, string>): (key: string, value?: string) => string;
export declare function formatBytes(n: number, locale: string): string;

/** A file for the image dialog. */
export interface MlcImageChoice {
  /** File name, shown and used as the first description. */
  name: string;
  size: number;
  /** The image can go into the document as data: URL. */
  embeddable: boolean;
  /** Path relative to the document, when a reference works. */
  reference?: string;
  /** The image itself, for embedding without a host. */
  dataUri?: string;
  /** The host's handle (MarkSafe: the file path). */
  path?: string;
}
/** The browser's file choice, read as data: URL; null when cancelled. */
export declare function chooseImageFile(accept?: string): Promise<MlcImageChoice | null>;
/** ![alt](target), with <…> around a target with spaces. */
export declare function imageMarkdown(alt: string, target: string): string;
