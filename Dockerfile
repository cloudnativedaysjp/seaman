# syntax=docker/dockerfile:1.4
### builder ###
FROM golang:1.20 as builder

WORKDIR /workspace
# Arguments
ARG APP_VERSION
ARG APP_COMMIT
# Copy the Go Modules
COPY --link go.mod go.mod
COPY --link go.sum go.sum
RUN go mod download
COPY --link . .
# Build
ARG GOOS=linux
ARG GOARCH=amd64
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags "\
  -X github.com/cloudnativedaysjp/seaman/internal/version.Version=${APP_VERSION} \
  -X github.com/cloudnativedaysjp/seaman/internal/version.Commit=${APP_COMMIT} \
  -s -w \
  " -trimpath -tags osusergo,netgo -a -o seaman ./cmd/seaman/

### runner ###
FROM alpine:3.18

LABEL org.opencontainers.image.authors="Shota Kitazawa, Kohei Ota"
LABEL org.opencontainers.image.url="https://github.com/cloudnativedaysjp/seaman"
LABEL org.opencontainers.image.source="https://github.com/cloudnativedaysjp/seaman/blob/main/Dockerfile"
WORKDIR /
RUN apk add -u git
COPY --link --from=builder /workspace/seaman .
USER 65532:65532

ENTRYPOINT ["/seaman"]
