FROM  quay.io/wasilak/golang:1.15-alpine as builder

WORKDIR /go/src/github.com/wasilak/cloudpile/

COPY ./src .

RUN go get github.com/GeertJohan/go.rice/rice

RUN rice embed-go && go build .

FROM quay.io/wasilak/alpine:3

COPY --from=builder /go/src/github.com/wasilak/cloudpile/cloudpile /cloudpile

CMD ["/cloudpile"]
