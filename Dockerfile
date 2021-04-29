ARG TARGETPLATFORM
ARG BUILDPLATFORM

FROM --platform=${BUILDPLATFORM} quay.io/wasilak/alpine:3

ARG GOOS
ARG GOARCH

ADD ./dist/cloudpile-$GOOS-$GOARCH /cloudpile

CMD ["/cloudpile"]
