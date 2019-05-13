FROM golang:1.12-alpine AS builder

ARG promu_version=0.4.0

RUN apk --no-cache add git \
    && wget -qO- https://github.com/prometheus/promu/releases/download/v${promu_version}/promu-${promu_version}.linux-amd64.tar.gz | tar -xzvf - -C /opt \
    && ln -s /opt/promu-0.4.0.linux-amd64/promu /usr/local/bin/promu

WORKDIR /src/burrow_exporter
COPY . .
RUN promu build

FROM busybox
LABEL maintainer "Alex Simenduev <shamil.si@gmail.com>"

COPY --from=builder /src/burrow_exporter/burrow_exporter /usr/local/bin/
ENTRYPOINT ["burrow_exporter"]