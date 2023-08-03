FROM golang:1.19.11-alpine as builder

ARG GIT_COMMIT
ARG GIT_DATE

ENV GIT_COMMIT=${GIT_COMMIT}
ENV GIT_DATE=${GIT_DATE}

RUN apk add --no-cache make gcc musl-dev linux-headers git

WORKDIR /sether

RUN git clone --depth 1 https://github.com/setherplatform/seth-go-ethereum.git

COPY . sether-node

WORKDIR /sether/sether-node

RUN make

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /sether/sether-node/build/sether /usr/local/bin/

EXPOSE 6060 18545 18546

ENTRYPOINT ["sether", "--datadir=/data", "--testnet", "--tracenode", "--genesis=/data/genesis.g", "--genesis.allowExperimental", "--verbosity=4", "--cache=3600", "--http", "--http.addr=0.0.0.0", "--http.corsdomain=*", "--http.api=eth,web3,net,txpool,art,abft,debug", "--ws", "--ws.addr=0.0.0.0", "--ws.origins=*", "--ws.api=eth,web3,net,txpool,art,abft,debug", "--metrics.influxdbv2", "--validator.password=/data/password"]
