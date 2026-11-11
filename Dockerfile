FROM golang:1.26-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /webhook .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /webhook /webhook
LABEL org.opencontainers.image.source="https://github.com/Gyro-Hosting/cert-manager-webhook-gyro" org.opencontainers.image.description="cert-manager ACME DNS-01 webhook for Gyro DNS" org.opencontainers.image.licenses="MIT"
USER 65532:65532
ENTRYPOINT ["/webhook"]
