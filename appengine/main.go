package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/logging"
	"cloud.google.com/go/storage"
)

var (
	startupTime time.Time
	client      *storage.Client
	logger      *logging.Logger
)

func main() {
	if err := run(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	startupTime = time.Now()

	logClient, err := setupLogger(context.Background())
	if err != nil {
		return fmt.Errorf("setupLogger: %w", err)
	}
	defer logClient.Close()

	logger = logClient.Logger("app")

	if err := setup(context.Background()); err != nil {
		return fmt.Errorf("setup: %w", err)
	}

	http.HandleFunc("/_ah/warmup", warmupHandler)
	http.HandleFunc("/_ah/mail/", mailHandler)
	http.HandleFunc("/", indexHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log(fmt.Sprintf("defaulting to port %s", port))
	}

	log(fmt.Sprintf("listening on port %s. uptime: %s\n", port, time.Since(startupTime)))
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		return fmt.Errorf("http.ListenAndServe: %w", err)
	}

	return nil
}

func setupLogger(ctx context.Context) (*logging.Client, error) {
	fmt.Fprint(os.Stdout, "start setupLogger\n")

	client, err := logging.NewClient(ctx, fmt.Sprintf("projects/%s", os.Getenv("GOOGLE_CLOUD_PROJECT")))
	if err != nil {
		return nil, fmt.Errorf("logging.NewClient: %w", err)
	}

	if err := client.Ping(ctx); err != nil {
		return nil, fmt.Errorf("client.Ping: %w", err)
	}

	client.OnError = func(e error) {
		fmt.Fprintf(os.Stderr, "logging: %v\n", e)
	}

	fmt.Fprint(os.Stdout, "finished setupLogger\n")
	return client, nil
}

func log(payload interface{}) {
	logger.Log(logging.Entry{Payload: payload})
}

func setup(ctx context.Context) error {
	log("start setup")

	var err error
	client, err = storage.NewClient(ctx)
	if err != nil {
		return err
	}

	log("finished setup")
	return nil
}

func warmupHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r, fmt.Sprintf("warmup done. uptime: %s\n", time.Since(startupTime)))
}

func logRequest(r *http.Request, payload interface{}) {
	logger.Log(logging.Entry{
		Payload: payload,
		HTTPRequest: &logging.HTTPRequest{
			Request: r,
		},
	})
}

func mailHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if strings.HasPrefix(r.URL.Path, "/_ah/mail/500") {
		errorHandler(w, r, errors.New("error response test"))
		return
	}

	name := fmt.Sprintf("%d.eml", time.Now().UnixNano())
	logRequest(r, fmt.Sprintf("Start receiving a mail. name: %s", name))

	obj := client.Bucket(os.Getenv("RAW_EML_BUCKET")).Object(name)
	ctx := context.Background()
	wc := obj.NewWriter(ctx)

	if _, err := io.Copy(wc, r.Body); err != nil {
		errorHandler(w, r, fmt.Errorf("error saving to the storage: %w", err))
		return
	}

	if err := wc.Close(); err != nil {
		errorHandler(w, r, fmt.Errorf("error closing storage.(*Writer): %v\n", err))
		return
	}

	logRequest(r, fmt.Sprintf("received a mail. name: %s\n", name))
	dump, _ := httputil.DumpRequest(r, false)
	logRequest(r, fmt.Sprintf("request: %s", dump))
}

func errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	http.Error(w, http.StatusText(status), status)
	dump, _ := httputil.DumpRequest(r, false)
	logger.Log(logging.Entry{
		Payload: fmt.Sprintf("error: %v request: %s", err, dump),
		HTTPRequest: &logging.HTTPRequest{
			Request: r,
			Status:  status,
		},
	})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	logRequest(r, fmt.Sprintf("OK. uptime: %s\n", time.Since(startupTime)))
}
