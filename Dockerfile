FROM golang:alpine as builder

WORKDIR /source

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLED=0 GOOS=linux go build -o /webhook-server main.go 


FROM alpine:latest

LABEL org.opencontainers.image.source=https://github.com/igresc/validating-admission-webhook-go
LABEL org.opencontainers.image.licenses=MIT

COPY --from=builder /webhook-server /webhook-server
CMD [ "/webhook-server", "--port", "8443", "--tls-key", "/etc/certs/tls.key", "--tls-cert", "/etc/certs/tls.crt" ]

EXPOSE 8443