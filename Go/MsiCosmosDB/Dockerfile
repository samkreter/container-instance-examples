FROM golang:1.9.2 as builder
WORKDIR  /go/src/workdir/
COPY . /go/src/workdir/
#RUN go test ./... -v
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o run .

FROM alpine:3.8
RUN apk --update add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/workdir/run .
COPY --from=builder /go/src/workdir/index.html .
CMD ["./run"]
