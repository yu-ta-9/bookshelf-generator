package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func init() {
	// load env
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Cannot load .env: %v", err)
	}

	// connect to redis
	url := os.Getenv("REDIS_URL")
	opts, err := redis.ParseURL(url)
	if err != nil {
		panic(err)
	}
	redisClient = redis.NewClient(opts)
}

func main() {
	r := gin.Default()
	// 関数の登録
	r.SetFuncMap(template.FuncMap{
		"mul": func(a, b int) int {
			return a * b
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
	})
	r.LoadHTMLGlob("templates/*")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello World!"})
	})

	r.GET("/", func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")

		isbns := c.QueryArray("isbns[]")
		if len(isbns) == 0 {
			c.JSON(http.StatusOK, gin.H{})
			return
		}

		books, err := getBookData(c, isbns)
		if err != nil {
			fmt.Printf("error fetching book data, %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "error fetching book data"})
			return
		}

		c.Header("Content-Type", "image/svg+xml")
		c.HTML(http.StatusOK, "bookshelf.html", gin.H{
			"books":      books,
			"bookLength": len(books),
		})
	})

	r.Run(":8080")
}

type Book struct {
	Title       string
	ImageBase64 string
	InfoLink    string
}

// getBookData fetches book data
func getBookData(c *gin.Context, isbns []string) ([]Book, error) {
	key := os.Getenv("GOOGLE_BOOKS_API_KEY")
	referer := os.Getenv("HOST_URL")

	var books []Book

	// MEMO: Use pipeline to reduce the number of round trips
	pipe := redisClient.Pipeline()
	cmds := map[string]*redis.StringCmd{}
	for _, isbn := range isbns {
		cmds[isbn] = pipe.Get(c, isbn)
	}
	// fetch all data at once
	pipe.Exec(c)

	for i, isbn := range isbns {
		// MEMO: 最大5冊まで取得
		if i == 5 {
			break
		}

		var book Book
		var data []byte

		val, err := cmds[isbn].Result()

		// MEMO: If there is no cache, get it from API
		if err != nil || val == "" {
			// NOTE:
			// There is a bug in the Google Books API that find no books by ISBN.
			// To avoid this, use the q parameter instead of the `ISBN:` not `isbn:`.
			// https://stackoverflow.com/questions/36142556/google-books-api-not-giving-output
			url := "https://www.googleapis.com/books/v1/volumes?key=" + key + "&q=ISBN:" + isbn

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				fmt.Println("Error creating request:", err)
				continue
			}
			req.Header.Set("Referer", referer)
			client := &http.Client{}
			resp, err := client.Do(req)

			if err != nil || resp.StatusCode != http.StatusOK {
				continue
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				continue
			}

			data = body
			// MEMO: cache book data for 60 days
			// Consider to terms, limited to a maximum of 60 days.
			redisClient.Set(c, isbn, body, 24*time.Hour*60)
		} else {
			data = []byte(val)
		}

		err = json.Unmarshal(data, &book)
		if err != nil {
			fmt.Printf("error unmarshalling book data, %v isbn = "+isbn, err)
			continue
		}

		books = append(books, book)
	}

	return books, nil
}

// encode encodes image to base64
func encode(url string) string {
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	image, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(image)
}

func (book *Book) UnmarshalJSON(byte []byte) error {
	type GoogleBook struct {
		Items []struct {
			VolumeInfo struct {
				Title      string `json:"title"`
				ImageLinks struct {
					Medium string `json:"thumbnail"`
				} `json:"imageLinks"`
				InfoLink string `json:"infoLink"`
			} `json:"volumeInfo"`
		} `json:"items"`
	}
	var googleBook GoogleBook
	err := json.Unmarshal(byte, &googleBook)
	if err != nil {
		fmt.Println(err)
	}

	if len(googleBook.Items) == 0 {
		return nil
	}

	book.Title = googleBook.Items[0].VolumeInfo.Title
	book.ImageBase64 = encode(googleBook.Items[0].VolumeInfo.ImageLinks.Medium)
	book.InfoLink = googleBook.Items[0].VolumeInfo.InfoLink

	return err
}
