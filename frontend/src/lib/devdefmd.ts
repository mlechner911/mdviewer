/**
 * Per-locale welcome document content (body only — H1 title is from i18n).
 * It shows what MarkSafe does; the logo is embedded, as an example of an
 * embedded image.
 */
const LOGO = 'data:image/png;base64,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';

export const c_welcomeMarkdown: Record<'en' | 'de' | 'es' | 'fr', string> = {
	en: `

![MarkSafe](${LOGO} "logo.png")

**Turn Markdown into fillable PDF forms.**

MarkSafe is a private Markdown editor for the desktop. You write plain Markdown on the left; on the right you see the result at once, as HTML or exactly as the PDF prints.

> [!TIP]
> This page is an example: change something on the left and the preview follows. The logo above is embedded in the document. The editor shows it as 🖼 logo.png; a click shows the data, ▲ or Esc shortens it again.

## For SOPs, forms and documentation

- **Fillable forms:** \`[@#location ______]\` becomes a text field, \`[@#checked]\` a check box, \`[@#date TT.MM.JJJJ]\` a date field. No form designer needed – the field dialog writes the syntax for you (\`[_]\` in the toolbar, \`Ctrl+Shift+F\`).
- **Forms with a schema:** field names such as \`[@#probe.name]\` come from a field schema. The editor suggests them, explains them on hover and finds typos.
- **Fill in and keep a record:** fill in forms right in MarkSafe, with the document's text and images next to the fields. What gets saved is a record with who, when and which form version; the form itself stays unchanged.
- **Images:** insert them with 🖼, from the clipboard or by dropping a file. Embedded, a document stays complete even when files move.
- **PDF as printed:** a page header with kind, ID and version from the front matter, and an optional watermark such as DRAFT. Preview, export and printing come from the same PDF.
- **Templates:** new SOPs, procedures and forms – and also README, CHANGELOG or meeting minutes – start from templates you can change without an update.
- **Checks while you write:** broken fields, a missing ID or version and table columns that would be too narrow are reported right away, often with a quick fix.

## And everything Markdown is known for

- **GFM:** tables, task lists, footnotes and alert boxes such as \`> [!WARNING]\`.
- **Diagrams and formulas:** Mermaid and KaTeX right in the text, such as $e^{i\\pi} + 1 = 0$.

\`\`\`mermaid
graph LR
    A[Markdown] --> B(MarkSafe)
    B --> C[HTML preview]
    B --> D[PDF form]
    D --> E[Fill in with any PDF viewer or in MarkSafe]
\`\`\`

## Privacy first

Your documents stay on your machine. Images and links from outside load only after you allow them. MarkSafe does not track your usage and uploads nothing to a cloud.
`,
	de: `

![MarkSafe](${LOGO} "logo.png")

**Aus Markdown wird ein ausfüllbares PDF-Formular.**

MarkSafe ist ein privater Markdown-Editor für den Desktop. Links schreiben Sie einfaches Markdown, rechts sehen Sie sofort das Ergebnis, als HTML oder genau so, wie das PDF gedruckt wird.

> [!TIP]
> Diese Seite ist ein Beispiel: Ändern Sie links etwas, die Vorschau folgt sofort. Das Logo oben ist im Dokument eingebettet. Der Editor zeigt es als 🖼 logo.png; ein Klick zeigt die Daten, ▲ oder Esc kürzt sie wieder.

## Für SOPs, Formblätter und Dokumentation

- **Ausfüllbare Formulare:** \`[@#ort ______]\` wird ein Textfeld, \`[@#geprueft]\` ein Ankreuzfeld, \`[@#datum TT.MM.JJJJ]\` ein Datumsfeld. Ohne Formular-Designer – die Schreibweise übernimmt der Feld-Dialog (\`[_]\` in der Werkzeugleiste, \`Strg+Umschalt+F\`).
- **Formulare mit Schema:** Feldnamen wie \`[@#probe.name]\` kommen aus einem Feld-Schema. Der Editor schlägt sie vor, erklärt sie beim Darüberfahren und findet Tippfehler.
- **Ausfüllen mit Nachweis:** Formblätter direkt in MarkSafe ausfüllen, mit Text und Bildern des Dokuments neben den Feldern. Gespeichert wird ein Nachweis mit wer, wann und welcher Formblatt-Version; das Formblatt selbst bleibt unverändert.
- **Bilder:** einfügen mit 🖼, aus der Zwischenablage oder per Hineinziehen. Eingebettet bleibt ein Dokument vollständig, auch wenn Dateien verschoben werden.
- **PDF wie gedruckt:** Seitenkopf mit Art, Kennung und Version aus dem Front Matter, wahlweise ein Wasserzeichen wie ENTWURF. Vorschau, Export und Drucken kommen aus demselben PDF.
- **Vorlagen:** Neue SOPs, Verfahrensanweisungen und Formblätter, aber auch README, CHANGELOG oder Besprechungsprotokolle entstehen aus Vorlagen, die sich ohne Update anpassen lassen.
- **Prüfung beim Schreiben:** Fehlerhafte Felder, fehlende Kennung oder Version und zu schmale Tabellenspalten meldet der Editor sofort, oft mit Schnellkorrektur.

## Und alles, was Markdown ausmacht

- **GFM:** Tabellen, Tasklisten, Fußnoten und Hinweisboxen wie \`> [!WARNING]\`.
- **Diagramme und Formeln:** Mermaid und KaTeX direkt im Text, etwa $e^{i\\pi} + 1 = 0$.

\`\`\`mermaid
graph LR
    A[Markdown] --> B(MarkSafe)
    B --> C[HTML-Vorschau]
    B --> D[PDF-Formular]
    D --> E[Ausfüllen im PDF-Programm oder in MarkSafe]
\`\`\`

## Datenschutz an erster Stelle

Ihre Dokumente bleiben auf Ihrem Rechner. Bilder und Links von außen laden erst nach Ihrer Freigabe. MarkSafe verfolgt Ihre Nutzung nicht und lädt nichts in eine Cloud.
`,
	es: `

![MarkSafe](${LOGO} "logo.png")

**Convierta Markdown en formularios PDF rellenables.**

MarkSafe es un editor de Markdown privado para el escritorio. A la izquierda escribe Markdown sencillo; a la derecha ve el resultado al instante, como HTML o tal como se imprime el PDF.

> [!TIP]
> Esta página es un ejemplo: cambie algo a la izquierda y la vista previa lo sigue. El logotipo de arriba está incrustado en el documento. El editor lo muestra como 🖼 logo.png; un clic muestra los datos y ▲ o Esc los vuelve a acortar.

## Para SOP, formularios y documentación

- **Formularios rellenables:** \`[@#lugar ______]\` se convierte en un campo de texto, \`[@#revisado]\` en una casilla y \`[@#fecha TT.MM.JJJJ]\` en un campo de fecha. Sin diseñador de formularios: el diálogo de campos escribe la sintaxis (\`[_]\` en la barra, \`Ctrl+Mayús+F\`).
- **Formularios con esquema:** nombres de campo como \`[@#probe.name]\` vienen de un esquema de campos. El editor los sugiere, los explica al pasar el ratón y detecta errores.
- **Rellenar con registro:** rellene formularios en MarkSafe, con el texto y las imágenes del documento junto a los campos. Se guarda un registro con quién, cuándo y qué versión; el formulario no cambia.
- **Imágenes:** insértelas con 🖼, desde el portapapeles o arrastrando un archivo. Incrustadas, el documento sigue completo aunque se muevan los archivos.
- **PDF tal como se imprime:** encabezado con tipo, identificador y versión del front matter y, si se desea, una marca de agua como BORRADOR. Vista previa, exportación e impresión salen del mismo PDF.
- **Plantillas:** nuevos SOP, procedimientos y formularios, pero también README, CHANGELOG o actas, se crean a partir de plantillas que se pueden adaptar sin actualizar.
- **Comprobación al escribir:** campos erróneos, falta de identificador o versión y columnas demasiado estrechas se señalan enseguida, a menudo con corrección rápida.

## Y todo lo que caracteriza a Markdown

- **GFM:** tablas, listas de tareas, notas al pie y avisos como \`> [!WARNING]\`.
- **Diagramas y fórmulas:** Mermaid y KaTeX directamente en el texto, como $e^{i\\pi} + 1 = 0$.

\`\`\`mermaid
graph LR
    A[Markdown] --> B(MarkSafe)
    B --> C[Vista HTML]
    B --> D[Formulario PDF]
    D --> E[Rellenar en un visor PDF o en MarkSafe]
\`\`\`

## La privacidad, lo primero

Sus documentos se quedan en su equipo. Las imágenes y enlaces externos solo se cargan con su permiso. MarkSafe no rastrea su uso ni sube nada a la nube.
`,
	fr: `

![MarkSafe](${LOGO} "logo.png")

**Transformez Markdown en formulaires PDF remplissables.**

MarkSafe est un éditeur Markdown privé pour le bureau. À gauche, vous écrivez du Markdown simple ; à droite, vous voyez aussitôt le résultat, en HTML ou tel que le PDF sera imprimé.

> [!TIP]
> Cette page est un exemple : modifiez quelque chose à gauche, l'aperçu suit. Le logo ci-dessus est intégré au document. L'éditeur l'affiche comme 🖼 logo.png ; un clic montre les données, ▲ ou Échap les raccourcit à nouveau.

## Pour les SOP, formulaires et la documentation

- **Formulaires remplissables :** \`[@#lieu ______]\` devient un champ texte, \`[@#verifie]\` une case à cocher, \`[@#date TT.MM.JJJJ]\` un champ date. Sans concepteur de formulaires : la boîte de dialogue des champs écrit la syntaxe (\`[_]\` dans la barre, \`Ctrl+Maj+F\`).
- **Formulaires avec schéma :** des noms de champ comme \`[@#probe.name]\` viennent d'un schéma de champs. L'éditeur les propose, les explique au survol et repère les fautes.
- **Remplir avec preuve :** remplissez les formulaires dans MarkSafe, avec le texte et les images du document à côté des champs. Est enregistrée une preuve avec qui, quand et quelle version ; le formulaire reste inchangé.
- **Images :** insérez-les avec 🖼, depuis le presse-papiers ou par glisser-déposer. Intégré, un document reste complet même si des fichiers sont déplacés.
- **PDF tel qu'imprimé :** en-tête avec type, identifiant et version tirés du front matter, et au besoin un filigrane comme BROUILLON. Aperçu, export et impression viennent du même PDF.
- **Modèles :** nouvelles SOP, procédures et formulaires, mais aussi README, CHANGELOG ou comptes rendus, naissent de modèles modifiables sans mise à jour.
- **Vérification en cours de frappe :** champs erronés, identifiant ou version manquants et colonnes trop étroites sont signalés aussitôt, souvent avec une correction rapide.

## Et tout ce qui fait Markdown

- **GFM :** tableaux, listes de tâches, notes de bas de page et encadrés comme \`> [!WARNING]\`.
- **Diagrammes et formules :** Mermaid et KaTeX directement dans le texte, par exemple $e^{i\\pi} + 1 = 0$.

\`\`\`mermaid
graph LR
    A[Markdown] --> B(MarkSafe)
    B --> C[Aperçu HTML]
    B --> D[Formulaire PDF]
    D --> E[Remplir dans un lecteur PDF ou dans MarkSafe]
\`\`\`

## La confidentialité d'abord

Vos documents restent sur votre ordinateur. Les images et liens externes ne se chargent qu'avec votre accord. MarkSafe ne suit pas votre utilisation et n'envoie rien dans un cloud.
`,
};

/**
 * Detect browser locale fallback to 'en'. Called once at app start.
 */
export const c_getLocaleInitial = (): string => {
	if (typeof navigator === 'undefined') return 'en';
	const lang = navigator.language.split('-')[0].toLowerCase();
	return ['en', 'de', 'es', 'fr'].includes(lang) ? lang : 'en';
};
