### builder ###
FROM golang:1.19 as builder
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags osusergo,netgo -a -o chatbot .

### runner ###
FROM alpine:3.13.6
WORKDIR /
RUN apk add -u git
COPY --from=builder /workspace/chatbot .
USER 65532:65532

ENTRYPOINT ["/chatbot"]
