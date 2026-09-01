<script lang="ts">
import { browser } from '$app/environment';
import { menuVisible, theme } from '../lib/stores';
import { getBackend } from '../lib/backend';

interface HamburgerMenuProps {
	onFileNew?: () => void;
	onFileOpenLocal?: () => void;
	onFileOpenRecent?: () => void;
	onEditCopyAsMarkdown?: () => void;
	onEditCopyAsHtml?: () => void;
	onEditCopyAsRTF?: () => void;
	onEditCopyImage?: () => void;
	onEditCopyImageText?: () => void;
	onFormatToggle?: () => void;
	onFormatWordCount?: () => void;
	onViewZoomIn?: () => void;
	onViewZoomOut?: () => void;
	onViewResetZoom?: () => void;
	onHelpAbout?: () => void;
	onHelpShowShortcuts?: () => void;
	onOpenSettings?: () => void;
	onAboutWails?: () => void;
	onOpenProductPage?: () => void;
}

let {
	onFileNew,
	onFileOpenLocal,
	onFileOpenRecent,
	onEditCopyAsMarkdown,
	onEditCopyAsHtml,
	onEditCopyAsRTF,
	onEditCopyImage,
	onEditCopyImageText,
	onFormatToggle,
	onFormatWordCount,
	onViewZoomIn,
	onViewZoomOut,
	onViewResetZoom,
	onHelpAbout,
	onHelpShowShortcuts,
	onOpenSettings,
	onAboutWails,
	onOpenProductPage,
} = $props();

let isOpen = $state(false);

function closeMenu() {
	isOpen = false;
	menuVisible.set(false);
}

function handleMenuEvent(action: string) {
	if (!browser) return;
	
	const backend = getBackend();
	
	switch(action) {
		case 'file-new':
			onFileNew?.();
			break;
		case 'file-open-local':
			onFileOpenLocal?.();
			break;
		case 'file-open-recent':
			onFileOpenRecent?.();
			break;
		case 'edit-copy-markdown':
			onEditCopyAsMarkdown?.();
			break;
		case 'edit-copy-html':
			onEditCopyAsHtml?.();
			break;
		case 'edit-copy-rtf':
			onEditCopyAsRTF?.();
			break;
		case 'edit-copy-image':
			onEditCopyImage?.();
			break;
		case 'edit-copy-image-text':
			onEditCopyImageText?.();
			break;
		case 'format-toggle':
			onFormatToggle?.();
			break;
		case 'format-word-count':
			onFormatWordCount?.();
			break;
		case 'view-zoom-in':
			onViewZoomIn?.();
			break;
		case 'view-zoom-out':
			onViewZoomOut?.();
			break;
		case 'view-reset-zoom':
			onViewResetZoom?.();
			break;
		case 'help-about':
			onHelpAbout?.();
			break;
		case 'help-show-shortcuts':
			onHelpShowShortcuts?.();
			break;
		case 'settings':
			onOpenSettings?.();
			break;
		case 'about-wails':
			onAboutWails?.();
			break;
		case 'product-page':
			onOpenProductPage?.();
			break;
	}
	
	closeMenu();
}

const menuSections = [
	{
		name: 'file',
		icon: 'file',
		label: 'datei',
		items: [
			{ action: 'file-new', icon: 'file-plus', label: 'neu' },
			{ action: 'file-open-local', icon: 'folder-open', label: 'lokale-datei-open' },
			{ action: 'file-open-recent', icon: 'clock', label: 'zuletzt-offen' },
		]
	},
	{
		name: 'edit',
		icon: 'edit',
		label: 'bearbeiten',
		items: [
			{ action: 'edit-copy-markdown', icon: 'copy', label: 'Als markdown kopieren' },
			{ action: 'edit-copy-html', icon: 'code', label: 'Als html kopieren' },
			{ action: 'edit-copy-rtf', icon: 'file-text', label: 'Als rtf kopieren' },
			{ action: 'edit-copy-image', icon: 'image', label: 'Als bild kopieren' },
			{ action: 'edit-copy-image-text', icon: 'copy-code', label: 'bildtext kopieren' },
		]
	},
	{
		name: 'format',
		icon: 'type',
		label: 'format',
		items: [
			{ action: 'format-toggle', icon: 'columns', label: 'seitenlayout-umschalten' },
			{ action: 'format-word-count', icon: 'hash', label: 'woerterzaehler' },
		]
	},
	{
		name: 'view',
		icon: 'eye',
		label: 'ansicht',
		items: [
			{ action: 'view-zoom-in', icon: 'zoom-in', label: 'vergroessern' },
			{ action: 'view-zoom-out', icon: 'zoom-out', label: 'verkleinern' },
			{ action: 'view-reset-zoom', icon: 'zoom-reset', label: 'zoom-zuruecksetzen' },
		]
	},
	{
		name: 'help',
		icon: 'help-circle',
		label: 'hilfe',
		items: [
			{ action: 'help-show-shortcuts', icon: 'keyboard', label: 'shortcuts-show' },
			{ action: 'settings', icon: 'settings', label: 'einstellungen' },
			{ action: 'help-about', icon: 'info', label: 'ueber-marksafe' },
		]
	},
];
</script>

<div class="relative">
	<!-- Hamburger Button -->
	<button
		class="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
		on:click={() => {
			isOpen = !isOpen;
			menuVisible.set(!isOpen);
		}}
	:aria-expanded={isOpen}
	aria-label="menu"
	title="Menü"
>
	{#if isOpen}
		<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
		</svg>
	{:else}
		<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
		</svg>
	{/if}
</button>

<!-- Dropdown Menu -->
{#if isOpen}
	<div class="fixed inset-0 z-40" on:click={closeMenu} />
	<div class="fixed top-12 left-4 right-4 md:left-auto md:right-auto md:top-12 md:min-w-[280px] z-50 bg-white dark:bg-gray-800 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-700 max-h-[70vh] overflow-y-auto">
		<div class="p-2">
			{#each menuSections as section}
				<div class="mb-2 last:mb-0">
					<div class="px-3 py-1.5 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">
						{i18n[section.label]}
					</div>
					<div class="mt-1 space-y-0.5">
						{#each section.items as item}
							<button
								class="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors text-left"
								on:click={() => handleMenuEvent(item.action)}
							>
								<span class="w-4 h-4 flex items-center justify-center">
									{#if item.icon === 'file-plus'}
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H4a2 2 0 01-2-2z" />
										</svg>
									{:else if item.icon === 'folder-open'}
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 19a2 2 0 01-2-2V7a2 2 0 012-2h4l2 2h4a2 2 0 012 2v1M5 19h14a2 2 0 002-2v-5a2 2 0 00-2-2H9a2 2 0 00-2 2v5a2 2 0 01-2 2z" />
										</svg>
									{:else if item.icon === 'clock'}
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
										</svg>
									{:else if item.icon === 'copy'}
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 012-2v-8a2 2 0 01-2-2h-8a2 2 0 01-2 2v8a2 2 0 012 2z" />
										</svg>
									{:else if item.icon === 'code'}
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
										</svg>
									{:else if item.icon === 'file-text'}
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
										</svg>
									{:else if item.icon === 'image'}
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
										</svg>
									{:else if item.icon === 'copy-code'}
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M16 10h.01M12 10h.01M9 16h6M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
										</svg>
									{:else if item.icon === 'columns'}
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 01-2 2h-2a2 2 0 01-2-2M9 7a2 2 0 012-2h2a2 2 0 012 2v10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 01-2-2h-2a2 2 0 01-2 2z" />
										</svg>
									{:else if item.icon === 'hash'}
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
										</svg>
									{:else if item.icon === 'zoom-in'}
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-2a7 7 0 11-14 0 7 7 0 0114 0zM15 9a5 5 0 11-10 0 5 5 0 0110 0z" />
										</svg>
									{:else if item.icon === 'zoom-out'}
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-2a7 7 0 11-14 0 7 7 0 0114 0zM5 9a5 5 0 11-10 0 5 5 0 0110 0z" />
										</svg>
									{:else if item.icon === 'zoom-reset'}
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 4l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
										</svg>
									{:else if item.icon === 'keyboard'}
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2zM3 10h18" />
										</svg>
									{:else if item.icon === 'settings'}
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
										</svg>
									{:else}
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
										</svg>
									{/if}
								</span>
								<span>{i18n[item.label]}</span>
							</button>
						{/each}
					</div>
				</div>
			{/each}
		</div>
		
		<div class="border-t border-gray-200 dark:border-gray-700 p-2 mt-2">
			<div class="space-y-0.5">
				<button
					class="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors text-left"
					on:click={() => handleMenuEvent('product-page')}
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
					</svg>
					<span>Website öffnen</span>
				</button>
				<button
					class="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors text-left"
					on:click={() => handleMenuEvent('about-wails')}
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z" />
					</svg>
					<span>Über MarkSafe</span>
				</button>
			</div>
		</div>
	</div>
{/if}
</div>
