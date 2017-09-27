package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"fmt"
)

func listProducts(w http.ResponseWriter, r *http.Request) {
	// list all products
}

func addProduct(w http.ResponseWriter, r *http.Request) {
	// add a product
}

func getProduct(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["productID"]
	log.Printf("fetching product with ID %q", id)
	// get a specific product
}

func helloName(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	fmt.Fprintln(w, "hello,", name)
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/hello/{name}", helloName).Methods("GET", "POST", "PUT", "DELETE")

	// handle all requests with the Gorilla router.
	http.Handle("/", r)
	if err := http.ListenAndServe("127.0.0.1:5080", nil); err != nil {
		log.Fatal(err)
	}
}