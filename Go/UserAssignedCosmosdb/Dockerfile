FROM golang:1.9.2 as builder

RUN apt-get update && apt-get install -y unzip --no-install-recommends && \
    apt-get autoremove -y && apt-get clean -y && apt-get install -y curl

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

WORKDIR  /go/src/workdir/

COPY Gopkg.toml Gopkg.lock ./

RUN dep ensure -vendor-only

COPY . /go/src/workdir/
#RUN go test ./... -v
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o run .

FROM alpine:3.8
RUN apk --update add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/workdir/run .
COPY --from=builder /go/src/workdir/index.html .
CMD ["./run"]
