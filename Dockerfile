# Build stage
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

RUN apk add --no-cache make git bash

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build-binary \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    VERSION=${VERSION} \
    COMMIT=${COMMIT} \
    BUILD_DATE=${BUILD_DATE}

# Runtime stage
FROM alpine:3.19
RUN apk add --no-cache ca-certificates bash curl
COPY --from=builder /src/bin/lissto /usr/local/bin/lissto
ENTRYPOINT ["lissto"]
