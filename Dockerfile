# syntax=docker/dockerfile:1
# Minecraft Dungeons' own services for OpenPak: the api.minecraftservices.com stand-in and the
# WebRTC signaling relay. Plain HTTP inside the pod network; Traefik terminates the console's TLS.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
ARG TARGETOS TARGETARCH VERSION=dev
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -buildvcs=false \
    -ldflags="-s -w -X main.version=${VERSION}" -o /out/dungeonsd ./cmd/dungeonsd

FROM gcr.io/distroless/static
COPY --from=build /out/dungeonsd /dungeonsd
USER nonroot:nonroot
EXPOSE 23520 23521 23522
HEALTHCHECK CMD ["/dungeonsd", "healthcheck"]
ENTRYPOINT ["/dungeonsd"]
