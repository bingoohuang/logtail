package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

func main() {
	addr := flag.String("addr", ":8812", "listen address")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Save a copy of this request for debugging.
		requestDump, err := httputil.DumpRequest(r, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(requestDump))

		_, _ = fmt.Fprintln(w, "OK")
	})

	log.Println("start mock server on", *addr)

	log.Fatal(http.ListenAndServe(*addr, nil))
}
