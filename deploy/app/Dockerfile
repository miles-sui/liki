FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY vendor/ ./vendor/
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -mod=vendor -o /25types-server ./cmd/app-server/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates curl
WORKDIR /app
COPY --from=build /25types-server /usr/local/bin/25types-server
RUN mkdir -p /var/lib/25types
EXPOSE 8080
USER 65534:65534
ENTRYPOINT ["25types-server", "--addr", ":8080", "--db", "/var/lib/25types/25types.db"]
