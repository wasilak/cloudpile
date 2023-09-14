FROM quay.io/wasilak/golang:1.21-alpine as builder

ADD . /app
WORKDIR /app/
RUN mkdir -p ../dist
RUN go build -o /cloudpile

FROM quay.io/wasilak/alpine:3

COPY --from=builder /cloudpile /cloudpile

CMD ["/cloudpile"]
