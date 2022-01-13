FROM golang:alpine as builder

RUN apk add --no-cache make git
WORKDIR /proxypool-src
COPY . /proxypool-src
#ENV GOPROXY https://goproxy.cn
RUN make docker && \
    mv bin/proxypool-docker /proxypool

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
#COPY config/assets /app/assets
COPY ./config/config.yaml /app/config/
COPY ./config/source.yaml /app/config/
COPY --from=builder /proxypool /app/

ENV TZ Asia/Shanghai
EXPOSE 12580
ENTRYPOINT ["/app/proxypool", "-d", "-c", "config/config.yaml"]
