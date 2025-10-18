FROM docker.io/golang:1.25 AS builder
RUN mkdir /app
WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o piwatcher .

FROM docker.io/alpine:latest
RUN mkdir /app && adduser -h /app -D piwatcher
WORKDIR /app
COPY --chown=piwatcher --from=builder /app/piwatcher .
ENTRYPOINT ["/app/piwatcher"]
