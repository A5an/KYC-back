FROM golang:1.21-alpine AS build
WORKDIR /platform

RUN apk update && apk add git make
RUN apk add upx
COPY . .

RUN go mod download

RUN go build -tags netgo -ldflags '-s -w' -o dist/core cmd/main.go

FROM alpine

RUN apk --no-cache add ca-certificates

WORKDIR /app/

COPY --from=build /platform/.config /app/.config
COPY --from=build /platform/dist/core /app/core

EXPOSE 5435

RUN ls .

ENTRYPOINT ["/app/core"]