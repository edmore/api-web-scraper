# syntax=docker/dockerfile:1

##
## Build
##
FROM golang:1.17-buster AS build

RUN apt-get update && apt-get install -y ca-certificates openssl
ARG cert_location=/usr/local/share/ca-certificates

# Get certificate from "github.com"
RUN openssl s_client -showcerts -connect github.com:443 </dev/null 2>/dev/null|openssl x509 -outform PEM > ${cert_location}/github.crt
# Get certificate from "proxy.golang.org"
#RUN openssl s_client -showcerts -connect proxy.golang.org:443 </dev/null 2>/dev/null|openssl x509 -outform PEM >  ${cert_location}/proxy.golang.crt
# Get certificate from "golang.org"
RUN openssl s_client -showcerts -connect golang.org:443 </dev/null 2>/dev/null|openssl x509 -outform PEM >  ${cert_location}/golang.crt
# Get certificate from "sum.golang.org"
RUN openssl s_client -showcerts -connect sum.golang.org:443 </dev/null 2>/dev/null|openssl x509 -outform PEM >  ${cert_location}/sum.golang.org.crt
# Update certificates
RUN update-ca-certificates

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /api-web-scraper .

##
## Deploy
##
# FROM alpine:latest
# RUN apk --no-cache add ca-certificates
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /api-web-scraper /api-web-scraper

EXPOSE 8080

#USER nonroot:nonroot

ENTRYPOINT ["/api-web-scraper"]