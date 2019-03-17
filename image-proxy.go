package main

import (
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
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
		img, format, err := image.Decode(resp.Body)
		if err != nil {
			http.Error(w, "image: "+err.Error(), http.StatusInternalServerError)
			return
		}

		switch format {
		case "jpeg":
			w.Header().Set("Content-Type", "image/jpg")
			if err := jpeg.Encode(w, img, &jpeg.Options{Quality: 85}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "gif":
			w.Header().Set("Content-Type", "image/gif")
			if err := gif.Encode(w, img, nil); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "png":
			w.Header().Set("Content-Type", "image/png")
			if err := png.Encode(w, img); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, "unsupported format: "+format, http.StatusInternalServerError)
			return
		}
	})

	fmt.Println("listening on", *listenAddr)
	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		log.Fatalln(err)
	}
}
