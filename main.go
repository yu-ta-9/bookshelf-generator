package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
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
	// 関数の登録
	r.SetFuncMap(template.FuncMap{
		"mul": func(a, b int) int {
			return a * b
		},
		"add": func(a, b int) int {
			return a + b
		},
	})
	r.LoadHTMLGlob("templates/*")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello World!"})
	})

	r.GET("/", func(c *gin.Context) {
		isbns := c.QueryArray("isbns[]")
		if len(isbns) == 0 {
			c.JSON(http.StatusOK, gin.H{})
			return
		}

		books, err := getBookData(isbns)
		if err != nil {
			fmt.Printf("error fetching book data, %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "error fetching book data"})
			return
		}

		c.Header("Content-Type", "image/svg+xml")
		c.HTML(http.StatusOK, "bookshelf.html", gin.H{
			"books": books,
			"bookLength": len(books),
		})
	})

	r.Run(":8080")
}

type Book struct {
	Title       string
	ImageBase64 string
}

// getBookData fetches book data
func getBookData(isbns []string) ([]Book, error) {
	key := os.Getenv("GOOGLE_BOOKS_API_KEY")

	var books []Book

	for i, isbn := range isbns {
		// MEMO: 最大5冊まで取得
		if i == 5 {
			break
		}

		var book Book

		url := "https://www.googleapis.com/books/v1/volumes?key=" + key + "&q=isbn:" + isbn

		fmt.Println(url)
		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		fmt.Println(resp.Body)

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(body, &book)
		if err != nil {
			panic(err)
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

	// MEMO: メモ書き
	// file, _ := os.Open(image)
	// defer file.Close()

	// fi, _ := file.Stat() //FileInfo interface
	// size := fi.Size()    //ファイルサイズ

	// data := make([]byte, size)
	// file.Read(data)

	return base64.StdEncoding.EncodeToString(image)
}

func (book *Book) UnmarshalJSON(byte []byte) error {
	type GoogleBook struct {
		Items []struct {
			VolumeInfo struct {
				Title string `json:"title"`
				ImageLinks struct {
					Medium string `json:"thumbnail"`
				} `json:"imageLinks"`
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

	return err
}
