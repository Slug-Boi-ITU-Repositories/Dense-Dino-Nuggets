FROM golang:1.25 AS builder

WORKDIR /app

COPY src/ ./src/
COPY go.mod go.sum ./
COPY templates/ ./templates/
COPY static/ ./static/

RUN ["go", "mod", "download"]
ENV CGO_ENABLED=1
RUN ["go", "build", "-o", "/out/main", "./src/main.go"]


FROM debian:bookworm-slim

WORKDIR /app

RUN groupadd -r app && useradd -r -g app -d /app -s /usr/sbin/nologin app \
    && mkdir -p /db \
    && chown app:app /db

COPY --from=builder /out/main .
COPY --from=builder /app/templates ./templates/
COPY --from=builder /app/static ./static/

EXPOSE 8080

USER app

CMD ["./main"]