FROM alpine:latest

ADD admission-webhook-security /admission-webhook-security
ENTRYPOINT ["./admission-webhook-security"]