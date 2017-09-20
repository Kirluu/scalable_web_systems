package main

import (
	"net/http"
	"log"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func main() {
	http.HandleFunc("/add", addItem)
	http.HandleFunc("/removeAll", removeAllItems)
	http.HandleFunc("/removeName", removeItemByName)
	http.HandleFunc("/getmarket", getItemsInSupermarket)
	http.HandleFunc("/totalprice", getItemsTotalPrice)

	err := http.ListenAndServe("localhost:5080", nil)
	if err != nil {
		log.Fatal(err)
	}
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