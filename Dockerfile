# Build stage
FROM alpine:3.20.2 AS builder
COPY . .
RUN apk add --no-cache upx go
RUN go build -ldflags "-s -w"
RUN upx --best --lzma deadsniper

# Run stage
FROM alpine:3.20.2
COPY --from=builder /deadsniper /deadsniper
ENTRYPOINT ["/deadsniper"]