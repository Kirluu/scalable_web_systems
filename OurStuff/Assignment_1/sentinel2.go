package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"io"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudkms/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/urlfetch"
	"strconv"
)

func adc() {
	ctx := context.Background()

	// [START auth_cloud_implicit]

	// For API packages whose import path is starting with "cloud.google.com/go",
	// such as cloud.google.com/go/storage in this case, if there are no credentials
	// provided, the client library will look for credentials in the environment.
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	it := storageClient.Buckets(ctx, "project-id")
	for {
		bucketAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(bucketAttrs.Name)
	}

	// For packages whose import path is starting with "google.golang.org/api",
	// such as google.golang.org/api/cloudkms/v1, use the
	// golang.org/x/oauth2/google package as shown below.
	oauthClient, err := google.DefaultClient(ctx, cloudkms.CloudPlatformScope)
	if err != nil {
		log.Fatal(err)
	}

	kmsService, err := cloudkms.New(oauthClient)
	if err != nil {
		log.Fatal(err)
	}

	// [END auth_cloud_implicit]
	_ = kmsService
}

func query(lat float64, long float64, w http.ResponseWriter, r *http.Request) (*bigquery.RowIterator, error) {
	//ctx := context.Background()
	ctx := appengine.NewContext(r)

	client, err := bigquery.NewClient(ctx, "kulr-178408")
	//client, err := bigquery.NewClient(ctx, "johaa-178408")
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

	fmt.Fprintf(w, "Your query: \n" + queryString)

	query := client.Query(queryString)

	// Use standard SQL syntax for queries.
	// See: https://cloud.google.com/bigquery/sql-reference/
	query.QueryConfig.UseStandardSQL = true
	return query.Read(ctx)
}

func printBaseUrls(w io.Writer, iter *bigquery.RowIterator) error {
	for {
		var baseUrl string
		err := iter.Next(&baseUrl)
		if err == iterator.Done {
			return nil
		}
		if err != nil {
			return err
		}

		fmt.Fprintf(w, baseUrl)
	}
}

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

func init() {
	http.HandleFunc("/", handler) // Overall default handler
	// http.HandleFunc("/add", addItem)
	// http.HandleFunc("/removeAll", removeAllItems)
	// http.HandleFunc("/removeName", removeItemByName)
	// http.HandleFunc("/getmarket", getItemsInSupermarket)
	// http.HandleFunc("/totalprice", getItemsTotalPrice)

	// http.HandleFunc("/complete", completeHandler)
	// http.HandleFunc("/incomplete", incompleteHandler)
	// http.HandleFunc("/goget", getHandler)
	http.HandleFunc("/test", testHandler)
	//http.HandleFunc("/images", getImages)
	http.HandleFunc("/bigquery", getBigquery)
	http.HandleFunc("/test2", testquery)

}


func main() {
	//adc()
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

func getBigquery(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	if (latStr == "") { fmt.Fprintf(w, "Missing query input: latitude (lat)")
		return }
	lat, latErr := strconv.ParseFloat(latStr, 64)
	if (latErr != nil) { fmt.Fprintf(w, "Failed to parse latitude!")
		return }

	longStr := r.URL.Query().Get("long")
	if (longStr == "") { fmt.Fprintf(w, "Missing query input: longitude (long)")
		return }
	long, longErr := strconv.ParseFloat(longStr, 64)
	if (longErr != nil) { fmt.Fprintf(w, "Failed to parse longitude!")
		return }

	// Perform query, given the now successfully parsed parameters
	iter, err := query(lat, long, w, r)

	if err == nil && iter != nil {
		fmt.Fprintf(w, "Now going to print results!")
		printBaseUrls(w, iter)
		return
	}
	fmt.Fprintf(w, "BigQuery contact failed %s", err)
}

func getImages(w http.ResponseWriter, r *http.Request) {

	lat := r.URL.Query().Get("lat")
	long := r.URL.Query().Get("long")

	newReq := getUnknownURL(lat, long)

	req, err := http.NewRequest("GET", newReq, nil)

	if err != nil {
		http.Error(w, "Failed ", http.StatusNotFound)
		return
	}

	dec := json.NewDecoder(req.Body)
	errDec := dec.Decode(&req)

	if errDec != nil {
		http.Error(w, "Failed json decode", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "lat long %s : %v ", lat, long)
}

func getUnknownURL(lat string, long string) string {

	return "some url"
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "MY MESSAGE - PRINT IT NOOOW!")
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

func completeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Some text!")

	// create a new App Engine context from the HTTP request.
	ctx := appengine.NewContext(r)

	p := &Item{Name: "gopher", SuperMarket: "Netto"}

	// create a new complete key of kind Person and value gopher.
	key := datastore.NewKey(ctx, "Person", "gopher", 0, nil)
	// put p in the datastore.
	key, err := datastore.Put(ctx, key, p)
	if err != nil {
		fmt.Fprintf(w, "Error occurred!")
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "gopher stored with key %v", key)
}

func incompleteHandler(w http.ResponseWriter, r *http.Request) {
	// create a new App Engine context from the HTTP request.
	ctx := appengine.NewContext(r)

	p := &Item{Name: "gopher", SuperMarket: "Netto"}

	// create a new complete key of kind Person.
	key := datastore.NewIncompleteKey(ctx, "Person", nil)
	// put p in the datastore.
	key, err := datastore.Put(ctx, key, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Fprintf(w, "Error!")
		return
	}
	fmt.Fprintf(w, "gopher stored with key %v", key)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	key := datastore.NewKey(ctx, "Person", "gopher", 0, nil)

	var p Item
	err := datastore.Get(ctx, key, &p)
	if err != nil {
		http.Error(w, "Person not found", http.StatusNotFound)
		return
	}
	fmt.Fprintln(w, p)
}

var items = []Item{}

type Item struct {
	Name        string `json:"name"`
	Price       int    `json:"price"`
	SuperMarket string `json:"superMarket"`
}

func addItem(w http.ResponseWriter, r *http.Request) {
	var i Item

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&i)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	items = append(items, i)
	fmt.Fprintf(w, "Added %v", i.Name)
}

func remove(s []Item, i int) []Item {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func removeItemByName(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "could not read body: %v", err)
		return
	}
	name := string(b)

	for index, e := range items {
		if e.Name == name {
			items = remove(items, index)
		}
	}

}

func removeAllItems(w http.ResponseWriter, r *http.Request) {
	items = []Item{}
}

func getItemsTotalPrice(w http.ResponseWriter, r *http.Request) {
	sum := 0
	for _, e := range items {
		sum += e.Price
	}

	fmt.Fprintf(w, "total price: %v", sum)
}

func getItemsInSupermarket(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "could not read body: %v", err)
		return
	}
	smarket := string(b)

	sum := []Item{}
	for _, e := range items {
		if e.SuperMarket == smarket {
			sum = append(sum, e)
		}
	}

	fmt.Fprintf(w, "Items in %s : %v ", smarket, sum)
}

func testquery(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	//client, err := bigquery.NewClient(ctx, "johaa-178408")
	client, err := bigquery.NewClient(ctx, "kulr-178408")
	if err != nil {
	}
	q := client.Query("SELECT base_url FROM `bigquery-public-data.cloud_storage_geo_index.sentinel_2_index` where west_lon < 60 and west_lon > 59.5 and south_lat > 80.9 and south_lat < 81 LIMIT 1000")
	q.QueryConfig.UseStandardSQL = true
	//q.DefaultProjectID = "bigquery-public-data"
	//q.DefaultDatasetID = "cloud_storage_geo_index"

	it, err2 := q.Read(ctx)

	if err2 != nil {
		log.Fatal(err2)
	}

	printResults(w, it)

}

func httpGet(url string) string {
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	str := buf.String()
	println(str)
	return str
}
