FROM golang:1.25 AS builder

WORKDIR /app

COPY src/ ./src/
COPY go.mod go.sum ./
COPY templates/ ./templates/
COPY static/ ./static/

RUN ["go", "mod", "download"]
ENV CGO_ENABLED=1
RUN ["go", "build", "-o", "/out/main", "./src/main.go"]


FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app -h /app app \
    && mkdir -p /db \
    && chown app:app /db

COPY --from=builder /out/main .
COPY --from=builder /app/templates ./templates/
COPY --from=builder /app/static ./static/

EXPOSE 8080

USER app

CMD ["./main"]