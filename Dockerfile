FROM golang:1.22.3-alpine3.20

WORKDIR /app

COPY ./ ./
RUN go mod download

# build the go app binary.
RUN GOOS=linux GOARCH=amd64 go build -mod=readonly -v -o server

EXPOSE 8080

# command to run the app.
CMD ./server