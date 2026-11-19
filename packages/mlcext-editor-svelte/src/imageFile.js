// Images for the image dialog without a desktop backend: the browser's file
// choice, read as a data: URL, and the Markdown for an image.

/**
 * @typedef {object} MlcImageChoice
 * @property {string} name File name, shown and used as the first description.
 * @property {number} size Bytes.
 * @property {boolean} embeddable The image can go into the document as data: URL.
 * @property {string} [reference] Path relative to the document, when a reference works (the PDF reads nothing outside the document's folder).
 * @property {string} [dataUri] The image itself, for embedding without a host.
 * @property {string} [path] The host's handle (MarkSafe: the file path).
 */

/**
 * Lets the user choose an image file; null when the choice is cancelled.
 * The image comes back as data: URL, so it can only be embedded.
 * @param {string} [accept]
 * @returns {Promise<MlcImageChoice | null>}
 */
export function chooseImageFile(accept = 'image/png,image/jpeg,image/gif,image/webp,image/svg+xml') {
  return new Promise((resolve, reject) => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = accept;
    input.addEventListener('cancel', () => resolve(null));
    input.addEventListener('change', () => {
      const file = input.files?.[0];
      if (!file) return resolve(null);
      const reader = new FileReader();
      reader.onerror = () => reject(reader.error);
      reader.onload = () => resolve({ name: file.name, size: file.size, embeddable: true, dataUri: String(reader.result) });
      reader.readAsDataURL(file);
    });
    input.click();
  });
}

/**
 * Markdown for an image: ![alt](target). A target with spaces or
 * parentheses goes into <…>; brackets in the description are escaped.
 * @param {string} alt
 * @param {string} target
 */
export function imageMarkdown(alt, target) {
  const a = alt.replace(/([\\[\]])/g, '\\$1');
  const t = /[\s()<>]/.test(target) && !target.startsWith('data:') ? `<${target.replace(/[<>]/g, encodeURIComponent)}>` : target;
  return `![${a}](${t})`;
}
