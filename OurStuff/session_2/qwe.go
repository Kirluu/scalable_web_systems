package main

import (
	"fmt"
	"net/http"
	"log"
	"strings"
	"io/ioutil"
	"encoding/json"
)






func main() {
	http.Handle("/hello", helloHandler{})
	http.Handle("/bye", byeHandler{})
	http.HandleFunc("/", decodeHandler)

	err := http.ListenAndServe("localhost:5080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

type helloHandler struct {

}

func (h helloHandler)ServeHTTP(w http.ResponseWriter, r *http.Request){
	fmt.Fprintln(w, "Hello, web")
}

type byeHandler struct {

}

func (h byeHandler)ServeHTTP(w http.ResponseWriter, r *http.Request){
	fmt.Fprintln(w, "Bye, web")
}





func main2() {
	// try changing the value of this url
	res, err := http.Get("https://golang.org/foo")
	if err != nil {
		log.Fatal(err)
	}

	if(strings.Contains(res.Status, "404")){
		fmt.Println(res.Status)
		//fmt.Println(ioutil.ReadAll(res.Body))
		var q, _ = ioutil.ReadAll(res.Body)
		fmt.Println(q)
		res.Body.Close()
	}
}

func main3(){
	//doPut()
	//doGet()
	doGetEmbeddedQueryString()
}

func doGet() {
	req, err := http.NewRequest("GET", "https://http-methods.appspot.com/johaa/Message", nil)
	if err != nil {
		log.Fatalf("could not create request: %v", err)
	}
	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("http request failed: %v", err)
	}
	fmt.Println(res.Status)
	var q, _ = ioutil.ReadAll(res.Body)
	fmt.Println(string(q))
	res.Body.Close()
}



func doPut() {
	req, err := http.NewRequest("PUT", "https://http-methods.appspot.com/johaa/Message", strings.NewReader("Our beatifull string"))
	if err != nil {
		log.Fatalf("could not create request: %v", err)
	}
	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("http request failed: %v", err)
	}
	fmt.Println(res.Status)
}

func doGetEmbeddedQueryString() {
	req, err := http.NewRequest("GET", "https://http-methods.appspot.com/Hungary/", nil)
	query := req.URL.Query()
	query.Add("v", "true")
	req.URL.RawQuery = query.Encode()
	if err != nil {
		log.Fatalf("could not create request: %v", err)
	}
	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("http request failed: %v", err)
	}
	fmt.Println(res.Status)
	var q, _ = ioutil.ReadAll(res.Body)
	fmt.Println(string(q))
	res.Body.Close()
}

func decodeHandler(w http.ResponseWriter, r *http.Request) {
	var p Person

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Name is %v and age is %v", p.Name, p.AgeYears)
}

type Person struct {
	Name     string `json:"name"`
	AgeYears int    `json:"age_years"`
}