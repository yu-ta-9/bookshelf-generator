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

		books, err := getBookData(isbns)
		if err != nil {
			fmt.Printf("error fetching book data, %v", err)
			return
		}

		c.Header("Content-Type", "image/svg+xml")
		c.HTML(http.StatusOK, "bookshelf.html", gin.H{
			"books": books,
			"bookLength": len(books),
		})
	})

	r.Run(":8088")
}

type Book struct {
	Title       string
	ImageBase64 string
}

// getBookData fetches book data
func getBookData(isbns []string) ([]Book, error) {
	key := os.Getenv("RAKUTEN_BOOKS_API_KEY")

	// url := "https://api.openbd.jp/v1/get?isbn=9784798178639"
	// url := "https://app.rakuten.co.jp/services/api/BooksBook/Search/20170404?format=json&outOfStockFlag=1" + "&applicationId=" + key + "&isbn=" + "9784798178189" // + "&isbn[]=" + "9784774189673"

	var books []Book

	for _, isbn := range isbns {
		var book Book

		url := "https://app.rakuten.co.jp/services/api/BooksBook/Search/20170404?format=json&outOfStockFlag=1" + "&applicationId=" + key + "&isbn=" + isbn

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
	type RakutenBook struct {
		Items []struct {
			Item struct {
				Title          string `json:"title"`
				MediumImageUrl string `json:"mediumImageUrl"`
			}
		} `json:"items"`
	}
	var rakutenBook RakutenBook
	err := json.Unmarshal(byte, &rakutenBook)
	if err != nil {
		fmt.Println(err)
	}

	if len(rakutenBook.Items) == 0 {
		return nil
	}

	book.Title = rakutenBook.Items[0].Item.Title
	book.ImageBase64 = encode(rakutenBook.Items[0].Item.MediumImageUrl)

	return err
}
