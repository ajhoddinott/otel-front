# Stage 1: Build frontend
FROM node:22-alpine@sha256:8094c002d08262dba12645a3b4a15cd6cd627d30bc782f53229a2ec13ee22a00 AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.24-bookworm@sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/internal/server/static/ ./internal/server/static/
RUN CGO_ENABLED=1 go build -trimpath \
    -ldflags="-s -w" \
    -o otel-front \
    ./cmd/viewer/main.go

# Stage 3: Distroless runtime (includes glibc + libstdc++ required by DuckDB/CGO)
FROM gcr.io/distroless/cc-debian12
COPY --from=go-builder /app/otel-front /usr/local/bin/otel-front
EXPOSE 8000 4317 4318
ENTRYPOINT ["/usr/local/bin/otel-front"]
CMD ["--no-browser"]
