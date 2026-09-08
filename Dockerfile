FROM golang:1.27-alpine AS builder

WORKDIR /app
RUN apk add --no-cache gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o lanfile ./cmd/lanfile

FROM alpine:3
WORKDIR /app


RUN apk add --no-cache ca-certificates

RUN adduser -D lanfileuser

COPY --from=builder /app/lanfile /app/lanfile
RUN mkdir -p /app/data \
    && chown -R lanfileuser:lanfileuser /app
USER lanfileuser
CMD ["./lanfile"]