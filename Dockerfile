FROM --platform=$BUILDPLATFORM golang:1.26.5-alpine AS build
WORKDIR /src
ARG TARGETOS
ARG TARGETARCH
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/github-repository-settings ./cmd/github-repository-settings

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/github-repository-settings /github-repository-settings
ENTRYPOINT ["/github-repository-settings"]
