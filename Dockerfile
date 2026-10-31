FROM golang:1.26-alpine3.24 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY migrations ./migrations
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/personal ./cmd

FROM alpine:3.24

RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/personal ./personal

USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["/app/personal"]
