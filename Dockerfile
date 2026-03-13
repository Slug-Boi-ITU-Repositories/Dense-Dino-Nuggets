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

RUN mkdir -p /db

COPY --from=builder /out/main .
COPY --from=builder /app/templates ./templates/
COPY --from=builder /app/static ./static/

EXPOSE 8080

CMD ["./main"]