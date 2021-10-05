# api-web-scraper

This is a Golang version of a web scraper. 

It makes use of a number of frameworks such the [gin](https://github.com/gin-gonic/gin) web framework with [redis](https://redis.io/) as a data store and [colly](https://github.com/gocolly/colly) for scraping.

Colly has some great features that make it suitable for fast, ethical scraping.

* https://github.com/gin-gonic/gin
* 

## Build and run

1. On the repo root directory run:

```bash
docker-compose up -d redis
go build
./api-web-scraper
```
2. To view page contents of a page:

  * on the terminal run: `curl http://localhost:8080/scraper/page-contents\?url\=https://www.example.com` or

  * on your browser or postman paste: `http://localhost:8080/scraper/page-contents?url=https://www.example.com`

3. update the `url` parameter to view the page contents of other URLs.

### Code Formatting

Format entire project in GO standard format.

```bash
go fmt $(go list ./... | grep -v /vendor/)
```

### Running tests:

Gingko is used for testing. To install ginkgo: `go install github.com/onsi/ginkgo/ginkgo@latest to install ginkgo`

To run tests: `go test ./...`

### Linting

TODO. Intention to use golangci.

### Assumptions

* no standard for login forms, so checked whether just one password field was present on a page. Two fields of type password are sometimes used for sign-up pages. Also worth noting is some pages have both sign-up and login pages on the same page.
* web application in my case is an API with a json response.

### Considerations and Improvements

* more tests in the various application layers, and testing of different code paths and response codes
* addition of linting, and fixing of any found issues
* exploring some caching - for example colly has some caching mechanisms, and other useful enhancements
* swagger file for contract