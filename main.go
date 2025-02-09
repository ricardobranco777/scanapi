package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"runtime"
	"slices"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

import flag "github.com/spf13/pflag"

const version = "0.2.0"

const maxBodySize = 1024

var services = map[string]struct {
	Method     string
	Route      string
	VersionKey string
}{
	"Bugzilla":        {http.MethodGet, "rest/version", "version"},
	"Docker Registry": {http.MethodGet, "v2/_catalog", ""},
	"Gitea":           {http.MethodGet, "api/v1/version", "version"},
	"GitLabv4":        {http.MethodGet, "api/v4/version", "version"},
	"Jira":            {http.MethodGet, "rest/api/2/serverInfo", "version"},
	"Pagure":          {http.MethodGet, "api/0/version", "version"},
	"Redmine":         {http.MethodGet, "issues.json", ""},
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
		limitedReader := io.LimitReader(resp.Body, maxBodySize)
		buf, err := io.ReadAll(limitedReader)
		if err != nil {
			return err
		}
		body := strings.TrimSpace(string(buf))
		// Skip non JSON
		if !strings.HasPrefix(body, "{") {
			return nil
		}
		if api.VersionKey != "" {
			body = strings.ReplaceAll(body, "\n", "")
			fmt.Printf("%s: %s\n", service, body)
		} else {
			fmt.Printf("%s\n", service)
		}
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
	var opts struct {
		header  []string
		timeout int
		version bool
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] URL\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringSliceVarP(&opts.header, "header", "H", nil, "HTTP header (may be specified multiple times")
	flag.IntVarP(&opts.timeout, "timeout", "t", 60, "timeout")
	flag.BoolVarP(&opts.version, "version", "", false, "print version and exit")
	flag.Parse()

	if opts.version {
		fmt.Printf("scanapi v%s %v %s/%s\n", version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	u, err := url.ParseRequestURI(flag.Args()[0])
	if err != nil {
		log.Fatal(err)
	}

	debug = os.Getenv("DEBUG") != ""

	headers := map[string]string{
		"User-Agent": "scanapi/" + version,
	}
	for _, header := range opts.header {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) != 2 {
			log.Fatalf("Invalid header: %s", header)
		}
		headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	timeout := time.Duration(opts.timeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := &http.Client{
		Timeout: timeout,
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(min(10, len(services)))

	for service := range services {
		service := service
		g.Go(func() error {
			err := checkVersion(ctx, client, headers, u.String(), service)
			return err
		})
	}
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
