FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY server-linux ./server
COPY migrations ./migrations

EXPOSE 8080

CMD ["./server"]
