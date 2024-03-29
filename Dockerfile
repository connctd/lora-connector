FROM alpine as alpine

RUN apk add -U --no-cache ca-certificates

FROM scratch

WORKDIR /
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ADD ./lora-connector_linux_amd64 /service
ENTRYPOINT ["/service"]
