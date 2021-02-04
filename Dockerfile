FROM golang:1.12 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build

FROM alpine

RUN apk --no-cache add ca-certificates

WORKDIR /erd
COPY --from=builder /app/erd-go .

ENTRYPOINT ["/erd/erd-go"]
