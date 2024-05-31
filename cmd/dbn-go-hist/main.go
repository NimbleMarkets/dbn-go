// Copyright (c) 2024 Neomantra Corp
//
// NOTE: this may incur billing, handle with care!
//

package main

import (
	"fmt"
	"os"

	"github.com/NimbleMarkets/dbn-go"
	dbn_hist "github.com/NimbleMarkets/dbn-go/hist"
	"github.com/neomantra/ymdflag"
	"github.com/spf13/cobra"
)

///////////////////////////////////////////////////////////////////////////////

var (
	databentoApiKey string

	dataset    string
	schemaStr  string
	allSymbols bool
	startYMD   ymdflag.YMDFlag
	endYMD     ymdflag.YMDFlag
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

func getMetadataQueryParams(symbols []string) dbn_hist.MetadataQueryParams {
	return dbn_hist.MetadataQueryParams{
		Dataset:   dataset,
		Symbols:   symbols,
		Schema:    schemaStr,
		DateRange: getDateRangeArg(),
		Mode:      dbn_hist.FeedMode_Historical,
		StypeIn:   dbn.SType_RawSymbol,
		Limit:     -1,
	}
}

func requireSymbolArgs(args []string) []string {
	if allSymbols {
		return []string{"ALL_SYMBOLS"}
	}
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, "must pass symbols as arguments or use --all\n")
		os.Exit(1)
	}
	return args
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
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

///////////////////////////////////////////////////////////////////////////////

func main() {

	cobra.OnInitialize()

	rootCmd.PersistentFlags().StringVarP(&databentoApiKey, "key", "k", "", "DataBento API key (or use DATABENT_API_KEY envvar)")

	rootCmd.AddCommand(listDatasetsCmd)
	listDatasetsCmd.Flags().VarP(&startYMD, "start", "t", "Start date as YYYYMMDD")
	listDatasetsCmd.Flags().VarP(&endYMD, "end", "e", "End date as YYYYMMDD")

	rootCmd.AddCommand(listPublishersCmd)

	rootCmd.AddCommand(listSchemasCmd)
	listSchemasCmd.Flags().StringVarP(&dataset, "dataset", "d", "", "Dataset to list schema for")
	listSchemasCmd.MarkFlagRequired("dataset")

	rootCmd.AddCommand(listFieldsCmd)
	listFieldsCmd.Flags().StringVarP(&schemaStr, "schema", "s", "", "Schema to list fields for")
	listFieldsCmd.MarkFlagRequired("schema")

	rootCmd.AddCommand(listUnitPricesCmd)
	listUnitPricesCmd.Flags().StringVarP(&dataset, "dataset", "d", "", "Dataset to list schema for")
	listUnitPricesCmd.MarkFlagRequired("dataset")

	rootCmd.AddCommand(getDatasetConditionCmd)
	getDatasetConditionCmd.Flags().StringVarP(&dataset, "dataset", "d", "", "Dataset to get condition for")
	getDatasetConditionCmd.Flags().VarP(&startYMD, "start", "t", "Start date as YYYYMMDD")
	getDatasetConditionCmd.Flags().VarP(&endYMD, "end", "e", "End date as YYYYMMDD")
	getDatasetConditionCmd.MarkFlagRequired("dataset")

	rootCmd.AddCommand(getDatasetRangeCmd)
	getDatasetRangeCmd.Flags().StringVarP(&dataset, "dataset", "d", "", "Dataset to get date range for")
	getDatasetRangeCmd.MarkFlagRequired("dataset")

	rootCmd.AddCommand(getRecordCountCmd)
	getRecordCountCmd.Flags().StringVarP(&dataset, "dataset", "d", "", "Dataset to get condition for")
	getRecordCountCmd.Flags().StringVarP(&schemaStr, "schema", "s", "", "Schema to get condition for")
	getRecordCountCmd.Flags().BoolVarP(&allSymbols, "all", "", false, "Get record count for all symbols")
	getRecordCountCmd.Flags().VarP(&startYMD, "start", "t", "Start date as YYYYMMDD.")
	getRecordCountCmd.Flags().VarP(&endYMD, "end", "e", "End date as YYYYMMDD")
	getRecordCountCmd.MarkFlagRequired("dataset")
	getRecordCountCmd.MarkFlagRequired("schema")

	rootCmd.AddCommand(getBillableSizeCmd)
	getBillableSizeCmd.Flags().StringVarP(&dataset, "dataset", "d", "", "Dataset to get billable size for")
	getBillableSizeCmd.Flags().StringVarP(&schemaStr, "schema", "s", "", "Schema to get billable size for")
	getBillableSizeCmd.Flags().BoolVarP(&allSymbols, "all", "", false, "Get record count for all symbols")
	getBillableSizeCmd.Flags().VarP(&startYMD, "start", "t", "Start date as YYYYMMDD")
	getBillableSizeCmd.Flags().VarP(&endYMD, "end", "e", "End date as YYYYMMDD")
	getBillableSizeCmd.MarkFlagRequired("dataset")
	getBillableSizeCmd.MarkFlagRequired("schema")

	rootCmd.AddCommand(getCostCmd)
	getCostCmd.Flags().StringVarP(&dataset, "dataset", "d", "", "Dataset to get cost for")
	getCostCmd.Flags().StringVarP(&schemaStr, "schema", "s", "", "Schema to get cost for")
	getCostCmd.Flags().BoolVarP(&allSymbols, "all", "", false, "Get record count for all symbols")
	getCostCmd.Flags().VarP(&startYMD, "start", "t", "Start date as YYYYMMDD")
	getCostCmd.Flags().VarP(&endYMD, "end", "e", "End date as YYYYMMDD")
	getCostCmd.MarkFlagRequired("dataset")
	getCostCmd.MarkFlagRequired("schema")

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

		dateRange := dbn_hist.DateRange{}
		if !startYMD.IsZero() {
			dateRange.Start = startYMD.AsTime()
		}
		if !endYMD.IsZero() {
			dateRange.End = endYMD.AsTime()
		}

		datasets, err := dbn_hist.ListDatasets(apiKey, dateRange)
		requireNoError(err)

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

		fields, err := dbn_hist.ListFields(apiKey, dbn.Encoding_Dbn, schema)
		requireNoError(err)

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

var getRecordCountCmd = &cobra.Command{
	Use:     "record-count",
	Aliases: []string{"rc"},
	Short:   "Queries DataBento Hist for record count for a GetRange query.  Args are in symbols",
	Args:    cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := requireDatabentoApiKey()
		symbols := requireSymbolArgs(args)
		params := getMetadataQueryParams(symbols)

		recordCount, err := dbn_hist.GetRecordCount(apiKey, params)
		requireNoError(err)

		fmt.Fprintf(os.Stdout, "%d\n", recordCount)
	},
}

var getBillableSizeCmd = &cobra.Command{
	Use:     "billable-size",
	Aliases: []string{"bs"},
	Short:   "Queries DataBento Hist for the billable size of a GetRange query",
	Args:    cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := requireDatabentoApiKey()
		symbols := requireSymbolArgs(args)
		params := getMetadataQueryParams(symbols)

		billable, err := dbn_hist.GetBillableSize(apiKey, params)
		requireNoError(err)

		fmt.Fprintf(os.Stdout, "%d\n", billable)
	},
}

var getCostCmd = &cobra.Command{
	Use:     "cost",
	Aliases: []string{"c"},
	Short:   "Queries DataBento Hist for the cost for a GetRange query",
	Args:    cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := requireDatabentoApiKey()
		symbols := requireSymbolArgs(args)
		params := getMetadataQueryParams(symbols)

		cost, err := dbn_hist.GetCost(apiKey, params)
		requireNoError(err)

		fmt.Fprintf(os.Stdout, "%0.06f\n", cost)
	},
}
