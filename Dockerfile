FROM  quay.io/wasilak/golang:1.15-alpine as builder

WORKDIR /go/src/github.com/wasilak/cloudpile/

COPY ./src .

RUN go get github.com/GeertJohan/go.rice/rice

RUN rice embed-go && go build .

FROM quay.io/wasilak/alpine:3

COPY --from=builder /go/src/github.com/wasilak/cloudpile/cloudpile /usr/local/bin/cloudpile
COPY --from=builder /go/src/github.com/wasilak/cloudpile/cloudpile_example.yml /etc/cloudpile/cloudpile.yml

CMD ["/usr/local/bin/cloudpile"]
