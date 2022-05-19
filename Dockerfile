FROM golang:1.18-alpine as builder
WORKDIR /src

RUN apk add --no-cache --update \
  git  \
    ca-certificates

COPY . /src
RUN go mod download
RUN go build -o /src

FROM alpine as secondBuilder
COPY --from=builder /src/ /app
WORKDIR /app
EXPOSE 8080
ENTRYPOINT /app/worth2watch2
