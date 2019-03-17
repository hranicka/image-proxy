package main

import (
	"flag"
	"fmt"
	"image/color"
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
	paramSource = flag.String("param-source", "url", "Source URL query string parameter")
	paramWidth  = flag.String("param-width", "w", "Width query string parameter")
	paramHeight = flag.String("param-height", "h", "Height query string parameter")
)

func main() {
	flag.Parse()

	timeoutSec, _ := strconv.Atoi(*timeout)
	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// parse and validate parameters
		srcURL := r.URL.Query().Get(*paramSource)
		if srcURL == "" {
			http.Error(w, "source url parameter is required", http.StatusBadRequest)
			return
		}

		width, _ := strconv.Atoi(r.URL.Query().Get(*paramWidth))
		height, _ := strconv.Atoi(r.URL.Query().Get(*paramHeight))
		if width < 0 || height < 0 {
			http.Error(w, "width or height parameter is invalid", http.StatusBadRequest)
			return
		}

		log.Printf("src %s, size %dx%d\n", srcURL, width, height)

		// parse and validate source url
		u, err := url.Parse(srcURL)
		if err != nil {
			http.Error(w, "cannot parse url: "+err.Error(), http.StatusBadRequest)
			return
		}
		if !u.IsAbs() {
			http.Error(w, "absolute url is required", http.StatusBadRequest)
			return
		}

		// load source file
		resp, err := client.Get(srcURL)
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
		srcImg, err := imaging.Decode(resp.Body)
		if err != nil {
			http.Error(w, "failed to open image: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// resize if needed
		dstImg := srcImg
		if width > 0 && height > 0 {
			dstImg = imaging.New(width, height, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
			dstImg = imaging.PasteCenter(dstImg, imaging.Fit(srcImg, width, height, imaging.Lanczos))
		}

		w.Header().Set("Content-Type", "image/jpg")
		if err := imaging.Encode(w, dstImg, imaging.JPEG); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	fmt.Println("listening on", *listenAddr)
	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		log.Fatalln(err)
	}
}
