FROM docker.m.daocloud.io/library/golang:1.23.1-alpine AS builder

WORKDIR /src

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn
ARG TARGETOS=linux
ARG TARGETARCH=amd64

COPY apps/api/go.mod ./
RUN go mod download

COPY apps/api ./
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /out/voidlab-api ./cmd/server

FROM docker.m.daocloud.io/library/alpine:3.20

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /out/voidlab-api /app/voidlab-api

USER app

EXPOSE 8080

ENTRYPOINT ["/app/voidlab-api"]
