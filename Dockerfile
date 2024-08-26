FROM alpine:3.20.2 AS builder
COPY . .
RUN apk add --no-cache go
RUN go build
ENTRYPOINT ["/deadsniper"]