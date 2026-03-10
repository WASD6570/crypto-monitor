FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod ./
COPY apps ./apps
COPY cmd ./cmd
COPY configs ./configs
COPY libs ./libs
COPY schemas ./schemas
COPY services ./services
COPY tests ./tests

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/market-state-api ./cmd/market-state-api

FROM alpine:3.22

RUN apk add --no-cache wget

WORKDIR /app

COPY --from=build /bin/market-state-api /usr/local/bin/market-state-api
COPY --from=build /src/configs /app/configs

EXPOSE 8080

CMD ["/usr/local/bin/market-state-api"]
