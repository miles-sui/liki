FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY vendor/ ./vendor/
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -mod=vendor -o /mingli-server ./cmd/mingli-server/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates curl
WORKDIR /app
COPY --from=build /mingli-server /usr/local/bin/mingli-server
EXPOSE 8080
USER 65534:65534
ENTRYPOINT ["mingli-server", "--addr", ":8080"]
