ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG CLOUDPILE_VERSION

FROM --platform=${BUILDPLATFORM} quay.io/wasilak/alpine:3

ARG GOOS
ARG GOARCH

ADD https://github.com/wasilak/cloudpile/releases/download/${CLOUDPILE_VERSION}/cloudpile-${GOOS}-${GOARCH}.zip /cloudpile

CMD ["/cloudpile"]
