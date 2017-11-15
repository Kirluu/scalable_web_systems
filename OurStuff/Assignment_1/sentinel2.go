package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"io"

	"strconv"
	"strings"

	"cloud.google.com/go/bigquery"

	"google.golang.org/api/iterator"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

func init() {
	http.HandleFunc("/", handler) // Overall default handler
	http.HandleFunc("/bigquery", getBigquery)

}

func main() {
	appengine.Main()

	// Connection string: "staging.johaa-178408.appspot.com"
	//resp, err := http.Get("staging.johaa-178408.appspot.com")

	/*err := http.ListenAndServe("localhost:5080", nil)
	if err != nil {
		log.Fatal(err)
	}*/
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
	baseUrlList, err := getBaseUrls(lat, long, w, r)
	if err != nil || baseUrlList == nil {
		fmt.Fprintf(w, "BigQuery contact failed %s", err)
		return
	}
	if len(baseUrlList) == 0 {
		fmt.Fprintf(w, "No base-URLs could be found for the given Long and Lat arguments.")
		return
	}

	// Now use first (arbitrary) base-URL to do a nice little request
	var baseUrl = baseUrlList[0]
	handleBaseUrl(w, r, baseUrl)

	//fmt.Fprintf(w, "\nReached the end of the handler!")
}

func getBaseUrls(lat float64, long float64, w http.ResponseWriter, r *http.Request) ([]string, error) {
	//fmt.Fprintf(w, "getBaseUrls\n")
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

	//fmt.Fprintf(w, "Your query: \n"+queryString+"\n\n")

	query := client.Query(queryString)

	// Use standard SQL syntax for queries.
	// See: https://cloud.google.com/bigquery/sql-reference/
	query.QueryConfig.UseStandardSQL = true

	var queryIterator, readErr = query.Read(ctx)
	if readErr != nil {
		return nil, readErr
	}

	//fmt.Fprintf(w, "\n\n")

	return printBaseUrls(w, queryIterator)
}

func printBaseUrls(w io.Writer, iter *bigquery.RowIterator) ([]string, error) {
	//fmt.Fprintf(w, "printBaseUrls\n")
	var resList []string

	for {
		var baseUrl []bigquery.Value
		err := iter.Next(&baseUrl)
		if err == iterator.Done {
			return resList, nil
		}
		if err != nil {
			return nil, err
		}
		resList = append(resList, fmt.Sprintf("%s", baseUrl[0]))
		//fmt.Fprintf(w, "\n%s", baseUrl[0])
	}
	//fmt.Fprintf(w, "\n\n")

	return resList, nil
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

type imageResults struct {
	links []string `json:"links"`
}

func handleBaseUrl(w http.ResponseWriter, r *http.Request, baseUrl string) error {
	//fmt.Fprintf(w, "apiPrefixRequest\n")
	// remove the useless front-part of the given base-URL + add known GRANULE-directory
	var baseUrlCorrect = strings.TrimPrefix(baseUrl, "gs://gcp-public-data-sentinel-2/") + "/GRANULE/"
	var requestUrl string = "https://www.googleapis.com/storage/v1/b/gcp-public-data-sentinel-2/o"

	var fullUrl string = requestUrl + "?delimiter=/&prefix=" + baseUrlCorrect

	//httpGet(w, r, fullUrl) // Prints the content string

	var apiRes, apiErr = httpGetApiResult(w, r, fullUrl)
	if apiErr != nil {
		log.Fatal("Failed to get API result.")
		return apiErr
	}

	var imgURL = fullUrl + strings.TrimPrefix(apiRes.Prefixes[0], baseUrlCorrect) + "IMG_DATA/"

	var apiImages, apiImgErr = httpGetApiResult(w, r, imgURL)
	if apiImgErr != nil {
		log.Fatal("Failed to get Images.")
		return apiImgErr
	}
	list := []string{}

	if len(apiImages.Items) > 0 {
		// Iterate returned prefixes
		for i := 0; i < len(apiImages.Items); i++ {
			item := apiImages.Items[i]
			//fmt.Fprintf(w, "\n%s\n", item.SelfLink)
			list = append(list, item.SelfLink)
		}
		enc := json.NewEncoder(w)
		enc.Encode(list)

	} else {
		fmt.Fprintf(w, "No items discovered under IMG_DATA.")
	}
	//fmt.Fprintf(w, "\n\n")

	return nil //no error
}

func httpGetApiResult(w http.ResponseWriter, r *http.Request, url string) (ApiResult, error) {
	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)
	resp, err := client.Get(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return ApiResult{}, err
	}

	var res ApiResult
	dec := json.NewDecoder(resp.Body)
	decErr := dec.Decode(&res)
	if decErr != nil {
		return ApiResult{}, decErr
	}

	//fmt.Fprintf(w, "ApiResult given Base-URL request:\n%s\n", res)

	return res, nil
}

func httpGet(w http.ResponseWriter, r *http.Request, url string) (string, error) {
	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)
	resp, err := client.Get(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", err
	}
	buf := new(bytes.Buffer)
	var _, readErr = buf.ReadFrom(resp.Body)
	if readErr != nil {
		log.Fatal(err)
		return "", readErr
	}
	str := buf.String()
	//fmt.Fprintf(w, "'httpGet' RESPONSE:\n\n%s", str)

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
