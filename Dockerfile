FROM golang:alpine

RUN apk add --no-cache git gcc musl-dev

RUN CGO_ENABLED=1 go get --tags "linux sqlite_stat4 sqlite_allow_uri_authority sqlite_fts5 sqlite_introspect sqlite_json" github.com/alash3al/sqler

ENTRYPOINT ["sqler"]

WORKDIR /root/