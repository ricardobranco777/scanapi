package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

import flag "github.com/spf13/pflag"

const version = "0.1.0"

var services = map[string]struct {
	Method string
	Route  string
}{
	"Bugzilla":     {http.MethodGet, "rest/version"},
	"Distribution": {http.MethodGet, "v2/"},
	"Gitea":        {http.MethodGet, "api/v1/version"},
	"GitLabv4":     {http.MethodGet, "api/v4/version"},
	"Jira":         {http.MethodGet, "rest/api/2/serverInfo"},
	"Pagure":       {http.MethodGet, "api/0/version"},
	"RedMine":      {http.MethodHead, "issues.json"},
}

var validStatuses = []int{http.StatusOK, http.StatusUnauthorized}

var debug bool

func checkVersion(ctx context.Context, client *http.Client, headers map[string]string, url string, service string) error {
	api := services[service]

	url = fmt.Sprintf("%s/%s", url, api.Route)
	req, err := http.NewRequestWithContext(ctx, api.Method, url, nil)
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if debug {
		logResponse(resp)
	}

	if !slices.Contains(validStatuses, resp.StatusCode) {
		return nil
	}

	if api.Method == http.MethodGet {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		fmt.Printf("%s: %s\n", service, string(body))
	} else {
		fmt.Printf("%s\n", service)
	}

	return nil
}

func logResponse(resp *http.Response) {
	dump, err := httputil.DumpRequestOut(resp.Request, true)
	if err != nil {
		log.Print(err)
	} else {
		fmt.Fprintf(os.Stderr, "\n%s", string(dump))
	}

	dump, err = httputil.DumpResponse(resp, true)
	if err != nil {
		log.Print(err)
	} else {
		fmt.Fprintf(os.Stderr, "\n%s\n", string(dump))
	}
}

func init() {
	log.SetFlags(0)
	log.SetPrefix("ERROR: ")
}

func main() {
	var headerValues []string
	var timeoutInt int

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] URL\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.IntVarP(&timeoutInt, "timeout", "t", 60, "Timeout")
	flag.StringSliceVarP(&headerValues, "header", "H", nil, "HTTP header (may be specified multiple times")

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	url := flag.Args()[0]

	debug = os.Getenv("DEBUG") != ""

	headers := map[string]string{
		"User-Agent": "scanapi/" + version,
	}
	for _, header := range headerValues {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) != 2 {
			log.Printf("Invalid header: %s", header)
			os.Exit(1)
		}
		headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	timeout := time.Duration(timeoutInt)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := &http.Client{
		Timeout: timeout * time.Second,
	}

	var wg sync.WaitGroup

	for service := range services {
		wg.Add(1)
		go func(service string) {
			defer wg.Done()
			checkVersion(ctx, client, headers, url, service)
		}(service)
	}

	wg.Wait()
}
