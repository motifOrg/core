# ---------- Go Builder ----------
FROM golang:1.26 AS backend
WORKDIR /app/core

# Install
COPY go.mod go.sum ./
RUN go mod download

# Copy
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /server ./cmd/server

# ---------- Runtime ----------
FROM gcr.io/distroless/static-debian12

WORKDIR /

COPY --from=backend /server /server
COPY --from=backend /app/core/internal/migrations /internal/migrations

EXPOSE 8000
ENTRYPOINT ["/server"]
