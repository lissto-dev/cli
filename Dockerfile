# Build stage
FROM golang:1.24-alpine AS builder

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w \
      -X github.com/lissto-dev/cli/cmd.Version=${VERSION} \
      -X github.com/lissto-dev/cli/cmd.Commit=${COMMIT} \
      -X github.com/lissto-dev/cli/cmd.Date=${BUILD_DATE}" \
    -o /lissto .

# Runtime stage
FROM alpine:3.19
RUN apk add --no-cache ca-certificates bash curl
COPY --from=builder /lissto /usr/local/bin/lissto
ENTRYPOINT ["lissto"]
