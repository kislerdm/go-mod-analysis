package main

import (
	"bytes"
	app "gomodanalysis"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"unsafe"
)

func main() {
	destinationDir := os.Getenv("DIR_STORE")
	if destinationDir == "" {
		log.Fatalln("env variable DIR_STORE must be set")
	}

	if destinationDir[len(destinationDir)-1] != '/' {
		destinationDir = destinationDir + "/"
	}

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	backoff := &app.Backoff{MaxSteps: 15, MaxDelay: 1 * time.Minute}

	const baseURI = "https://index.golang.org/index?limit=2000"

	lastTime := ""
	page := 1
	for {
		if lastTime != "" {
			lastTime = "&since=" + lastTime
		}
		url := baseURI + lastTime

		d, err := backoff.LinearDelay()
		if err != nil {
			log.Fatalln(err)
		}

		if d != 0 {
			log.Printf("delay %v sec. and call", d.Seconds())
		}
		time.Sleep(d)

		resp, err := client.Get(url)
		if err != nil {
			log.Fatalln(err)
		}

		if resp.StatusCode == 429 {
			backoff.UpCounter()
			continue
		}
		backoff.Reset()

		if resp.ContentLength == 0 {
			log.Println("done")
			break
		}

		var buf bytes.Buffer
		if _, err := buf.ReadFrom(resp.Body); err != nil {
			log.Fatalln(err)
		}

		pathOut := destinationDir + "/p_" + strconv.Itoa(page)

		if err := storeObject(buf.Bytes(), pathOut); err != nil {
			log.Fatalln(err)
		}

		lastTime = extractNextLastTime(buf.Bytes())

		page++
	}
}

func extractNextLastTime(v []byte) string {
	if len(v) < 30 {
		return ""
	}

	shift := 0
	if v[len(v)-1] == '\n' {
		shift = 1
	}

	cut := v[len(v)-29-shift : len(v)-2-shift]

	if cut[0] == '"' {
		cut = cut[1:]
	}

	return *(*string)(unsafe.Pointer(&cut))
}

func storeObject(data []byte, path string) error {
	return os.WriteFile(path, data, 0777)
}
