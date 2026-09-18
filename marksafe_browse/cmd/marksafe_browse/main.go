package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"marksafe_browse/internal/browse"
)

var (
	flagPath  string
	flagPort  int
	flagBind  string
	flagTheme string
	flagTree  bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "marksafe_browse [path]",
		Short: "MarkSafe Browse — Browse a Markdown directory in your browser",
		Long: `MarkSafe Browse starts a local web server to browse a Markdown directory.

The provided directory acts as a "hard boundary" — nothing outside it can be accessed.
If an index.md exists, it serves as the landing page. Otherwise, all .md files are
listed as a table of contents.`,
		Args: cobra.MaximumNArgs(1),
		Run: run,
	}

	rootCmd.Flags().StringVarP(&flagPath, "path", "p", "", "Root directory or index.md file")
	rootCmd.Flags().IntVarP(&flagPort, "port", "P", 8080, "Server port")
	rootCmd.Flags().StringVarP(&flagBind, "bind", "b", "127.0.0.1", "Bind address")
	rootCmd.Flags().StringVarP(&flagTheme, "theme", "t", "auto", "Theme: dark, light, or auto")
	rootCmd.Flags().BoolVar(&flagTree, "tree", false, "Show directory tree")

	rootCmd.MarkFlagRequired("path")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	// Determine root path
	root := flagPath
	if len(args) > 0 {
		root = args[0]
	}

	// Validate and create config
	cfg, err := browse.ValidateConfig(root, flagPort, flagBind, flagTheme, flagTree)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Print banner
	fmt.Println(cfg.ConfigString())

	// Verify root contains markdown files
	scanner, err := browse.NewScanner(cfg.Root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Scanner error: %v\n", err)
		os.Exit(1)
	}

	count, err := scanner.CountMarkdownFiles()
	if err != nil {
		log.Printf("Warning: could not count markdown files: %v", err)
	} else {
		log.Printf("Found %d markdown files", count)
	}

	if indexPath, found := scanner.FindIndex(); found {
		title, _ := browse.ExtractTitle(indexPath)
		log.Printf("Index page: index.md (title: %q)", title)
	} else {
		log.Printf("No index.md found — TOC page will list all .md files")
	}

	// Print TOC structure
	entries, err := scanner.Scan()
	if err != nil {
		log.Printf("Warning: could not scan directory: %v", err)
	} else {
		log.Printf("TOC (%d top-level entries):", len(entries))
		for _, e := range entries {
			log.Printf("  %s", formatEntry(e, 1))
		}
	}

	fmt.Printf("\n✅ Ready to serve %s on %s:%d\n", cfg.Root, cfg.Bind, cfg.Port)
	fmt.Printf("   Open http://%s:%d in your browser.\n", cfg.Bind, cfg.Port)
	fmt.Println("   (HTTP server starting — Phase 2)")
}

// formatEntry recursively formats a TOC entry for display.
func formatEntry(e browse.TOCEntry, indent int) string {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}
	if e.IsDir {
		return fmt.Sprintf("%s📁 %s/", prefix, e.Title)
	}
	return fmt.Sprintf("%s📄 %s", prefix, e.Title)
}
