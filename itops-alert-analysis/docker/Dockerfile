########################### Stage 0 ########################
ARG BUILD_IMAGE=acr.aishu.cn/ar/go.build:1.24.11-20251203
ARG BASE_IMAGE=acr.aishu.cn/public/ubuntu:22.04.20251014

FROM ${BUILD_IMAGE} AS builder

COPY ./server /go/src

COPY .netrc /root/.netrc
COPY id_rsa /root/.ssh/id_rsa

ARG SERVER_VERSION=5.0.3.4
ARG GOPROXY_URL="http://goproxy.cn,direct"

ENV GOPROXY=${GOPROXY_URL}

RUN set -ex; \
    cd /go/src; \
    go mod download -x; \
    go build -ldflags "-s -w" -o ./bin/itops-alert-analysis-server;

########################### Stage 1 ########################
# Copy working directory to the actual release docker images
FROM ${BASE_IMAGE} AS prod

ARG UID=1000
ARG GID=1000

RUN groupadd -r -g $GID anyrobot \
    && useradd -r -u $UID -g $GID anyrobot


COPY --from=builder --chown=$UID:$GID /go/src/bin /opt/itops-alert-analysis

ADD server/config/config.yaml /opt/itops-alert-analysis/


# 指定工作目录
WORKDIR /opt/itops-alert-analysis/

# Change user
USER $UID

# Expose port
EXPOSE 80

# 指定启动命令
ENTRYPOINT [ "./itops-alert-analysis-server" ]
