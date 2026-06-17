FROM alpine:3.21 AS certs
RUN apk add --no-cache ca-certificates
RUN adduser -D -u 1000 lingji

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=certs /etc/passwd /etc/passwd
COPY bin/lingji /usr/local/bin/lingji
USER lingji
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/lingji", "--addr", ":8080"]
