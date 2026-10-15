# syntax=docker/dockerfile:1.7

# ---- build stage ----
FROM golang:1.22.5-alpine3.20 AS builder

WORKDIR /src

# Install dependencies first so the layer caches across source changes.
COPY go.mod go.sum ./
RUN go mod download

# Now copy the source.
COPY . .

# Static build: no CGO so the binary runs on a minimal runtime image.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/app .

# ---- runtime stage ----
FROM alpine:3.20

# psql client so the same image can serve as the migration runner.
RUN apk add --no-cache postgresql16-client ca-certificates tzdata

# Non-root user.
RUN addgroup -S app -g 10001 && adduser -S -u 10001 -G app app

WORKDIR /app
COPY --from=builder /out/app /app/app
COPY migrations /app/migrations

USER app
EXPOSE 8080

ENV PORT=8080
ENTRYPOINT ["/app/app"]