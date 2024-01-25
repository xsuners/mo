##  MO

```
   ____ ___   ____
  / __ `__ \ / __ \
 / / / / / // /_/ /
/_/ /_/ /_/ \____/

```


我要创建一个 protobuf docker 镜像。要支持跨平台

根据 TARGETOS, TARGETARCH 下载对应的protobuf 并解压安装

其中 TARGETOS 取值范围 linux，darwin，windows。
TARGETARCH 取值范围 386,amd64,arm/v5,arm/v7,arm64,mips64le,ppc64le,s390x


protoc 支持的平台如下，注意做好对应：

protoc-25.2-linux-aarch_64.zip
protoc-25.2-linux-ppcle_64.zip
protoc-25.2-linux-s390_64.zip
protoc-25.2-linux-x86_32.zip
protoc-25.2-linux-x86_64.zip
protoc-25.2-osx-aarch_64.zip
protoc-25.2-osx-universal_binary.zip
protoc-25.2-osx-x86_64.zip
protoc-25.2-win32.zip
protoc-25.2-win64.zip

下载地址为

https://github.com/protocolbuffers/protobuf/releases/download/v25.2/protoc-25.2-linux-ppcle_64.zip


请帮我写一个dockerfile



```Dockerfile
# 第一阶段: 构建Protobuf
FROM alpine:latest AS builder

# 设置环境变量
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ENV TARGETOS=${TARGETOS} \
    TARGETARCH=${TARGETARCH}

# 设置工作目录
WORKDIR /protobuf

# 安装依赖工具
RUN apk --no-cache add wget && \
    \
    # 定义 Protobuf 版本和下载地址
    PROTOBUF_VERSION=25.2 \
    PROTOBUF_DOWNLOAD_URL=https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOBUF_VERSION} && \
    PROTOBUF_NAME="" && \
    \
    # 下载并解压对应平台的 Protobuf 文件
    case "${TARGETOS}" in \
        "linux") \
            case "${TARGETARCH}" in \
                "386") \
                    PROTOBUF_DOWNLOAD_URL = ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-linux-x86_32.zip && \
                    PROTOBUF_NAME=protoc-${PROTOBUF_VERSION}-linux-x86_32.zip;; \
                "amd64") \
                    PROTOBUF_DOWNLOAD_URL = ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-linux-x86_64.zip && \
                    PROTOBUF_NAME=protoc-${PROTOBUF_VERSION}-linux-x86_64.zip;; \
                "arm/v5" | "arm/v7") \
                    PROTOBUF_DOWNLOAD_URL = ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-linux-arm_32.zip && \
                    PROTOBUF_NAME=protoc-${PROTOBUF_VERSION}-linux-arm_32.zip;; \
                "arm64") \
                    PROTOBUF_DOWNLOAD_URL = ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-linux-aarch_64.zip && \
                    PROTOBUF_NAME=protoc-${PROTOBUF_VERSION}-linux-aarch_64.zip;; \
                "mips64le") \
                    PROTOBUF_DOWNLOAD_URL = ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-linux-mips64le.zip && \
                    PROTOBUF_NAME=protoc-${PROTOBUF_VERSION}-linux-mips64le.zip;; \
                "ppc64le") \
                    PROTOBUF_DOWNLOAD_URL = ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-linux-ppcle_64.zip && \
                    PROTOBUF_NAME=protoc-${PROTOBUF_VERSION}-linux-ppcle_64.zip;; \
                "s390x") \
                    PROTOBUF_DOWNLOAD_URL = ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-linux-s390_64.zip && \
                    PROTOBUF_NAME=protoc-${PROTOBUF_VERSION}-linux-s390_64.zip;; \
                *) echo "Unsupported architecture" && exit 1 ;; \
            esac ;; \
        "darwin") \
            case "${TARGETARCH}" in \
                "amd64") \
                    PROTOBUF_DOWNLOAD_URL = ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-osx-x86_64.zip && \
                    PROTOBUF_NAME=protoc-${PROTOBUF_VERSION}-osx-x86_64.zip;; \
                "arm64") \
                    PROTOBUF_DOWNLOAD_URL = ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-osx-aarch_64.zip && \
                    PROTOBUF_NAME=protoc-${PROTOBUF_VERSION}-osx-aarch_64.zip;; \
                *) echo "Unsupported architecture" && exit 1 ;; \
            esac ;; \
        "windows") \
            case "${TARGETARCH}" in \
                "386") \
                    PROTOBUF_DOWNLOAD_URL = ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-win32.zip && \
                    PROTOBUF_NAME=protoc-${PROTOBUF_VERSION}-win32.zip;; \
                "amd64") \
                    PROTOBUF_DOWNLOAD_URL = ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-win64.zip && \
                    PROTOBUF_NAME=protoc-${PROTOBUF_VERSION}-win64.zip;; \
                *) echo "Unsupported architecture" && exit 1 ;; \
            esac ;; \
        *) echo "Unsupported operating system" && exit 1 ;; \
    esac && \
    \
    wget ${PROTOBUF_DOWNLOAD_URL}
    unzip ${PROTOBUF_NAME} && \
    rm ${PROTOBUF_NAME}

# 第二阶段: 最终镜像
FROM alpine:latest

# 复制从第一阶段构建的Protobuf二进制文件
COPY --from=builder /protobuf /protobuf

# 打印 Protobuf 版本信息
RUN protoc --version


docker buildx build \
    --builder=container \
    --platform=linux/arm64,darwin/arm64,windows/amd64,linux/amd64 \
    --tag skyasker/protoc \
    --push .