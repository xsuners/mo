# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:alpine AS builder
WORKDIR /workspace
ARG TARGETOS
ARG TARGETARCH
RUN apk --no-cache add wget && \
    PROTOBUF_VERSION=25.2 \
    PROTOBUF_DOWNLOAD_URL=https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOBUF_VERSION} && \
    case "${TARGETOS}" in \
    "linux") \
    case "${TARGETARCH}" in \
    "386") \
    wget ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-linux-x86_32.zip && \
    unzip protoc-${PROTOBUF_VERSION}-linux-x86_32.zip ;; \
    "amd64") \
    wget ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-linux-x86_64.zip && \
    unzip protoc-${PROTOBUF_VERSION}-linux-x86_64.zip ;; \
    "arm/v5" | "arm/v7") \
    wget ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-linux-arm_32.zip && \
    unzip protoc-${PROTOBUF_VERSION}-linux-arm_32.zip ;; \
    "arm64") \
    wget ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-linux-aarch_64.zip && \
    unzip protoc-${PROTOBUF_VERSION}-linux-aarch_64.zip ;; \
    "mips64le") \
    wget ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-linux-mips64le.zip && \
    unzip protoc-${PROTOBUF_VERSION}-linux-mips64le.zip ;; \
    "ppc64le") \
    wget ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-linux-ppcle_64.zip && \
    unzip protoc-${PROTOBUF_VERSION}-linux-ppcle_64.zip ;; \
    "s390x") \
    wget ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-linux-s390_64.zip && \
    unzip protoc-${PROTOBUF_VERSION}-linux-s390_64.zip ;; \
    *) echo "Unsupported architecture" && exit 1 ;; \
    esac ;; \
    "darwin") \
    case "${TARGETARCH}" in \
    "amd64") \
    wget ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-osx-x86_64.zip && \
    unzip protoc-${PROTOBUF_VERSION}-osx-x86_64.zip ;; \
    "arm64") \
    wget ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-osx-aarch_64.zip && \
    unzip protoc-${PROTOBUF_VERSION}-osx-aarch_64.zip ;; \
    *) echo "Unsupported architecture" && exit 1 ;; \
    esac ;; \
    "windows") \
    case "${TARGETARCH}" in \
    "386") \
    wget ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-win32.zip && \
    unzip protoc-${PROTOBUF_VERSION}-win32.zip ;; \
    "amd64") \
    wget ${PROTOBUF_DOWNLOAD_URL}/protoc-${PROTOBUF_VERSION}-win64.zip && \
    unzip protoc-${PROTOBUF_VERSION}-win64.zip ;; \
    *) echo "Unsupported architecture" && exit 1 ;; \
    esac ;; \
    *) echo "Unsupported operating system" && exit 1 ;; \
    esac && \
    \
    rm *.zip && \
    chmod +x bin/*


RUN go install github.com/golang/protobuf/protoc-gen-go@v1.5.3 && \
    go install github.com/xsuners/mo/cmd/protoc-gen-go-mo@v0.1.5 && \
    go install github.com/xsuners/mo/cmd/mo@v0.1.5


FROM --platform=$BUILDPLATFORM alpine:latest
COPY --from=builder /workspace/bin/* /bin/
COPY --from=builder /workspace/include/google /usr/local/include/google
COPY --from=builder /go/bin/mo /bin/
COPY --from=builder /go/bin/protoc-gen-go-mo /bin/
COPY --from=builder /go/bin/protoc-gen-go /bin/

