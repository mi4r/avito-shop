FROM golang:1.23.1

WORKDIR ${GOPATH}/avito-shop/
COPY . ${GOPATH}/avito-shop/

RUN go build -o /build ./cmd/shop/main.go 
RUN go clean -cache -modcache
EXPOSE 8080

CMD ["/build"]