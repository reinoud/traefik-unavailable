# syntax=docker/dockerfile:1
# Bouwt op het eigen platform en compileert voor het doelplatform: geen emulatie nodig voor arm64.
FROM --platform=$BUILDPLATFORM golang:1.27.4-alpine@sha256:70bd7ddea8b6cbe5d6943cde279d6b67feb10cc924fb9100d1d8757c1575b280 AS bouw
ENV CGO_ENABLED=0
WORKDIR /src
COPY go.mod *.go ./
# Tests vóór ARG TARGETARCH: één run, gedeeld door alle platformen.
RUN go vet ./... && go test ./...
ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags='-s -w' -o /vangnet .

FROM scratch
COPY --from=bouw /vangnet /vangnet
USER 65532:65532
EXPOSE 8080 8081
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --start-interval=1s CMD ["/vangnet", "-health"]
ENTRYPOINT ["/vangnet"]
