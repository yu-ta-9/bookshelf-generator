# bookshelf-generator

- It can generate a svg of bookshelf that put on specified book.
- You can display it on any pages composed of markdown.

## Stack

- Golang
- Gin

## Getting started

```shell
go mod download
docker compose up
go run main.go

# optional commands
go mod tidy
```

### env

- GOOGLE_BOOKS_API_KEY: Google Books API key
- REDIS_URL: Url of redis

## Usage

- You need to specify isbn-13 code of any books as `isbns[]` query parameter.
- This product can generate up to 5 items.

## Example

- shown images of specified books as below

![My bookshelf having a book](https://bookshelf.yu-ta-9.com?isbns[]=9784798178189)

![My bookshelf having two books](https://bookshelf.yu-ta-9.com?isbns[]=9784798178189&isbns[]=9784774189673)

![My bookshelf having three books](https://bookshelf.yu-ta-9.com?isbns[]=9784798178189&isbns[]=9784774189673&isbns[]=9784274226298)

![My bookshelf having four books](https://bookshelf.yu-ta-9.com?isbns[]=9784798178189&isbns[]=9784774189673&isbns[]=9784274226298&isbns[]=9784873115894)

![My bookshelf having five books](https://bookshelf.yu-ta-9.com?isbns[]=9784798178189&isbns[]=9784774189673&isbns[]=9784274226298&isbns[]=9784873115894&isbns[]=9784873115658)

## Note

- This application complies with the [terms](https://developers.google.com/books/branding) of google books api.
