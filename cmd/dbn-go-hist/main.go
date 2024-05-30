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
	startYMD        ymdflag.YMDFlag
	endYMD          ymdflag.YMDFlag
)

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
	listDatasetsCmd.Flags().VarP(&startYMD, "start", "", "Start date as YYYYMMDD")
	listDatasetsCmd.Flags().VarP(&endYMD, "end", "", "End date as YYYYMMDD")

	rootCmd.AddCommand(listPublishersCmd)
	rootCmd.AddCommand(listSchemasCmd)
	rootCmd.AddCommand(listFieldsCmd)
	rootCmd.AddCommand(listUnitPricesCmd)

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
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := requireDatabentoApiKey()
		dataset := args[0]
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
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := requireDatabentoApiKey()
		schemaStr := args[0]
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
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := requireDatabentoApiKey()
		dataset := args[0]
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
