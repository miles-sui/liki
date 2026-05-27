# Custom Caddy with Cloudflare DNS-01 challenge support + static frontend

FROM caddy:2-builder AS caddy-builder
ENV GOPROXY=https://goproxy.cn,direct \
    GONOSUMCHECK=* \
    GONOSUMDB=*
RUN xcaddy build \
    --with github.com/caddy-dns/cloudflare \
    --with github.com/mholt/caddy-ratelimit

FROM caddy:2-alpine
COPY --from=caddy-builder /usr/bin/caddy /usr/bin/caddy
COPY web/dist/ /app/frontend/dist/
