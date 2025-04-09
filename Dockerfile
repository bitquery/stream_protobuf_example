FROM golang:1.23-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN apk add --no-cache gcc musl-dev 

RUN go build -tags musl 


FROM alpine:3.21 AS runner

WORKDIR /app

ENV PATH=${PATH}:/app/bin

COPY --from=builder /app/ssl /app/ssl
COPY --from=builder /app/stream_protobuf_example /app/bin/stream_protobuf_example

CMD ["./bin/stream_protobuf_example"]
