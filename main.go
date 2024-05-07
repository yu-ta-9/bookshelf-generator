package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Cannot load .env: %v", err)
	}

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello World!"})
	})

	r.GET("/", func(c *gin.Context) {
		book, err := getBookData()
		fmt.Printf("book return: %v\n", book)
		if err != nil {
			fmt.Printf("error fetching book data, %v", err)
			return
		}

		c.Writer.Header().Set("Content-Type", "image/svg+xml")
		c.HTML(http.StatusOK, "bookshelf.html", gin.H{
			"items": book.Items,
		})
	})

	r.Run(":8080")
}

type Book struct {
	Items []struct {
		Item struct {
			Title          string `json:"title"`
			MediumImageUrl string `json:"mediumImageUrl"`
		}
	} `json:"items"`
	Count int `json:"pageCount"`
}

// getBookData fetches book data
func getBookData() (Book, error) {
	key := os.Getenv("RAKUTEN_BOOKS_API_KEY")

	// url := "https://api.openbd.jp/v1/get?isbn=9784798178639"
	url := "https://app.rakuten.co.jp/services/api/BooksBook/Search/20170404?format=json&outOfStockFlag=1" + "&isbn[]=" + "9784798178189" + "&isbn[]=" + "9784774189673" + "&applicationId=" + key

	var book Book

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(body, &book)
	if err != nil {
		panic(err)
	}

	return book, nil
}
