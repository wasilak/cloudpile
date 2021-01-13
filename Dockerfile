FROM  --platform=$BUILDPLATFORM quay.io/wasilak/golang:1.15-alpine as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM

RUN apk add --update --no-cache git

WORKDIR /go/src/github.com/wasilak/cloudpile/

RUN go get github.com/markbates/pkger/cmd/pkger
COPY --from=tonistiigi/xx:golang / /
COPY ./src .

RUN pkger && go build .

FROM --platform=$BUILDPLATFORM quay.io/wasilak/alpine:3

COPY --from=builder /go/src/github.com/wasilak/cloudpile/cloudpile /cloudpile

CMD ["/cloudpile"]
