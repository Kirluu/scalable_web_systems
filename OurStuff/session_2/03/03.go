package main

import (
	"net/http"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"io/ioutil"
	"net/url"
)

func helloName(w http.ResponseWriter, r *http.Request) {
	//name := mux.Vars(r)["name"]
	paramHandler(w, r)
	//fmt.Fprintln(w, "hello,", name)
	//bodyHandler(w, r)
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/hello", helloName).Methods("GET", "POST", "PUT", "DELETE")

	// handle all requests with the Gorilla router.
	http.Handle("/", r)
	if err := http.ListenAndServe("127.0.0.1:5080", nil); err != nil {
		log.Fatal(err)
	}
}

func paramHandler(w http.ResponseWriter, r *http.Request) {
	//name := r.FormValue("name")
	name := r.Form.Get("name")
	if name == "" {
		name = "friend"
	}
	//fmt.Fprintf(w, "Hello, %s!", name)

	u, _ := url.Parse(r.URL.String())

	qpars := u.Query()

	for k, e := range qpars {

		fmt.Fprintf(w,"%v", e)
		//fmt.Fprintf(w,e[1])
		fmt.Fprint(w, k)


	}

}

func bodyHandler(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "could not read body: %v", err)
		return
	}
	name := string(b)
	if name == "" {
		name = "friend"
	}
	fmt.Fprintf(w, "Hello, %s!", name)
}