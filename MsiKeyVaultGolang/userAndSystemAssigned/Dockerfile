FROM alpine:3.8
RUN apk --update add ca-certificates
WORKDIR /root/
COPY ./msi-golang ./msi-golang
CMD ["./msi-golang"]
