FROM golang:alpine as builder

RUN apk add --no-cache make git
WORKDIR /proxypool-src
COPY . /proxypool-src
RUN go mod download && \
    make docker && \
    mv bin/proxypool-docker /proxypool

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY ./assets /app/assets
COPY ./config/config.yaml /app/config/
COPY ./config/source.yaml /app/config/
COPY --from=builder /proxypool /app/
ENTRYPOINT ["/app/proxypool", "-d"]
