FROM golang:1.23-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN apk add --no-cache gcc musl-dev 

RUN go build -tags musl 

CMD ["./stream_protobuf_example"]
