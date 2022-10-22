package main

import (
	"bytes"
	app "gomodanalysis"
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
		Backoff: &app.Backoff{
			MaxDelay: 4 * time.Second,
			MaxSteps: 2,
		},
		Verbose: true,
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

	skipped := map[string]bool{}

	fetcher := func(pageStr string) bool {
		log.Println("fetch page" + pageStr)
		resp, err := client.Fetch(constructQuery(pageStr))
		switch err.(type) {
		case nil:
			delete(skipped, pageStr)
		case app.TimeoutError:
			log.Printf(err.Error() + " skip\n")
			skipped[pageStr] = true
			return false
		default:
			log.Fatalln(err)
		}

		var buf bytes.Buffer
		_, err = buf.ReadFrom(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		objPath := destinationDir + "page_" + pageStr

		if err := storeObject(buf.Bytes(), objPath+".html"); err != nil {
			log.Fatalln(err)
		}

		if doesNotContainData.Match(buf.Bytes()) {
			log.Println("stop")
			return true
		}

		return false
	}

	page := 1
	if os.Getenv("PAGE") != "" {
		page, _ = strconv.Atoi(os.Getenv("PAGE"))
	}

	for {
		pageStr := strconv.Itoa(page)
		if fetcher(pageStr) {
			break
		}
		page++
	}

	for len(skipped) > 0 {
		for pageStr, _ := range skipped {
			fetcher(pageStr)
		}
	}

}

func constructQuery(s string) string {
	return "https://github.com/search?o=desc&q=module+extension%3Amod+language%3AText&s=indexed&type=Code&p=" + s
}

func storeObject(data []byte, path string) error {
	return os.WriteFile(path, data, 0777)
}
