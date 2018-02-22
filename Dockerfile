FROM golang:1.10-alpine AS build

RUN mkdir -p /go/src/github.com/benjdewan/pachelbel
WORKDIR /go/src/github.com/benjdewan/pachelbel
COPY . /go/src/github.com/benjdewan/pachelbel

RUN apk add --no-cache \
    make \
    git

RUN make

FROM alpine:latest as pachelbel
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /go/src/github.com/benjdewan/pachelbel/pachelbel-linux /pachelbel
ENTRYPOINT [ "/pachelbel" ]
CMD [ "--help" ]

