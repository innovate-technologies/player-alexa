ARG ARCH
FROM golang as build 

COPY ./ /go/src/github.com/innovate-technologies/player-alexa
WORKDIR /go/src/github.com/innovate-technologies/player-alexa

ARG GO_ARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH="${GO_ARCH}" go build -a -installsuffix cgo ./

ARG ARCH
FROM multiarch/alpine:${ARCH}-edge

RUN apk add --no-cache ca-certificates

COPY --from=build /go/src/github.com/innovate-technologies/player-alexa/player-alexa /usr/local/bin/player-alexa

CMD [ "player-alexa" ]