# bookshelf-generator

- It can generate a svg of bookshelf that put on specified book.
- You can display it on any pages composed of markdown.

## Stack

- Golang
- Gin

## Getting started

```shell
go mod tidy
go run main.go
```

### env

- GOOGLE_BOOKS_API_KEY: Google Books API key

## Usage

- You need to specify isbn code of any books as `isbns[]` query parameter.
- This product can generate up to 5 items.

## Example

- shown images of specified books as below

![My bookshelf](https://bookshelf-generator.onrender.com?isbns[]=9784798178189&isbns[]=9784774189673&isbns[]=9784274226298)
