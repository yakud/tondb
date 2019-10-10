GOOS=linux GOARCH=amd64 go build -o ./blocks-stream-receiver ./cmd/blocks-stream-receiver/ && \
  scp ./blocks-stream-receiver root@144.76.140.152:/tmp/blocks-stream-receiver && \
  rm ./blocks-stream-receiver

GOOS=linux GOARCH=amd64 go build -o ./ton-api ./cmd/ton-api/ && \
  scp ./ton-api root@144.76.140.152:/tmp/ton-api && \
  rm ./ton-api

#  scp ./blocks-stream-receiver akisilev@46.4.4.150:/tmp/blocks-stream-receiver && \
