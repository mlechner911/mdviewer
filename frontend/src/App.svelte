<script lang="ts">
  import { onMount } from 'svelte';
  import { RenderMarkdown, OpenFile, SaveFile } from '../wailsjs/go/main/App.js'
  import Preview from './components/Preview.svelte';

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
  let theme: 'light' | 'dark' | 'sepia' | 'monochrome' = 'dark';
  let fontSize: number = 90;
  
  // Resizable logic
  let splitWidth: number = 50; // percentage
  let isResizing = false;

  function startResizing() {
    isResizing = true;
  }

  function stopResizing() {
    isResizing = false;
  }

  function onMouseMove(event: MouseEvent) {
    if (!isResizing) return;
    splitWidth = (event.clientX / window.innerWidth) * 100;
    // Constraints
    if (splitWidth < 10) splitWidth = 10;
    if (splitWidth > 90) splitWidth = 90;
  }

  // Debounce logic
  let timeout: ReturnType<typeof setTimeout>;
  
  function debouncedUpdate(value: string) {
    clearTimeout(timeout);
    timeout = setTimeout(async () => {
      try {
        htmlContent = await RenderMarkdown(value);
      } catch (err) {
        console.error("Failed to render markdown:", err);
      }
    }, 300); // 300ms delay
  }

  async function handleOpen() {
    try {
      const content = await OpenFile();
      if (content !== undefined && content !== null) {
        markdown = content;
      }
    } catch (err) {
      console.error("Failed to open file:", err);
    }
  }

  async function handleSave() {
    try {
      await SaveFile(markdown);
    } catch (err) {
      console.error("Failed to save file:", err);
    }
  }

  function adjustFontSize(delta: number) {
    fontSize = Math.min(Math.max(fontSize + delta, 50), 200);
  }

  onMount(() => {
    debouncedUpdate(markdown);
  });

  $: if (markdown !== undefined) {
    debouncedUpdate(markdown);
  }
</script>

<svelte:window on:mousemove={onMouseMove} on:mouseup={stopResizing} />

<main class="flex h-screen w-full overflow-hidden bg-slate-900 text-slate-100 flex-col select-none">
  <!-- Toolbar -->
  <div class="h-12 bg-slate-800 border-b border-slate-700 flex items-center px-4 gap-4 shrink-0 z-20">
    <div class="flex gap-2">
      <button on:click={handleOpen} title="Open File" class="px-3 py-1 bg-slate-700 hover:bg-slate-600 rounded text-sm transition-colors">Open</button>
      <button on:click={handleSave} title="Save File" class="px-3 py-1 bg-slate-700 hover:bg-slate-600 rounded text-sm transition-colors">Save</button>
    </div>
    
    <div class="h-6 w-px bg-slate-700"></div>

    <div class="flex gap-2 items-center">
      <span class="text-xs text-slate-400">Theme:</span>
      <select bind:value={theme} class="bg-slate-700 text-xs rounded border-none focus:ring-1 focus:ring-blue-500 py-1 cursor-pointer">
        <option value="light">Light</option>
        <option value="dark">Dark</option>
        <option value="sepia">Sepia</option>
        <option value="monochrome">Monochrome</option>
      </select>
    </div>

    <div class="h-6 w-px bg-slate-700"></div>

    <div class="flex gap-1 items-center">
      <span class="text-xs text-slate-400 mr-1">Zoom:</span>
      <button on:click={() => adjustFontSize(-5)} class="w-6 h-6 flex items-center justify-center bg-slate-700 hover:bg-slate-600 rounded text-sm font-bold">-</button>
      <span class="text-[10px] text-slate-300 w-8 text-center font-mono">{fontSize}%</span>
      <button on:click={() => adjustFontSize(5)} class="w-6 h-6 flex items-center justify-center bg-slate-700 hover:bg-slate-600 rounded text-sm font-bold">+</button>
    </div>

    <div class="flex-1"></div>
    <span class="text-xs text-slate-500 font-mono">MD Viewer v0.1</span>
  </div>

  <div class="flex flex-1 overflow-hidden relative">
    <!-- Editor Column -->
    <div class="flex flex-col border-r border-slate-700 relative" style="width: {splitWidth}%;">
      <div class="p-2 bg-slate-800 text-xs font-bold uppercase tracking-wider text-slate-400 border-b border-slate-700 shrink-0">
        Editor
      </div>
      <textarea
        bind:value={markdown}
        spellcheck="false"
        autocorrect="off"
        autocapitalize="off"
        class="flex-1 p-4 bg-slate-900 text-slate-100 focus:outline-none resize-none font-mono text-sm select-text"
        placeholder="Type markdown here..."
      ></textarea>
      
      <!-- Splitter Handle -->
      <div 
        on:mousedown={startResizing}
        class="absolute top-0 right-0 w-1 h-full cursor-col-resize hover:bg-blue-500/50 transition-colors z-10"
      ></div>
    </div>

    <!-- Preview Column -->
    <div class="flex-1 flex flex-col relative">
      <!-- Overlay during resize to prevent iframe/select issues -->
      {#if isResizing}
        <div class="absolute inset-0 z-50"></div>
      {/if}
      <div class="p-2 bg-slate-800 text-xs font-bold uppercase tracking-wider text-slate-400 border-b border-slate-700 shrink-0">
        Preview
      </div>
      <Preview html={htmlContent} {theme} {fontSize} />
    </div>
  </div>
</main>

<style>
  :global(body) {
    user-select: none;
  }
</style>
