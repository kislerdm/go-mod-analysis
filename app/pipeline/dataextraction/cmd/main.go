package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/kislerdm/gomodanalysis/app/pipeline"
	"github.com/kislerdm/gomodanalysis/app/pipeline/dataextraction"
)

var (
	client    pipeline.GBQClient
	storePath string
)

func init() {
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		Log.Fatal("PROJECT_ID env variable must be set")
	}

	storePath = os.Getenv("STORE_PATH")
	if storePath == "" {
		Log.Fatal("STORE_PATH env variable must be set")
	}

	var err error
	client, err = pipeline.NewGBQClient(context.Background(), projectID)
	if err != nil {
		Log.Fatal("cannot init gbq client: " + err.Error())
	}
}

type logger struct {
	wOut io.Writer
	wErr io.Writer
}

func (l *logger) Info(msg string) {
	l.printer("INFO", l.wOut, msg)
}

func (l *logger) Error(msg string) {
	l.printer("ERROR", l.wErr, msg)
}

func (l *logger) Fatal(msg string) {
	l.printer("FATAL", l.wErr, msg)
	os.Exit(1)
}

func (l *logger) Debug(msg string) {
	l.printer("DEBUG", l.wOut, msg)
}

func (l *logger) Warning(msg string) {
	l.printer("WARNING", l.wOut, msg)
}

func (l *logger) time() string {
	return "[" + time.Now().UTC().Format(time.RFC3339Nano) + "]"
}

func (l *logger) printer(prefix string, w io.Writer, msg string) {
	msg = l.time() + "[" + prefix + "] " + msg
	if _, err := fmt.Fprintln(w, msg); err != nil {
		panic(err)
	}
}

var Log = logger{os.Stdout, os.Stderr}

func main() {
	defer func() { _ = client.Close() }()

	Log.Info("start")

	t0 := time.Now()

	listModules, err := dataextraction.ListModulesToFetch(context.Background(), client)
	if err != nil {
		Log.Fatal("error fetching list of modules: " + err.Error())
	}
	Log.Info(
		strconv.Itoa(len(listModules)) + " modules found. elapsed: " + strconv.FormatInt(
			time.Since(t0).Milliseconds(), 10,
		) + " ms.",
	)

	if len(listModules) == 0 {
		Log.Info("done.")
		os.Exit(0)
	}

	cntWorkers := 20
	if c, err := strconv.Atoi(os.Getenv("WORKERS")); err == nil {
		cntWorkers = c
	}

	goPkgClient := dataextraction.NewGoPackagesClient(
		&http.Client{
			Transport: &http.Transport{
				TLSHandshakeTimeout: 30 * time.Second,
				DisableKeepAlives:   false,
				DisableCompression:  false,
			},
			Timeout: 60 * time.Second,
		},
		30,
	)

	var wg sync.WaitGroup
	pool := make(chan struct{}, cntWorkers)
	for _, m := range listModules {
		wg.Add(1)
		pool <- struct{}{}
		go func(m dataextraction.Module, wg *sync.WaitGroup, writerClient pipeline.GBQClient) {
			defer func() { wg.Done(); <-pool }()

			Log.Info("[pkg:" + m.Name + "] fetch start")
			t0 := time.Now()

			o, err := dataextraction.ExtractGoPkgData(m.Name, m.Version, goPkgClient)

			Log.Info(
				"[pkg:" + m.Name + "] fetch ended after " + strconv.FormatInt(
					time.Since(t0).Milliseconds(), 10,
				) + " ms.",
			)

			srore := func(m dataextraction.Module, o dataextraction.PkgData) error {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				Log.Info("[pkg:" + m.Name + "] store start")
				t0 = time.Now()

				if err := client.Write(ctx, o, storePath); err != nil {
					return err
				}

				Log.Info(
					"[pkg:" + m.Name + "] store ended after " + strconv.FormatInt(
						time.Since(t0).Milliseconds(), 10,
					) + " ms.",
				)

				return nil
			}

			switch err.(type) {
			case nil:
				if err := srore(m, o); err != nil {
					Log.Error("[pkg:" + m.Name + "] gbq store error: " + err.Error())
					break
				}

			case dataextraction.ErrExtractGoPkgData:
				if !err.(dataextraction.ErrExtractGoPkgData).IsHTTPStatus(http.StatusNotFound) {
					Log.Error("[pkg:" + m.Name + "] fetch error: " + err.Error())
					return
				}

				Log.Warning("[pkg:" + m.Name + "] fetch error: " + err.Error())
				if err := srore(m, o); err != nil {
					Log.Error("[pkg:" + m.Name + "] gbq store error: " + err.Error())
					break
				}

			default:
				Log.Error("[pkg:" + m.Name + "] fetch error:\n" + err.Error())
			}

		}(m, &wg, client)
	}
	wg.Wait()
}
