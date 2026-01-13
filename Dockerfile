FROM alpine:3.19
RUN apk add --no-cache ca-certificates bash curl
COPY lissto /usr/local/bin/lissto
RUN chmod +x /usr/local/bin/lissto
ENTRYPOINT ["lissto"]
