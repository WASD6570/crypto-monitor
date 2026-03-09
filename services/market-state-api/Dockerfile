FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod ./
COPY apps ./apps
COPY cmd ./cmd
COPY libs ./libs
COPY schemas ./schemas
COPY services ./services
COPY tests ./tests

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/market-state-api ./cmd/market-state-api

FROM alpine:3.22

RUN apk add --no-cache wget

COPY --from=build /bin/market-state-api /usr/local/bin/market-state-api

EXPOSE 8080

CMD ["/usr/local/bin/market-state-api"]
