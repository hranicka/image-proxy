package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/disintegration/imaging"
)

var (
	listenAddr  = flag.String("listen", "localhost:8080", "Listening address")
	timeout     = flag.String("timeout", "10", "HTTP timeout in seconds")
	sourceParam = flag.String("source", "url", "Source URL query string parameter")
)

func main() {
	flag.Parse()

	timeoutSec, _ := strconv.Atoi(*timeout)
	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		src := r.URL.Query().Get(*sourceParam)
		if src == "" {
			http.Error(w, "source url parameter is required", http.StatusBadRequest)
			return
		}
		log.Println(src)

		// parse and validate source url
		u, err := url.Parse(src)
		if err != nil {
			http.Error(w, "cannot parse url: "+err.Error(), http.StatusBadRequest)
			return
		}
		if !u.IsAbs() {
			http.Error(w, "absolute url is required", http.StatusBadRequest)
			return
		}

		// load source file
		resp, err := client.Get(src)
		if err != nil {
			http.Error(w, "download failed: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			http.Error(w, "download failed: "+resp.Status, resp.StatusCode)
			return
		}

		// decode source image
		img, err := imaging.Decode(resp.Body)
		if err != nil {
			http.Error(w, "failed to open image: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "image/jpg")
		if err := imaging.Encode(w, img, imaging.JPEG); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	fmt.Println("listening on", *listenAddr)
	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		log.Fatalln(err)
	}
}
