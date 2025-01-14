// Copyright (c) 2025 Neomantra Corp

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/NimbleMarkets/dbn-go"
	dbn_file "github.com/NimbleMarkets/dbn-go/internal/file"
	"github.com/spf13/cobra"
)

///////////////////////////////////////////////////////////////////////////////

var (
	verbose bool

	destDir string // destination directory

	forceZstdInput = false // force input to be zstd, irrespective of filename suffix
)

func requireNoError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}
}

func requireNoErrorWithoutPrint(err error) {
	if err != nil {
		os.Exit(1)
	}
}

///////////////////////////////////////////////////////////////////////////////

func main() {
	cobra.OnInitialize()

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	rootCmd.AddCommand(printMetadataCmd)
	printMetadataCmd.Flags().BoolVarP(&forceZstdInput, "zstd", "z", false, "Input is zstd (useful for handling zstd on stdin)")

	rootCmd.AddCommand(splitFilesCmd)
	splitFilesCmd.Flags().BoolVarP(&forceZstdInput, "zstd", "z", false, "Input is zstd (useful for handling zstd on stdin)")
	splitFilesCmd.Flags().StringVarP(&destDir, "dest", "d", "", "Destination directory")
	splitFilesCmd.MarkFlagRequired("dest")

	rootCmd.AddCommand(jsonPrintCmd)

	err := rootCmd.Execute()
	requireNoErrorWithoutPrint(err)
}

///////////////////////////////////////////////////////////////////////////////

var rootCmd = &cobra.Command{
	Use:   "dbn-go-file",
	Short: "dbn-go-file processes Databento DBN files",
	Long:  "dbn-go-file processes Databento DBN files",
}

///////////////////////////////////////////////////////////////////////////////

var printMetadataCmd = &cobra.Command{
	Use:   "metadata file...",
	Short: `Prints the specified file's metadata as JSON`,
	Long:  `Prints the specified file's metadata as JSON`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Run the split of the files
		for _, sourceFile := range args {
			if err := printMetadata(sourceFile, forceZstdInput); err != nil {
				fmt.Fprintf(os.Stderr, "error: reading %s: %s\n", sourceFile, err.Error())
			}
		}
	},
}

func printMetadata(sourceFile string, forceZstd bool) error {
	dbnFile, dbnCloser, err := dbn.MakeCompressedReader(sourceFile, forceZstd)
	defer dbnCloser.Close()

	dbnScanner := dbn.NewDbnScanner(dbnFile)
	metadata, err := dbnScanner.Metadata()
	if err != nil {
		return fmt.Errorf("scanner failed to read metadata: %w", err)
	}

	jstr, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	fmt.Printf("%s\n", jstr)
	return nil
}

///////////////////////////////////////////////////////////////////////////////

var jsonPrintCmd = &cobra.Command{
	Use:   "json file...",
	Short: `Prints the specified file's records as JSON`,
	Long:  `Prints the specified file's records as JSON`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Run the split of the files
		for _, sourceFile := range args {
			if err := dbn_file.WriteDbnFileAsJson(sourceFile, forceZstdInput, os.Stdout); err != nil {
				fmt.Fprintf(os.Stderr, "error: splitting %s: %s\n", sourceFile, err.Error())
			}
		}
	},
}

///////////////////////////////////////////////////////////////////////////////

var splitFilesCmd = &cobra.Command{
	Use:   "split file...",
	Short: `Splits Databento download folders into "<feed>/<instrument_id>/Y/M/D/feed-YMD.type.dbn.zst"`,
	Long: `Splits Databento download folders into "<feed>/<instrument_id>/Y/M/D/feed-YMD.type.dbn.zst"
For example, underlying symbols for front month for a given date
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// create destination directory if it doesn't exist
		if destDir == "" {
			fmt.Fprintf(os.Stderr, "error: --dest cannot be empty.  Use '.' for current directory.\n")
			os.Exit(1)
		}
		if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
			fmt.Fprintf(os.Stderr, "error: dest directory creation failed with: %s\n", err.Error())
			os.Exit(1)
		}

		// Run the split of the files
		for _, sourceFile := range args {
			if err := dbn_file.SplitFile(sourceFile, destDir, false, verbose); err != nil {
				fmt.Fprintf(os.Stderr, "error: splitting %s: %s\n", sourceFile, err.Error())
			}
		}
	},
}

///////////////////////////////////////////////////////////////////////////////
