FROM golang:1.15.5-alpine3.12 as builder

# Set workdir
WORKDIR /sources

# Add source files
COPY . .

# Install minimum necessary dependencies
RUN apk add --no-cache make gcc libc-dev

RUN make build

# ----------------------------

FROM alpine:3.12

COPY --from=builder /sources/build/ /usr/local/bin/

RUN addgroup tendermint && adduser -S -G tendermint tendermint -h /data

USER tendermint

WORKDIR /data

EXPOSE 26656

ENTRYPOINT ["tenderseed", "start"]
