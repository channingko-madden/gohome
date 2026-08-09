FROM docker.io/golang:1.26 AS builder
RUN mkdir /app
WORKDIR /app
COPY . /app
# Pi
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w"
# x86
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w"

FROM docker.io/alpine:latest
RUN mkdir /app && adduser -h /app -D picoexport
WORKDIR /app
COPY --chown=picoexport --from=builder /app/picoexport .
CMD ["/app/picoexport"]
