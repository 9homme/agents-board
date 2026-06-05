# Multi-stage Go build.
# Stage 1 — build both api-server and mcp-server binaries.
# Stage 2 — minimal runtime image (distroless static).
#
# Architecture ref: REQ005 §6 (US008 scope), D-012.
#
# Default CMD starts api-server. Override via compose `command:` to run mcp-server.

FROM golang:1.25-alpine AS build

WORKDIR /work

# Download dependencies before copying source (cache-friendly layer ordering).
COPY go.mod go.sum ./
RUN go mod download

# Copy full source and compile both binaries.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api-server ./cmd/api-server
RUN CGO_ENABLED=0 GOOS=linux go build -o mcp-server ./cmd/mcp-server

# Stage 2 — distroless static runtime for minimal attack surface.
FROM gcr.io/distroless/static-debian12 AS runtime

WORKDIR /

COPY --from=build /work/api-server .
COPY --from=build /work/mcp-server .

# Run as non-root (distroless static ships with `nonroot`, UID 65532).
USER nonroot:nonroot

# api-server is the default entry; compose may override with ["./mcp-server"].
CMD ["./api-server"]
