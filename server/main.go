package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", queryViews)
	log.Fatal(http.ListenAndServe(":80", nil))
}
