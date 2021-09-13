FROM golang:1.16.5-alpine3.14@sha256:3361eb3ffa949cf0cf60c13778697183a22684e0a0a5edf9ccb9d2f1ae4da873 as BUILD

WORKDIR /build
COPY . .
RUN go generate ./... && \
    go build -tags netgo -o /usr/bin/cftest ./cmd/cftest

FROM alpine:3.14.0@sha256:234cb88d3020898631af0ccbbcca9a66ae7306ecd30c9720690858c1b007d2a0

COPY --from=build /usr/bin/cftest /usr/bin/cftest

ENTRYPOINT ["/usr/bin/cftest"]

