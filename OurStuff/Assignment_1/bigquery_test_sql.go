// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Command simpleapp queries the Shakespeare sample dataset in Google BigQuery.


package main
/*
package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"golang.org/x/net/context"
)

func main() {
	proj := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if proj == "" {
		fmt.Println("GOOGLE_CLOUD_PROJECT environment variable must be set.")
		os.Exit(1)
	}

	rows, err := query("bigquery-public-data")
	if err != nil {
		log.Fatal(err)
	}
	if err := printResults(os.Stdout, rows); err != nil {
		log.Fatal(err)
	}
}

// query returns a slice of the results of a query.
func query(proj string) (*bigquery.RowIterator, error) {
	ctx := context.Background()

	client, err := bigquery.NewClient(ctx, proj)
	if err != nil {
		return nil, err
	}
    // old prefix : bigquery-public-data: - this is the project iD ?!?!?!?
	query := client.Query(
		`SELECT BASE_URL
		 FROM cloud_storage_geo_index.sentinel_2_index
		 WHERE west_lon < 60 and west_lon > 59 and south_lat > 80 and south_lat < 81
		 LIMIT 1000;`)
	query := client.Query("SELECT base_url FROM cloud_storage_geo_index.sentinel_2_index where west_lon < 60 and south_lat > 80 LIMIT 1000")


	// Use standard SQL syntax for queries.
	// See: https://cloud.google.com/bigquery/sql-reference/
	query.QueryConfig.UseStandardSQL = true
	return query.Read(ctx)
}

// printResults prints results from a query to the Shakespeare dataset.
func printResults(w io.Writer, iter *bigquery.RowIterator) error {
	for {
		var row []bigquery.Value
		err := iter.Next(&row)
		if err == iterator.Done {
			return nil
		}
		if err != nil {
			return err
		}

		fmt.Fprintln(w, "titles:")
		ts := row[0].([]bigquery.Value)
		for _, t := range ts {
			record := t.([]bigquery.Value)
			title := record[0].(string)
			cnt := record[1].(int64)
			fmt.Fprintf(w, "\t%s: %d\n", title, cnt)
		}

		words := row[1].(int64)
		fmt.Fprintf(w, "total unique words: %d\n", words)
	}
}
*/