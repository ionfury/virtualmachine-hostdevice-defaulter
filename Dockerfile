FROM golang:1.21 as builder
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . . 

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o webhook-server cmd/webhook-server/main.go

FROM alpine:latest  
WORKDIR /root/
COPY --from=builder /workspace/webhook-server .
CMD ["./webhook-server"]

