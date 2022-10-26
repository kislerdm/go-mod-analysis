package main

import (
	"context"
	"gomodanalysis/indexmodules"
	"log"
	"os"
)

func main() {
	c, err := indexmodules.NewConfigWriter()
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	writer, err := indexmodules.NewWriter(ctx, c)
	if err != nil {
		log.Fatalln(err)
	}

	dataset := os.Getenv("DATASET")
	table := os.Getenv("TABLE")

	if dataset == "" {
		log.Fatalln("env variable DATASET must be set")
	}

	if table == "" {
		log.Fatalln("env variable TABLE must be set")
	}

	pathOut := "datasets/" + dataset + "/tables/" + table

	reader := indexmodules.NewReader()

	q := map[string]string{"since": indexmodules.GetLastPaginationIndex()}

	for {
		resp, err := reader.Fetch(q)
		if err != nil {
			log.Fatalln(err)
		}

		if resp == nil {
			log.Println("done")
			break
		}

		d, err := resp.Decode()
		if err != nil {
			log.Fatalln(err)
		}

		output, err := indexmodules.ConvertToStoreFormat(d)
		if err != nil {
			log.Fatalln(err)
		}

		rowsRecorded, err := writer.Store(ctx, output, pathOut)
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("%v recorded\n", rowsRecorded)

		q["since"] = d[len(d)-1].Timestamp
	}
}
