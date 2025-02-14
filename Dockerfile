FROM quay.io/wasilak/golang:1.24-alpine as builder

COPY . /app
WORKDIR /app/
RUN mkdir -p ../dist
RUN go build -o /cloudpile

FROM quay.io/wasilak/alpine:3

COPY --from=builder /cloudpile /cloudpile

CMD ["/cloudpile"]
