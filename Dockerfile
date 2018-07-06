FROM golang as builder

RUN go get github.com/kaishuu0123/erd-go

FROM alpine

WORKDIR /erd
COPY --from=builder /go/bin/erd-go .

ENTRYPOINT ["/erd/erd-go"]
