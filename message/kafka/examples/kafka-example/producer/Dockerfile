FROM golang:1.23.4-bookworm

WORKDIR app

COPY . .

RUN go build -o app .

ENTRYPOINT ["./app"]