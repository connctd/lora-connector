FROM scratch
ADD ./lora-connector_linux_amd64 /service
ENTRYPOINT ["/service"]