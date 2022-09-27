### builder ###
FROM golang:1.19 as builder
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags osusergo,netgo -a -o seaman .

### runner ###
FROM alpine:3.16.2
WORKDIR /
RUN apk add -u git
COPY --from=builder /workspace/seaman .
USER 65532:65532

ENTRYPOINT ["/seaman"]
