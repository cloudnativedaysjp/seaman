### builder ###
FROM golang:1.19 as builder

WORKDIR /workspace
# Arguments
ARG APP_VERSION
ARG APP_COMMIT
# Copy the Go Modules
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY . .
# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "\
  -X github.com/cloudnativedaysjp/seaman/version.Version=${APP_VERSION} \
  -X github.com/cloudnativedaysjp/seaman/version.Commit=${APP_COMMIT} \
  " -tags osusergo,netgo -a -o seaman .

### runner ###
FROM alpine:3.17.0
WORKDIR /
RUN apk add -u git
COPY --from=builder /workspace/seaman .
USER 65532:65532

ENTRYPOINT ["/seaman"]
