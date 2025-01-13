// Copyright (c) 2024 Neomantra Corp
//
// NOTE: this may incur billing, handle with care!
//

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/NimbleMarkets/dbn-go"
	dbn_hist "github.com/NimbleMarkets/dbn-go/hist"
	dbn_tui "github.com/NimbleMarkets/dbn-go/internal/tui"
	"github.com/charmbracelet/huh"
	"github.com/neomantra/ymdflag"
	"github.com/segmentio/encoding/json"
	"github.com/spf13/cobra"
)

///////////////////////////////////////////////////////////////////////////////

var (
	databentoApiKey string

	dataset   string
	schemaStr string

	allSymbols  bool
	symbolsFile string

	stypeIn  dbn.SType = dbn.SType_RawSymbol
	stypeOut dbn.SType = dbn.SType_InstrumentId

	outputFile string
	emitJSON   bool // emit json from responses

	encoding    dbn.Encoding    = dbn.Encoding_Dbn
	compression dbn.Compression = dbn.Compress_ZStd

	jobID       string
	stateFilter string

	startYMD ymdflag.YMDFlag
	endYMD   ymdflag.YMDFlag

	useForce bool
)

func getDateRangeArg() dbn_hist.DateRange {
	dateRange := dbn_hist.DateRange{}
	if !startYMD.IsZero() {
		dateRange.Start = startYMD.AsTime()
	}
	if !endYMD.IsZero() {
		dateRange.End = endYMD.AsTime()
	}
	return dateRange
}

// getMetadataQueryParams returns a MetadataQueryParams struct based on CLI globals and the given symbols.
func getMetadataQueryParams(symbols []string) dbn_hist.MetadataQueryParams {
	return dbn_hist.MetadataQueryParams{
		Dataset:   dataset,
		Symbols:   symbols,
		Schema:    schemaStr,
		DateRange: getDateRangeArg(),
		Mode:      dbn_hist.FeedMode_Historical,
		StypeIn:   stypeIn,
		Limit:     -1,
	}
}

// SubmitJobParams returns a MetadataQueryParams struct based on CLI globals and the given symbols.
func getSubmitJobParams(symbols []string) dbn_hist.SubmitJobParams {
	symbolsStr := strings.Join(symbols, ",")
	schema, _ := dbn.SchemaFromString(schemaStr)
	return dbn_hist.SubmitJobParams{
		Dataset:     dataset,
		Symbols:     symbolsStr,
		Schema:      schema,
		DateRange:   getDateRangeArg(),
		Encoding:    encoding,
		Compression: compression,
		Packaging:   dbn_hist.Packaging_Tar,
		Delivery:    dbn_hist.Delivery_Download,
		StypeIn:     stypeIn,
		StypeOut:    stypeOut,
	}
}

// Returns a list of symbols from the command line arguments, or otherwise exits with an error.
// If --all is set, returns ["ALL_SYMBOLS"].  Also handles loading from a symbol file.
func requireSymbolArgs(args []string) []string {
	if allSymbols {
		return []string{"ALL_SYMBOLS"}
	}
	result := args

	if symbolsFile != "" {
		symbols, err := loadSymbolFile(symbolsFile)
		requireNoError(err)
		result = append(result, symbols...)
	}

	if len(result) == 0 {
		fmt.Fprint(os.Stderr, "must pass symbols as arguments or use --file or --all\n")
		os.Exit(1)
	}
	return result
}

// loadSymbolFile loads a newline delimited file of symbols from `filename` and returns them as a slice.
// Returns an error if any.
// Empty lines and rows starting with `#“ are ignored.
func loadSymbolFile(filename string) ([]string, error) {
	if filename == "" {
		return nil, fmt.Errorf("filename was empty")
	}

	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var symbols []string
	scanner := bufio.NewScanner(bytes.NewReader(fileBytes))
	for scanner.Scan() {
		text := scanner.Text()
		if len(text) > 0 && text[0] != '#' { // skip empty lines and comments
			symbols = append(symbols, strings.TrimSpace(text))
		}
	}
	if err := scanner.Err(); err != nil {
		return symbols, err
	}
	return symbols, nil
}

///////////////////////////////////////////////////////////////////////////////

func requireDatabentoApiKey() string {
	if databentoApiKey == "" {
		databentoApiKey = os.Getenv("DATABENTO_API_KEY")
		if databentoApiKey == "" {
			fmt.Fprint(os.Stderr, "DATABENTO_API_KEY not set.  Use --key or DATABENTO_API_KEY envvar.\n")
			os.Exit(1)
		}
	}
	return databentoApiKey
}

func requireNoError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}
}

func requireNoErrorMsg(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s\n", msg, err.Error())
		os.Exit(1)
	}
}

func requireHumanConfirmation(promptTitle string, verbName string) {
	doVerb := false
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Affirmative(fmt.Sprintf("Yes, %s", verbName)).
				Negative("No, Cancel").
				Title(promptTitle).
				Value(&doVerb),
		))
	err := form.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "confirmation error: %s\n", err.Error())
		os.Exit(1)
	}
	if !doVerb {
		os.Exit(0)
	}
}

///////////////////////////////////////////////////////////////////////////////

func main() {

	cobra.OnInitialize()

	rootCmd.PersistentFlags().StringVarP(&databentoApiKey, "key", "k", "", "DataBento API key (or use DATABENT_API_KEY envvar)")

	rootCmd.AddCommand(listDatasetsCmd)
	listDatasetsCmd.Flags().VarP(&startYMD, "start", "t", "Start date as YYYYMMDD")
	listDatasetsCmd.Flags().VarP(&endYMD, "end", "e", "End date as YYYYMMDD")
	listDatasetsCmd.Flags().BoolVarP(&emitJSON, "json", "j", false, "Emit JSON instead of simple summary")

	rootCmd.AddCommand(listPublishersCmd)
	listPublishersCmd.Flags().BoolVarP(&emitJSON, "json", "j", false, "Emit JSON instead of simple summary")

	rootCmd.AddCommand(listSchemasCmd)
	listSchemasCmd.Flags().StringVarP(&dataset, "dataset", "d", "", "Dataset to list schema for")
	listSchemasCmd.Flags().BoolVarP(&emitJSON, "json", "j", false, "Emit JSON instead of simple summary")
	listSchemasCmd.MarkFlagRequired("dataset")

	rootCmd.AddCommand(listFieldsCmd)
	listFieldsCmd.Flags().StringVarP(&schemaStr, "schema", "s", "", "Schema to list fields for")
	listFieldsCmd.Flags().VarP(&encoding, "encoding", "", "Encoding to use ('dbn', 'csv', 'json')")
	listFieldsCmd.Flags().BoolVarP(&emitJSON, "json", "j", false, "Emit JSON instead of simple summary")
	listFieldsCmd.MarkFlagRequired("schema")

	rootCmd.AddCommand(listUnitPricesCmd)
	listUnitPricesCmd.Flags().StringVarP(&dataset, "dataset", "d", "", "Dataset to list schema for")
	listUnitPricesCmd.Flags().BoolVarP(&emitJSON, "json", "j", false, "Emit JSON instead of simple summary")
	listUnitPricesCmd.MarkFlagRequired("dataset")

	rootCmd.AddCommand(getDatasetConditionCmd)
	getDatasetConditionCmd.Flags().StringVarP(&dataset, "dataset", "d", "", "Dataset to get condition for")
	getDatasetConditionCmd.Flags().VarP(&startYMD, "start", "t", "Start date as YYYYMMDD")
	getDatasetConditionCmd.Flags().VarP(&endYMD, "end", "e", "End date as YYYYMMDD")
	getDatasetConditionCmd.Flags().BoolVarP(&emitJSON, "json", "j", false, "Emit JSON instead of simple summary")
	getDatasetConditionCmd.MarkFlagRequired("dataset")

	rootCmd.AddCommand(getDatasetRangeCmd)
	getDatasetRangeCmd.Flags().StringVarP(&dataset, "dataset", "d", "", "Dataset to get date range for")
	getDatasetRangeCmd.MarkFlagRequired("dataset")

	rootCmd.AddCommand(getCostCmd)
	getCostCmd.Flags().StringVarP(&dataset, "dataset", "d", "", "Dataset to get cost for")
	getCostCmd.Flags().StringVarP(&schemaStr, "schema", "s", "", "Schema to get cost for")
	getCostCmd.Flags().StringVarP(&symbolsFile, "file", "f", "", "Newline delimited file to read symbols from (# is comment)")
	getCostCmd.Flags().BoolVarP(&allSymbols, "all", "", false, "Get record count for all symbols")
	getCostCmd.Flags().VarP(&startYMD, "start", "t", "Start date as YYYYMMDD")
	getCostCmd.Flags().VarP(&endYMD, "end", "e", "End date as YYYYMMDD")
	getCostCmd.Flags().BoolVarP(&emitJSON, "json", "j", false, "Emit JSON instead of simple summary")
	getCostCmd.Flags().VarP(&stypeIn, "sin", "", "Set stype_in: one of instrument_id, id, instr, raw_symbol, raw, smart, continuous, parent, nasdaq, cms (default: 'raw')")
	getCostCmd.Flags().VarP(&stypeOut, "sout", "", "Set stype_out: one of instrument_id, id, instr, raw_symbol, raw, smart, continuous, parent, nasdaq, cms (default: 'id')")
	getCostCmd.MarkFlagRequired("dataset")
	getCostCmd.MarkFlagRequired("schema")

	rootCmd.AddCommand(listJobsCmd)
	listJobsCmd.Flags().StringVarP(&stateFilter, "state", "", "", "Comma-separated Filter for job states. Can include 'received', 'queued', 'processing', 'done', and 'expired'.")
	listJobsCmd.Flags().BoolVarP(&emitJSON, "json", "j", false, "Emit JSON instead of simple summary")
	listJobsCmd.Flags().VarP(&startYMD, "start", "t", "Start date as YYYYMMDD (optional)")

	rootCmd.AddCommand(listFilesCmd)
	listFilesCmd.Flags().StringVarP(&jobID, "job", "", "", "Job ID to list files for")
	listFilesCmd.Flags().BoolVarP(&emitJSON, "json", "j", false, "Emit JSON instead of simple summary")
	listFilesCmd.MarkFlagRequired("job")

	rootCmd.AddCommand(submitJobCmd)
	submitJobCmd.Flags().StringVarP(&dataset, "dataset", "d", "", "Dataset to request")
	submitJobCmd.Flags().StringVarP(&schemaStr, "schema", "s", "", "Schema to request")
	submitJobCmd.Flags().VarP(&encoding, "encoding", "", "Encoding to use ('dbn', 'csv', 'json')")
	submitJobCmd.Flags().VarP(&compression, "compression", "", "Compression to use ('none', 'zstd')")
	submitJobCmd.Flags().StringVarP(&symbolsFile, "file", "f", "", "Newline delimited file to read symbols from (# is comment)")
	submitJobCmd.Flags().BoolVarP(&allSymbols, "all", "", false, "Request data for all symbols")
	submitJobCmd.Flags().BoolVarP(&useForce, "force", "", false, "Do not warn about all symbols or cost")
	submitJobCmd.Flags().VarP(&startYMD, "start", "t", "Start date as YYYYMMDD.")
	submitJobCmd.Flags().VarP(&endYMD, "end", "e", "End date as YYYYMMDD")
	submitJobCmd.Flags().VarP(&stypeIn, "sin", "", "Set stype_in: one of instrument_id, id, instr, raw_symbol, raw, smart, continuous, parent, nasdaq, cms (default: 'raw')")
	submitJobCmd.Flags().VarP(&stypeOut, "sout", "", "Set stype_out: one of instrument_id, id, instr, raw_symbol, raw, smart, continuous, parent, nasdaq, cms (default: 'id')")
	submitJobCmd.MarkFlagRequired("dataset")
	submitJobCmd.MarkFlagRequired("schema")
	submitJobCmd.MarkFlagRequired("start")

	rootCmd.AddCommand(getRangeCmd)
	getRangeCmd.Flags().StringVarP(&dataset, "dataset", "d", "", "Dataset to request")
	getRangeCmd.Flags().StringVarP(&schemaStr, "schema", "s", "", "Schema to request")
	getRangeCmd.Flags().VarP(&encoding, "encoding", "", "Encoding to use ('dbn', 'csv', 'json')")
	getRangeCmd.Flags().VarP(&compression, "compression", "", "Compression to use ('none', 'zstd')")
	getRangeCmd.Flags().StringVarP(&symbolsFile, "file", "f", "", "Newline delimited file to read symbols from (# is comment)")
	getRangeCmd.Flags().BoolVarP(&allSymbols, "all", "", false, "Request data for all symbols")
	getRangeCmd.Flags().BoolVarP(&useForce, "force", "", false, "Do not warn about all symbols or cost")
	getRangeCmd.Flags().VarP(&startYMD, "start", "t", "Start date as YYYYMMDD.")
	getRangeCmd.Flags().VarP(&endYMD, "end", "e", "End date as YYYYMMDD")
	getRangeCmd.Flags().VarP(&stypeIn, "sin", "", "Set stype_in: one of instrument_id, id, instr, raw_symbol, raw, smart, continuous, parent, nasdaq, cms (default: 'raw')")
	getRangeCmd.Flags().VarP(&stypeOut, "sout", "", "Set stype_out: one of instrument_id, id, instr, raw_symbol, raw, smart, continuous, parent, nasdaq, cms (default: 'id')")
	getRangeCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for data ('-' is stdout)")
	getRangeCmd.MarkFlagRequired("dataset")
	getRangeCmd.MarkFlagRequired("schema")
	getRangeCmd.MarkFlagRequired("start")
	getRangeCmd.MarkFlagRequired("output")

	rootCmd.AddCommand(resolveCmd)
	resolveCmd.Flags().StringVarP(&dataset, "dataset", "d", "", "Dataset to get cost for")
	resolveCmd.Flags().StringVarP(&schemaStr, "schema", "s", "", "Schema to get cost for")
	resolveCmd.Flags().BoolVarP(&allSymbols, "all", "", false, "Get record count for all symbols")
	resolveCmd.Flags().VarP(&startYMD, "start", "t", "Start date as YYYYMMDD")
	resolveCmd.Flags().VarP(&endYMD, "end", "e", "End date as YYYYMMDD")
	resolveCmd.Flags().VarP(&stypeIn, "sin", "", "Set stype_in: one of instrument_id, id, instr, raw_symbol, raw, smart, continuous, parent, nasdaq, cms (default: 'raw')")
	resolveCmd.Flags().VarP(&stypeOut, "sout", "", "Set stype_out: one of instrument_id, id, instr, raw_symbol, raw, smart, continuous, parent, nasdaq, cms (default: 'id')")
	resolveCmd.Flags().BoolVarP(&emitJSON, "json", "j", false, "Emit JSON instead of simple summary")
	resolveCmd.MarkFlagRequired("dataset")
	resolveCmd.MarkFlagRequired("start")

	rootCmd.AddCommand(tuiCmd)

	err := rootCmd.Execute()
	requireNoError(err)
}

///////////////////////////////////////////////////////////////////////////////

var rootCmd = &cobra.Command{
	Use:   "dbn-go-hist",
	Short: "dbn-go-hist queries the DataBento Historical API.",
	Long:  "dbn-go-hist queries the DataBento Historical API.",
}

var listDatasetsCmd = &cobra.Command{
	Use:     "datasets",
	Aliases: []string{"d"},
	Short:   "Queries DataBento Hist for datasets and prints them",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := requireDatabentoApiKey()
		dateRange := getDateRangeArg()

		datasets, err := dbn_hist.ListDatasets(apiKey, dateRange)
		requireNoError(err)

		if emitJSON {
			printJSON(datasets)
			return
		}
		for _, dataset := range datasets {
			fmt.Fprintf(os.Stdout, "%s\n", dataset)
		}
	},
}

var listPublishersCmd = &cobra.Command{
	Use:     "publishers",
	Aliases: []string{"p"},
	Short:   "Queries DataBento Hist for publishers and prints them",
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := requireDatabentoApiKey()
		publishers, err := dbn_hist.ListPublishers(apiKey)
		requireNoError(err)

		if emitJSON {
			printJSON(publishers)
			return
		}
		for _, publisher := range publishers {
			fmt.Fprintf(os.Stdout, "%d %s %s %s\n",
				publisher.PublisherID,
				publisher.Venue,
				publisher.Dataset,
				publisher.Description,
			)
		}
	},
}

var listSchemasCmd = &cobra.Command{
	Use:     "schemas",
	Aliases: []string{"s"},
	Short:   "Queries DataBento Hist for publishers and prints them",
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := requireDatabentoApiKey()
		schemas, err := dbn_hist.ListSchemas(apiKey, dataset)
		requireNoError(err)

		if emitJSON {
			printJSON(schemas)
			return
		}
		for _, schema := range schemas {
			fmt.Fprintf(os.Stdout, "  %s\n", schema)
		}
	},
}

var listFieldsCmd = &cobra.Command{
	Use:     "fields",
	Aliases: []string{"f"},
	Short:   "Queries DataBento Hist for fields of a schema/encoding and prints them",
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := requireDatabentoApiKey()
		schema, err := dbn.SchemaFromString(schemaStr)
		requireNoError(err)

		fields, err := dbn_hist.ListFields(apiKey, encoding, schema)
		requireNoError(err)

		if emitJSON {
			printJSON(fields)
			return
		}
		for _, field := range fields {
			fmt.Fprintf(os.Stdout, "%s %s\n", field.Name, field.TypeName)
		}
	},
}

var listUnitPricesCmd = &cobra.Command{
	Use:     "unit-prices",
	Aliases: []string{"u", "up"},
	Short:   "Queries DataBento Hist for unit prices of a dataset",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := requireDatabentoApiKey()
		unitPrices, err := dbn_hist.ListUnitPrices(apiKey, dataset)
		requireNoError(err)

		if emitJSON {
			printJSON(unitPrices)
			return
		}
		for _, unitPrice := range unitPrices {
			fmt.Fprintf(os.Stdout, "%s %s\n", dataset, unitPrice.Mode)
			for schema, price := range unitPrice.UnitPrices {
				fmt.Fprintf(os.Stdout, "    %s  %0.4f\n", schema, price)
			}
		}
	},
}

var getDatasetConditionCmd = &cobra.Command{
	Use:     "dataset-condition",
	Aliases: []string{"dc"},
	Short:   "Queries DataBento Hist for condition of a dataset",
	Args:    cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := requireDatabentoApiKey()
		conditions, err := dbn_hist.GetDatasetCondition(apiKey, dataset, getDateRangeArg())
		requireNoError(err)

		if emitJSON {
			printJSON(conditions)
			return
		}
		for _, c := range conditions {
			fmt.Fprintf(os.Stdout, "%s %s %s %s\n", dataset, c.Condition, c.Date, c.LastModified)
		}
	},
}

var getDatasetRangeCmd = &cobra.Command{
	Use:     "dataset-range",
	Aliases: []string{"dr"},
	Short:   "Queries DataBento Hist for date range of a dataset",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := requireDatabentoApiKey()
		datasetRange, err := dbn_hist.GetDatasetRange(apiKey, dataset)
		requireNoError(err)

		fmt.Fprintf(os.Stdout, "%s start:%s end: %s\n",
			dataset,
			datasetRange.Start.Format("2006-01-02"),
			datasetRange.End.Format("2006-01-02"),
		)
	},
}

var getCostCmd = &cobra.Command{
	Use:     "cost",
	Aliases: []string{"c"},
	Short:   "Queries DataBento Hist for the cost and size of a GetRange query",
	Args:    cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := requireDatabentoApiKey()
		symbols := requireSymbolArgs(args)
		metaParams := getMetadataQueryParams(symbols)

		cost, err := dbn_hist.GetCost(apiKey, metaParams)
		requireNoError(err)
		dataSize, err := dbn_hist.GetBillableSize(apiKey, metaParams)
		requireNoError(err)
		recordCount, err := dbn_hist.GetRecordCount(apiKey, metaParams)
		requireNoError(err)

		if emitJSON {
			printJSON(map[string]interface{}{
				"query":        metaParams,
				"cost":         cost,
				"data_size":    dataSize,
				"record_count": recordCount,
			})
			return
		}
		fmt.Fprintf(os.Stdout, "%s  %s   $ %0.06f  %d bytes  %d records\n",
			metaParams.Dataset, metaParams.Schema, cost, dataSize, recordCount)
	},
}

var listJobsCmd = &cobra.Command{
	Use:     "jobs",
	Aliases: []string{"lj", "j"},
	Short:   "Lists DataBento Hist jobs",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := requireDatabentoApiKey()

		jobs, err := dbn_hist.ListJobs(apiKey, stateFilter, getDateRangeArg().Start)
		requireNoError(err)

		if emitJSON {
			printJSON(jobs)
			return
		}
		for _, job := range jobs {
			jstr, err := json.Marshal(job)
			requireNoError(err)
			fmt.Fprintf(os.Stdout, "%s\n", jstr)
		}
	},
}

var listFilesCmd = &cobra.Command{
	Use:     "files",
	Aliases: []string{"lf", "f"},
	Short:   "Lists files for the given DataBento Hist JobID",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := requireDatabentoApiKey()

		files, err := dbn_hist.ListFiles(apiKey, jobID)
		requireNoError(err)

		if emitJSON {
			printJSON(files)
			return
		}
		for _, file := range files {
			jstr, err := json.Marshal(file)
			requireNoError(err)
			fmt.Fprintf(os.Stdout, "%s\n", jstr)
		}
	},
}

var submitJobCmd = &cobra.Command{
	Use:     "submit-job",
	Aliases: []string{"submit"},
	Short:   "Submit a data request job to the Hist API",
	Args:    cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := requireDatabentoApiKey()
		symbols := requireSymbolArgs(args)
		jobParams := getSubmitJobParams(symbols)

		requireBudgetApproval(apiKey, symbols, &jobParams)

		batchJob, err := dbn_hist.SubmitJob(apiKey, jobParams)
		requireNoErrorMsg(err, "error submitting job")

		jstr, err := json.Marshal(batchJob)
		requireNoErrorMsg(err, "error unmarshalling batchJob")

		fmt.Fprintf(os.Stdout, "%s\n", jstr)
	},
}

var getRangeCmd = &cobra.Command{
	Use:     "get-range",
	Aliases: []string{"range"},
	Short:   "Download a range of data from the Hist API",
	Args:    cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for file creation first
		var writer io.Writer
		var closer io.Closer
		fileCloser := func() {
			if closer != nil {
				closer.Close()
			}
		}
		if outputFile != "-" {
			file, err := os.Create(outputFile)
			requireNoErrorMsg(err, "error creating output file")
			writer, closer = file, file
		} else {
			writer, closer = os.Stdout, nil
		}
		defer fileCloser()

		// Now proceed with request
		apiKey := requireDatabentoApiKey()
		symbols := requireSymbolArgs(args)
		jobParams := getSubmitJobParams(symbols)

		requireBudgetApproval(apiKey, symbols, &jobParams)

		dbnData, err := dbn_hist.GetRange(apiKey, jobParams)
		requireNoErrorMsg(err, "error getting range")

		// Write the output
		_, err = writer.Write(dbnData)
		requireNoErrorMsg(err, "error writing output")
	},
}

var resolveCmd = &cobra.Command{
	Use:     "resolve",
	Aliases: []string{"symbols"},
	Short:   "Resolve symbols via the Databento Symbology API",
	Args:    cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := requireDatabentoApiKey()
		symbols := requireSymbolArgs(args)
		dateRange := getDateRangeArg()

		resolveParams := dbn_hist.ResolveParams{
			Dataset:   dataset,
			Symbols:   symbols,
			StypeIn:   stypeIn,
			StypeOut:  stypeOut,
			DateRange: dateRange,
		}

		resolution, err := dbn_hist.SymbologyResolve(apiKey, resolveParams)
		requireNoError(err)

		if emitJSON {
			printJSON(resolution)
			return
		}

		// human mode just print the symbols
		for symbol := range resolution.Mappings {
			fmt.Fprintf(os.Stdout, "%s\n", symbol)
		}
	},
}

var tuiCmd = &cobra.Command{
	Use:     "tui",
	Aliases: []string{"symbols"},
	Short:   "dbn-go-hist TUI",
	Args:    cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		config := dbn_tui.Config{
			DatabentoApiKey: requireDatabentoApiKey(),
		}
		err := dbn_tui.Run(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		}
	},
}

//////////////////////////////////////////////////////////////////////////////

func requireBudgetApproval(apiKey string, symbols []string, params *dbn_hist.SubmitJobParams) {
	if useForce {
		return
	}

	if params.Symbols == "" || strings.Contains(params.Symbols, "ALL_SYMBOLS") {
		fmt.Fprint(os.Stderr, "Submitting for ALL_SYMBOLS...\n")
		requireHumanConfirmation(
			"Are you sure you want to submit for ALL_SYMBOLS?",
			"Submit for ALL")
	}

	// Request cost of this job
	fmt.Fprintf(os.Stderr, "Getting cost estimates for job...\n")
	metaParams := getMetadataQueryParams(symbols)
	cost, err := dbn_hist.GetCost(apiKey, metaParams)
	requireNoError(err)
	dataSize, err := dbn_hist.GetBillableSize(apiKey, metaParams)
	requireNoError(err)
	recordCount, err := dbn_hist.GetRecordCount(apiKey, metaParams)
	requireNoError(err)

	fmt.Fprintf(os.Stderr, "Estimated cost of $%.2f for %d records and %d bytes of data.\n",
		cost, recordCount, dataSize)
	requireHumanConfirmation("Are you sure you want to submit?\n", "Submit Job")
}

// printJSON is a generic helper to print a value as JSON to stdout.
func printJSON[T any](val T) {
	jstr, err := json.Marshal(val)
	requireNoError(err)
	fmt.Fprintf(os.Stdout, "%s\n", jstr)
}
