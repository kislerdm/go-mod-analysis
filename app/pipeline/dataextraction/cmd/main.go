package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/kislerdm/gomodanalysis/app/pipeline"
	"github.com/kislerdm/gomodanalysis/app/pipeline/dataextraction"
)

var client pipeline.GBQClient

func init() {
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		log.Fatalln("PROJECT_ID env variable must be set")
	}

	var err error
	client, err = pipeline.NewGBQClient(context.Background(), projectID)
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	defer func() { _ = client.Close() }()

	listModules, err := dataextraction.ListModulesToFetch(client)
	if err != nil {
		log.Fatalln(err)
	}

	if len(listModules) == 0 {
		log.Println("done")
		os.Exit(0)
	}

	cntWorkers := 20
	if c, err := strconv.Atoi(os.Getenv("WORKERS")); err == nil {
		cntWorkers = c
	}

	goPkgClient := dataextraction.GoPackagesClient{
		&http.Client{
			Transport: &http.Transport{
				TLSHandshakeTimeout: 30 * time.Second,
				DisableKeepAlives:   false,
				DisableCompression:  false,
				MaxIdleConns:        cntWorkers,
			},
			Timeout: 60 * time.Second,
		},
	}

	var wg sync.WaitGroup
	data := make(chan dataextraction.PkgData, cntWorkers)

	for _, m := range listModules {
		wg.Add(1)
		go func(m dataextraction.Module, data chan dataextraction.PkgData, wg *sync.WaitGroup) {
			defer wg.Done()
			o, _ := dataextraction.ExtractGoPkgData(m.Name, m.Version, &goPkgClient)
			data <- o
		}(m, data, &wg)

	}
}
