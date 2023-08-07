FROM golang:1.20-alpine

WORKDIR /usr/src/app/

COPY src/ ./
RUN go mod tidy
RUN go mod download
RUN go build -v -o app cmd/main.go

EXPOSE 8080

CMD ["app"]
