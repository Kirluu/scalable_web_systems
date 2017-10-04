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
<<<<<<< HEAD
	http.HandleFunc("/query", queryHandler)
=======
	
>>>>>>> 7d83d2ac66bc1e93c859c64d792860c4fb7696b1
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

	p := &Item{Name: "gopher", SuperMarket: "Netto", Price: 5}

	// create a new complete key of kind Person and value gopher.
	key := datastore.NewKey(ctx, "Item", "gopher", 0, nil)
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

	p := &Item{Name: "gopher", SuperMarket: "Netto", Price: 5}

	// create a new complete key of kind Person.
	key := datastore.NewIncompleteKey(ctx, "Item", nil)
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

	key := datastore.NewKey(ctx, "Item", "gopher", 0, nil)

	var p Item
	err := datastore.Get(ctx, key, &p)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	fmt.Fprintln(w, p)
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	var p []Item

	// create a new query on the kind Person
	q := datastore.NewQuery("Item")

	// select only values where field Age is 10 or lower
	q = q.Filter("Price =", 0)

	// order all the values by the Name field
	q = q.Order("Name")

	// and finally execute the query retrieving all values into p.
	_, err := q.GetAll(ctx, &p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, p)
}

type Item struct {
	Name     string `json:"name"`
	Price int    `json:"price"`
	SuperMarket string    `json:"superMarket"`
}

func addItem(w http.ResponseWriter, r *http.Request){
	var i Item

	// PARSE JSON
	dec := json.NewDecoder(r.Body)
	if dec == nil {
		fmt.Fprintf(w, "json decoder failed to be created, given body")
	}
	err := dec.Decode(&i)
	if err != nil {
		fmt.Fprintf(w, "decoding resulted in an error! Check your format!")
		//http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// PUT IN DB
	ctx := appengine.NewContext(r)
	key := datastore.NewIncompleteKey(ctx, "Item", nil)
	key, putErr := datastore.Put(ctx, key, &i)
	if putErr != nil {
		fmt.Fprintf(w, "Put failed!")
		http.Error(w, putErr.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "gopher stored with key %v", key)
}

func removeItemByName(w http.ResponseWriter, r *http.Request) {
	b, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		fmt.Fprintf(w, "could not read body: %v", readErr)
		return
	}
	name := string(b)

	ctx := appengine.NewContext(r)

	var keyArray []*datastore.Key
	_, err := datastore.NewQuery("Item").Filter("Name =", name).KeysOnly().GetAll(ctx, &keyArray)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	deleteErr := datastore.DeleteMulti(ctx, keyArray)
	if deleteErr != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func removeAllItems(w http.ResponseWriter, r *http.Request){
	ctx := appengine.NewContext(r)

	var keyArray []*datastore.Key
	_, err := datastore.NewQuery("Item").KeysOnly().GetAll(ctx, &keyArray)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	deleteErr := datastore.DeleteMulti(ctx, keyArray)
	if deleteErr != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getItemsTotalPrice(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	var p []Item

	// create a new query on the kind Person
	q := datastore.NewQuery("Item")

	// and finally execute the query retrieving all values into p.
	_, err := q.GetAll(ctx, &p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sum := 0
	for _, e := range p  {
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

	ctx := appengine.NewContext(r)

	var items []Item

	// create a new query on the kind Person
	q := datastore.NewQuery("Item").Filter("SuperMarket =", smarket)

	// and finally execute the query retrieving all values into p.
	_, getErr := q.GetAll(ctx, &items)
	if getErr != nil {
		http.Error(w, getErr.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Items with supermarket: '%s' : %v ", smarket, items)
}