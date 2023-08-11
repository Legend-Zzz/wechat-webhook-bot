FROM golang:1.21 AS build

WORKDIR /app
COPY . .

RUN go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,https://goproxy.io,direct \
    && go mod tidy \
    && GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w" -o main

FROM alpine:latest

WORKDIR /app
COPY --from=build /app/main .
COPY templates templates

EXPOSE 8000

CMD ["./main"]