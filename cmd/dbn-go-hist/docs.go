// Copyright (c) 2024 Neomantra Corp

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	docsOutputDir        string
	docsEnableAutoGenTag bool
	docsHugo             bool
)

var docsCmd = &cobra.Command{
	Use:    "docs",
	Short:  "Generate documentation for dbn-go-hist",
	Hidden: true,
	Long: `Generate documentation for all dbn-go-hist commands.

Subcommands:
  markdown  Generate plain markdown (default)
  man       Generate man pages

The auto-generation tag (timestamp footer) is disabled by default for stable,
reproducible files. Use --enableAutoGenTag for publishing.

Examples:
  dbn-go-hist docs                                                # Generate markdown docs in ./docs/
  dbn-go-hist docs markdown -o ./wiki                             # Generate markdown in custom directory
  dbn-go-hist docs markdown --hugo -o docs/hugo/content/command   # Generate Hugo-compatible docs
  dbn-go-hist docs man -o docs/man                                # Generate man pages`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDocsMarkdown(cmd, args)
	},
}

var docsMarkdownCmd = &cobra.Command{
	Use:   "markdown",
	Short: "Generate markdown documentation",
	Long: `Generate markdown documentation for all dbn-go-hist commands.

By default, generates plain markdown suitable for GitHub wikis and basic
documentation sites. Use --hugo to generate markdown with YAML front matter
for Hugo static site generator.`,
	RunE: runDocsMarkdown,
}

var docsManCmd = &cobra.Command{
	Use:   "man",
	Short: "Generate man pages",
	Long: `Generate man pages for all dbn-go-hist commands.

Man pages are generated in roff format suitable for installation
in /usr/share/man/man1 or /usr/local/share/man/man1.`,
	RunE: runDocsMan,
}

func runDocsMarkdown(cmd *cobra.Command, args []string) error {
	if err := os.MkdirAll(docsOutputDir, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	rootCmd.DisableAutoGenTag = !docsEnableAutoGenTag

	if docsHugo {
		prepender := func(filename string) string {
			name := filepath.Base(filename)
			name = strings.TrimSuffix(name, filepath.Ext(name))
			title := strings.ReplaceAll(name, "_", " ")
			return fmt.Sprintf(`---
title: "%s"
---

`, title)
		}
		linkHandler := func(name string) string {
			return name
		}

		if err := doc.GenMarkdownTreeCustom(rootCmd, docsOutputDir, prepender, linkHandler); err != nil {
			return fmt.Errorf("generate markdown: %w", err)
		}
		fmt.Print("(Hugo front matter enabled)\n")
	} else {
		if err := doc.GenMarkdownTree(rootCmd, docsOutputDir); err != nil {
			return fmt.Errorf("generate markdown: %w", err)
		}
	}

	count := countFiles(docsOutputDir, ".md")
	fmt.Printf("Generated %d markdown files in %s\n", count, docsOutputDir)
	return nil
}

func runDocsMan(cmd *cobra.Command, args []string) error {
	if err := os.MkdirAll(docsOutputDir, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	rootCmd.DisableAutoGenTag = !docsEnableAutoGenTag

	header := &doc.GenManHeader{
		Title:   "DBN-GO-HIST",
		Section: "1",
	}
	if err := doc.GenManTree(rootCmd, header, docsOutputDir); err != nil {
		return fmt.Errorf("generate man pages: %w", err)
	}

	count := countFiles(docsOutputDir, ".1")
	fmt.Printf("Generated %d man pages in %s\n", count, docsOutputDir)
	return nil
}

func countFiles(dir, ext string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	var count int
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ext {
			count++
		}
	}
	return count
}
