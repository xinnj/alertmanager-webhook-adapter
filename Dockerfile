FROM --platform=$BUILDPLATFORM golang:1.26 AS builder

ARG TARGETOS
ARG TARGETARCH

COPY / /src

RUN cd /src \
  && CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH make build \
  && mv /src/_output/alertmanager-webhook-adapter /alertmanager-webhook-adapter

FROM golang:1.26

COPY --from=builder /alertmanager-webhook-adapter /alertmanager-webhook-adapter

ENTRYPOINT [ "/alertmanager-webhook-adapter" ]
CMD ["--debug"]
