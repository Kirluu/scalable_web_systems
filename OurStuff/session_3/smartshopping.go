package main

import (
	"net/http"
	"log"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/urlfetch"
)

func main() {
	appengine.Main()




	// Connection string: "staging.johaa-178408.appspot.com"
	//resp, err := http.Get("staging.johaa-178408.appspot.com")



	err := http.ListenAndServe("localhost:5080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	http.HandleFunc("/", handler) // Overall default handler
	http.HandleFunc("/add", addItem)
	http.HandleFunc("/removeAll", removeAllItems)
	http.HandleFunc("/removeName", removeItemByName)
	http.HandleFunc("/getmarket", getItemsInSupermarket)
	http.HandleFunc("/totalprice", getItemsTotalPrice)

	http.HandleFunc("/complete", completeHandler)
	http.HandleFunc("/incomplete", incompleteHandler)
	http.HandleFunc("/goget", getHandler)
	
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
	Name     string `json:"name"`
	Price int    `json:"price"`
	SuperMarket string    `json:"superMarket"`
}

func addItem(w http.ResponseWriter, r *http.Request){
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


	for index, e := range items  {
		if(e.Name == name){
			items = remove(items, index)
		}
	}

}

func removeAllItems(w http.ResponseWriter, r *http.Request){
	items = []Item{}
}

func getItemsTotalPrice(w http.ResponseWriter, r *http.Request) {
	sum := 0
	for _, e := range items  {
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
	for _, e := range items  {
		if(e.SuperMarket == smarket){
			sum = append(sum, e)
		}
	}

	fmt.Fprintf(w, "Items in %s : %v ", smarket, sum)
}