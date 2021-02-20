package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
)

var startupTime time.Time
var client *storage.Client

func main() {
	if err := setup(context.Background()); err != nil {
		log.Fatalf("setup: %v", err)
	}

	http.HandleFunc("/_ah/warmup", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("warmup done. uptime: %s\n", time.Since(startupTime))
	})
	http.HandleFunc("/_ah/mail/", mailHandler)
	http.HandleFunc("/", indexHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	log.Printf("listening on port %s. uptime: %s\n", port, time.Since(startupTime))
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
}

func setup(ctx context.Context) error {
	startupTime = time.Now()
	log.Print("start setup")

	var err error
	if client, err = storage.NewClient(ctx); err != nil {
		return err
	}

	log.Print("finished setup")
	return nil
}

func mailHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	log.Print("Start receiving a mail\n")

	obj := client.Bucket(os.Getenv("BUCKET_NAME")).Object(fmt.Sprintf("mail/%d.eml", time.Now().UnixNano()))
	ctx := context.Background()
	wc := obj.NewWriter(ctx)

	if _, err := io.Copy(wc, r.Body); err != nil {
		log.Printf("error saving to the storage: %v\n", err)
		return
	}

	if err := wc.Close(); err != nil {
		log.Printf("error closing storage.(*Writer): %v\n", err)
		return
	}

	log.Printf("received a mail. name: %s\n", obj.ObjectName())
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	fmt.Fprintf(w, "OK. uptime: %s\n", time.Since(startupTime))
}
