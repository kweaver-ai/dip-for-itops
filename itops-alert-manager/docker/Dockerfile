########################### Stage 0 ########################
ARG BUILD_IMAGE=acr.aishu.cn/ar/go.build:1.24.11-20251203
ARG BASE_IMAGE=acr.aishu.cn/public/ubuntu:22.04.20240430

FROM ${BUILD_IMAGE} AS builder
ARG GO_PROXY=http://10.4.30.214:18005,http://repository.aishu.cn:8081/repository/go-all,direct
COPY ./server /go/src
COPY .netrc /root/.netrc
COPY ./id_rsa /root/.ssh/id_rsa

RUN set -ex; \
    cd /go/src; \
    export GOPROXY=$GO_PROXY; \
    go mod download -x; \
    CGO_ENABLED=0 go build -ldflags "-s -w" -o ./bin/itops-alert-manager-server;

########################### Stage 1 ########################
# Copy working directory to the actual release docker images
FROM ${BASE_IMAGE} AS prod

COPY --from=builder  /go/src/bin /opt/itops-alert-manager

ADD server/config/config.yaml /opt/itops-alert-manager/config/
ADD server/locale /opt/itops-alert-manager/locale


# 指定工作目录
WORKDIR /opt/itops-alert-manager/

RUN set -ex; \
    cd /opt/itops-alert-manager/config; \
    cat config.yaml; \
    cd /opt/itops-alert-manager/; \
    ls;

# Expose port
EXPOSE 13046


# 指定启动命令
ENTRYPOINT [ "./itops-alert-manager-server" ]

