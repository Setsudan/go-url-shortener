package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const idLength = 4

var letterRunes = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

type URLRecord struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

func main() {
	r := gin.Default()

	// Read the stored URLs from a file, if it exists
	urls := make(map[string]string)
	file, err := os.Open("urls.json")
	if err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		var storedURLs []URLRecord
		err := decoder.Decode(&storedURLs)
		if err == nil {
			for _, urlRecord := range storedURLs {
				urls[urlRecord.ID] = urlRecord.URL
			}
		}
	}

	rand.Seed(time.Now().UnixNano())

	// Route to redirect the shortened URL to its original URL
	r.GET("/:id", func(c *gin.Context) {
		id := c.Param("id")
		url, ok := urls[id]
		if ok {
			c.Redirect(http.StatusMovedPermanently, url)
		} else {
			c.String(http.StatusNotFound, "URL not found")
		}
	})

	// Route to shorten a URL
	r.POST("/shorten", func(c *gin.Context) {
		url := c.PostForm("url")
		if url == "" {
			c.String(http.StatusBadRequest, "URL is required")
			return
		}

		// Generate a unique ID
		var id string
		for {
			id = randString(idLength)
			_, ok := urls[id]
			if !ok {
				break
			}
		}

		urls[id] = url

		// Write the updated URLs to a file
		file, err := os.Create("urls.json")
		if err != nil {
			c.String(http.StatusInternalServerError, "Could not write to file")
			return
		}
		defer file.Close()
		encoder := json.NewEncoder(file)
		var urlRecords []URLRecord
		for id, url := range urls {
			urlRecords = append(urlRecords, URLRecord{ID: id, URL: url})
		}
		err = encoder.Encode(urlRecords)
		if err != nil {
			c.String(http.StatusInternalServerError, "Could not write to file")
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": id})
	})

	r.Run(":8080")
}
