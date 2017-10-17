package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"io"

	"cloud.google.com/go/bigquery"
	/*"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudkms/v1"*/
	"strconv"
	"strings"

	"google.golang.org/api/iterator"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

func getBaseUrls(lat float64, long float64, w http.ResponseWriter, r *http.Request) ([]string, error) {
	//ctx := context.Background()
	ctx := appengine.NewContext(r)

	//client, err := bigquery.NewClient(ctx, "kulr-178408")
	client, err := bigquery.NewClient(ctx, "johaa-178408")
	if err != nil {
		fmt.Fprintf(w, "error when creating BigQuery client from appengine context!")
		return nil, err
	}

	latLess := lat - 0.5
	latMore := lat + 0.5
	longLess := long - 0.5
	longMore := long + 0.5

	// Without params:
	//queryString := "SELECT base_url FROM `bigquery-public-data.cloud_storage_geo_index.sentinel_2_index` where west_lon < 60 and west_lon > 59.5 and south_lat > 80.9 and south_lat < 81 LIMIT 1000"
	// with params:
	queryString := fmt.Sprintf("SELECT base_url FROM `bigquery-public-data.cloud_storage_geo_index.sentinel_2_index` where west_lon < %g and west_lon > %g and south_lat < %g and south_lat > %g LIMIT 1000", longMore, longLess, latMore, latLess)

	fmt.Fprintf(w, "Your query: \n"+queryString)

	query := client.Query(queryString)

	// Use standard SQL syntax for queries.
	// See: https://cloud.google.com/bigquery/sql-reference/
	query.QueryConfig.UseStandardSQL = true

	var queryIterator, readErr = query.Read(ctx)
	if readErr != nil {
		return nil, readErr
	}

	return printBaseUrls(w, queryIterator)
}

func printBaseUrls(w io.Writer, iter *bigquery.RowIterator) ([]string, error) {
	var resList []string

	for {
		var baseUrl string
		err := iter.Next(&baseUrl)
		if err == iterator.Done {
			return resList, err
		}
		if err != nil {
			return nil, err
		}
		resList = append(resList, baseUrl)
		fmt.Fprintf(w, baseUrl)
	}

	return resList, nil
}

func init() {
	http.HandleFunc("/", handler) // Overall default handler
	http.HandleFunc("/apitest", getApiTestQuery)
	http.HandleFunc("/bigquery", getBigquery)
	http.HandleFunc("/test2", testquery)

}

func main() {
	appengine.Main()

	// Connection string: "staging.johaa-178408.appspot.com"
	//resp, err := http.Get("staging.johaa-178408.appspot.com")
	/*
		err := http.ListenAndServe("localhost:5080", nil)
		if err != nil {
			log.Fatal(err)
		}
	*/
}

type ApiResult struct {
	Prefixes []string  `json:"prefixes"`
	Items    []ApiItem `json:"items"`
}

type ApiItem struct {
	Id        string `json:"id"`
	SelfLink  string `json:"selfLink"`
	Name      string `json:"name"`
	MediaLink string `json:"mediaLink"`
}

func getBigquery(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	if latStr == "" {
		fmt.Fprintf(w, "Missing query input: latitude (lat)")
		return
	}
	lat, latErr := strconv.ParseFloat(latStr, 64)
	if latErr != nil {
		fmt.Fprintf(w, "Failed to parse latitude!")
		return
	}

	longStr := r.URL.Query().Get("long")
	if longStr == "" {
		fmt.Fprintf(w, "Missing query input: longitude (long)")
		return
	}
	long, longErr := strconv.ParseFloat(longStr, 64)
	if longErr != nil {
		fmt.Fprintf(w, "Failed to parse longitude!")
		return
	}

	// Perform query, given the now successfully parsed parameters
	baseUrlIter, err := getBaseUrls(lat, long, w, r)
	if err != nil || baseUrlIter == nil {
		fmt.Fprintf(w, "BigQuery contact failed %s", err)
		return
	}
	if len(baseUrlIter) == 0 {
		fmt.Fprintf(w, "No base-URLs retrieved.")
		return
	}

	// Now use first (arbitrary) base-URL to do a nice little request
	var baseUrl = baseUrlIter[0]
	handleBaseUrl(w, baseUrl)

	fmt.Fprintf(w, "\nReached the end of the handler!")
}

func getApiTestQuery(w http.ResponseWriter, r *http.Request) {
	// Test the api access giving a specific base-URL
	handleBaseUrl(w, "gs://gcp-public-data-sentinel-2/tiles/41/X/MK/S2A_MSIL1C_20170810T110621_N0205_R137_T41XMK_20170810T110621.SAFE")
}

func handleBaseUrl(w http.ResponseWriter, baseUrl string) error {

	var prefixes, apiPrefErr = apiPrefixesRequest(baseUrl)
	if apiPrefErr != nil {
		fmt.Fprintf(w, "Failed when getting prefixes: %s", apiPrefErr)
	}
	fmt.Fprintf(w, "Succeeded in extracting prefixes:")
	if len(prefixes) > 0 {
		fmt.Fprintf(w, prefixes[0]) // TODO: Loop over prefixes and print each, or something like that
	}

	// Full success, return no error:
	return nil
}

func apiPrefixesRequest(baseUrl string) ([]string, error) {
	// remove the useless front-part of the given base-URL + add known GRANULE-directory
	var baseUrlCorrect = strings.TrimPrefix(baseUrl, "gs://gcp-public-data-sentinel-2/") + "/GRANULE/"
	var requestUrl string = "https://www.googleapis.com/storage/v1/b/gcp-public-data-sentinel-2/o"

	var fullUrl string = requestUrl + "?delimiter=/&prefix=" + baseUrlCorrect

	httpGet(fullUrl) // Prints the content string

	var apiRes, apiErr = httpGetApiResult(fullUrl)
	if apiErr != nil {
		log.Fatal("Failed to get API result.")
		return []string{}, apiErr
	}

	return apiRes.Prefixes, nil
}

func httpGetApiResult(url string) (ApiResult, error) {
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		return ApiResult{}, nil
	}

	var res ApiResult
	dec := json.NewDecoder(response.Body)
	decErr := dec.Decode(&res)
	if decErr != nil {
		return ApiResult{}, decErr
	}
	return res, nil
}

func testquery(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	//client, err := bigquery.NewClient(ctx, "johaa-178408")
	client, err := bigquery.NewClient(ctx, "kulr-178408")
	if err != nil {
	}
	//q := client.Query("SELECT * FROM `bigquery-public-data.cloud_storage_geo_index.sentinel_2_index` where west_lon < 60 and west_lon > 59.5 and south_lat > 80.9 and south_lat < 81 LIMIT 1000")
	q := client.Query("SELECT base_url FROM `bigquery-public-data.cloud_storage_geo_index.sentinel_2_index` where west_lon < 60 and west_lon > 59.5 and south_lat > 80.9 and south_lat < 81 LIMIT 1000")
	q.QueryConfig.UseStandardSQL = true
	//q.DefaultProjectID = "bigquery-public-data"
	//q.DefaultDatasetID = "cloud_storage_geo_index"

	it, err2 := q.Read(ctx)

	if err2 != nil {
		log.Fatal(err2)
	}

	//fmt.Fprintln(w, it)
	//printResults(w, it)

	printBaseUrls(w, it)
}

func httpGet(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	buf := new(bytes.Buffer)
	var _, readErr = buf.ReadFrom(response.Body)
	if readErr != nil {
		return "", readErr
	}
	str := buf.String()
	println(str)
	return str, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	// first create a new context
	c := appengine.NewContext(r)
	// and use that context to create a new http client
	client := urlfetch.Client(c)

	// now we can use that http client as before
	res, err := client.Get("http://google.com")
	if err != nil {
		http.Error(w, fmt.Sprintf("could not get google: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Got Google with status %s\n", res.Status)
}

func get(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)
	resp, err := client.Get("https://www.google.com/")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "HTTP GET returned status %v", resp.Status)
}
