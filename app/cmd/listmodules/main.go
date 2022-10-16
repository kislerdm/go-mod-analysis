package main

import (
	"bytes"
	"encoding/json"
	app "gomodanalysis"
	"gomodanalysis/cmd/listmodules/parsehtml"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

func main() {
	userSession := os.Getenv("GITHUB_SESSION")
	userName := os.Getenv("GITHUB_LOGIN")

	if userSession == "" || userName == "" {
		log.Fatalln("env variables GITHUB_SESSION and GITHUB_LOGIN must be set")
	}

	destinationDir := os.Getenv("DIR_STORE")
	if destinationDir == "" {
		log.Fatalln("env variable DIR_STORE must be set")
	}

	if destinationDir[len(destinationDir)-1] != '/' {
		destinationDir = destinationDir + "/"
	}

	var doesNotContainData = regexp.MustCompile("next_page disabled")

	client := app.NewClient(app.Configuration{
		Cookies: []*http.Cookie{
			{
				Name:     "logged_id",
				Value:    "yes",
				Path:     "/",
				Domain:   ".github.com",
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			},
			{
				Name:     "user_session",
				Value:    userSession,
				Path:     "/",
				Domain:   "github.com",
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			},
			{
				Name:     "__Host-user_session_same_site",
				Value:    userSession,
				Path:     "/",
				Domain:   ".github.com",
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			},
			{
				Name:     "dotcom_user",
				Value:    userName,
				Path:     "/",
				Domain:   ".github.com",
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			},
		},
	})

	page := 90
	for {
		pageStr := strconv.Itoa(page)

		log.Println("fetch page" + pageStr)

		query := "https://github.com/search?o=desc&q=module+extension%3Amod+language%3AText&s=indexed&type=Code&p=" + pageStr

		resp, err := client.Fetch(query)
		if err != nil {
			if resp.StatusCode == 429 || resp.StatusCode == 400 {
				log.Println("too many requests, delay and retry")
			} else {
				log.Fatalln(err)
			}
			continue
		}

		var buf bytes.Buffer
		_, err = buf.ReadFrom(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		if doesNotContainData.Match(buf.Bytes()) {
			log.Println("stop")
			break
		}

		objPath := destinationDir + "page_" + pageStr

		if err := storeObject(buf.Bytes(), objPath+".html"); err != nil {
			log.Fatalln(err)
		}

		searchResults, err := parsehtml.ParseHtml(resp.Body)
		if err != nil {
			log.Println("error parsing page " + pageStr + " err: " + err.Error())
		}

		if len(searchResults) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		b, err := json.Marshal(searchResults)
		if err != nil {
			log.Fatalln("error marshaling search results for page " + pageStr + ": " + err.Error())
		}

		if err := storeObject(b, objPath+".json"); err != nil {
			log.Fatalln(err)
		}

		page++
	}
}

func storeObject(data []byte, path string) error {
	return os.WriteFile(path, data, 0777)
}
