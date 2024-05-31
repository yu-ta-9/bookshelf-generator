build:
	docker build --no-cache --tag bookshelf-generator:latest .

run:
	docker run -it --rm -v $(PWD)/output:/output -p 8080:8080 bookshelf-generator:latest
