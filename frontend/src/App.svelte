<script lang="ts">
  import { onMount } from 'svelte';
  import { RenderMarkdown, OpenFile, SaveFile } from '../wailsjs/go/main/App.js'
  import Preview from './components/Preview.svelte';
  import { themes } from './themes';

  let markdown: string = `# Welcome to MD Viewer

## Mermaid Diagrams
\`\`\`mermaid
graph TD
    A[Start] --> B{Is it working?}
    B -- Yes --> C[Great!]
    B -- No --> D[Check Logs]
\`\`\`

## Feature Verification Table

| Feature | Support | Engine | Notes |
| :--- | :---: | :--- | :--- |
| **GFM** | ✅ | Goldmark | Tables, Tasklists |
| **Theming** | ✅ | Svelte | Dark, Light, Sepia |
| **Mermaid** | ✅ | Mermaid.js | Dynamic Rendering |
| **Highlight** | ✅ | Chroma | Inline CSS |

## Language Gallery

### Java
\`\`\`java
public class HelloWorld {
    public static void main(String[] args) {
        System.out.println("Hello, Java!");
    }
}
\`\`\`
`;
  let htmlContent: string = "";
  let currentPreviewTheme = themes[0];
  let fontSize: number = 90;
  
  // App Frame Theming
  type AppTheme = 'dark' | 'light' | 'auto';
  let appTheme: AppTheme = 'dark';
  let effectiveAppTheme: 'dark' | 'light' = 'dark';

  function updateEffectiveTheme() {
    if (appTheme === 'auto') {
      effectiveAppTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    } else {
      effectiveAppTheme = appTheme;
    }
  }

  // Resizable logic
  let splitWidth: number = 50;
  let isResizing = false;

  function startResizing() { isResizing = true; }
  function stopResizing() { isResizing = false; }
  function onMouseMove(event: MouseEvent) {
    if (!isResizing) return;
    splitWidth = (event.clientX / window.innerWidth) * 100;
    if (splitWidth < 10) splitWidth = 10;
    if (splitWidth > 90) splitWidth = 90;
  }

  // Debounce logic
  let timeout: ReturnType<typeof setTimeout>;
  
  function debouncedUpdate(value: string, themeStyle: string) {
    clearTimeout(timeout);
    timeout = setTimeout(async () => {
      try {
        htmlContent = await RenderMarkdown(value, themeStyle);
      } catch (err) {
        console.error("Failed to render markdown:", err);
      }
    }, 300);
  }

  async function handleOpen() {
    try {
      const content = await OpenFile();
      if (content !== undefined && content !== null) { markdown = content; }
    } catch (err) { console.error("Failed to open file:", err); }
  }

  async function handleSave() {
    try { await SaveFile(markdown); } catch (err) { console.error("Failed to save file:", err); }
  }

  function adjustFontSize(delta: number) {
    fontSize = Math.min(Math.max(fontSize + delta, 50), 200);
  }

  onMount(() => {
    updateEffectiveTheme();
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handler = () => { if (appTheme === 'auto') updateEffectiveTheme(); };
    mediaQuery.addEventListener('change', handler);
    debouncedUpdate(markdown, currentPreviewTheme.chromaStyle);
    return () => mediaQuery.removeEventListener('change', handler);
  });

  $: if (appTheme) updateEffectiveTheme();
  $: if (markdown !== undefined || currentPreviewTheme !== undefined) {
    debouncedUpdate(markdown, currentPreviewTheme.chromaStyle);
  }

  // Dynamic classes for the App Frame
  $: toolbarClass = effectiveAppTheme === 'dark' ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-slate-100 border-slate-300 text-slate-900';
  $: editorClass = effectiveAppTheme === 'dark' ? 'bg-slate-900 text-slate-100 border-slate-700' : 'bg-white text-slate-900 border-slate-300';
  $: buttonClass = effectiveAppTheme === 'dark' ? 'bg-slate-700 hover:bg-slate-600' : 'bg-slate-200 hover:bg-slate-300';
  $: dividerClass = effectiveAppTheme === 'dark' ? 'bg-slate-700' : 'bg-slate-300';
</script>

<svelte:window on:mousemove={onMouseMove} on:mouseup={stopResizing} />

<main class="flex h-screen w-full overflow-hidden flex-col select-none {effectiveAppTheme === 'dark' ? 'bg-slate-900' : 'bg-white'}">
  <!-- Toolbar -->
  <div class="h-12 border-b flex items-center px-4 gap-4 shrink-0 z-20 {toolbarClass}">
    <div class="flex gap-2">
      <button on:click={handleOpen} title="Open" class="px-3 py-1 rounded text-sm transition-colors {buttonClass}">Open</button>
      <button on:click={handleSave} title="Save" class="px-3 py-1 rounded text-sm transition-colors {buttonClass}">Save</button>
    </div>
    
    <div class="h-6 w-px {dividerClass}"></div>

    <div class="flex gap-2 items-center">
      <span class="text-xs opacity-60">App:</span>
      <select bind:value={appTheme} class="text-xs rounded border-none py-1 cursor-pointer bg-transparent focus:ring-1 focus:ring-blue-500">
        <option value="auto">Auto</option>
        <option value="dark">Dark</option>
        <option value="light">Light</option>
      </select>
    </div>

    <div class="h-6 w-px {dividerClass}"></div>

    <div class="flex gap-2 items-center">
      <span class="text-xs opacity-60">Preview:</span>
      <select bind:value={currentPreviewTheme} class="text-xs rounded border-none py-1 cursor-pointer bg-transparent focus:ring-1 focus:ring-blue-500">
        {#each themes as theme}
          <option value={theme}>{theme.name}</option>
        {/each}
      </select>
    </div>

    <div class="h-6 w-px {dividerClass}"></div>

    <div class="flex gap-1 items-center">
      <span class="text-xs opacity-60 mr-1">Zoom:</span>
      <button on:click={() => adjustFontSize(-5)} class="w-6 h-6 flex items-center justify-center rounded text-sm font-bold {buttonClass}">-</button>
      <span class="text-[10px] opacity-60 w-8 text-center font-mono">{fontSize}%</span>
      <button on:click={() => adjustFontSize(5)} class="w-6 h-6 flex items-center justify-center rounded text-sm font-bold {buttonClass}">+</button>
    </div>

    <div class="flex-1"></div>
    <span class="text-xs opacity-40 font-mono">MD Viewer v0.1</span>
  </div>

  <div class="flex flex-1 overflow-hidden relative">
    <!-- Editor Column -->
    <div class="flex flex-col border-r relative {editorClass}" style="width: {splitWidth}%;">
      <div class="p-2 text-xs font-bold uppercase tracking-wider opacity-50 border-b shrink-0 {toolbarClass}">
        Editor
      </div>
      <textarea
        bind:value={markdown}
        spellcheck="false" autocorrect="off" autocapitalize="off"
        class="flex-1 p-4 focus:outline-none resize-none font-mono text-sm select-text bg-transparent"
        placeholder="Type markdown here..."
      ></textarea>
      <div on:mousedown={startResizing} class="absolute top-0 right-0 w-1 h-full cursor-col-resize hover:bg-blue-500/50 transition-colors z-10"></div>
    </div>

    <!-- Preview Column -->
    <div class="flex-1 flex flex-col relative">
      {#if isResizing} <div class="absolute inset-0 z-50"></div> {/if}
      <div class="p-2 text-xs font-bold uppercase tracking-wider opacity-50 border-b shrink-0 {toolbarClass}">
        Preview
      </div>
      <Preview html={htmlContent} theme={currentPreviewTheme} fontSize={fontSize} />
    </div>
  </div>
</main>

<style>
  :global(body) { user-select: none; margin: 0; }
  select option {
    background-color: white;
    color: black;
  }
  :global(.dark) select option {
    background-color: #1e293b;
    color: white;
  }
</style>