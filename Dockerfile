# syntax=docker/dockerfile:1

FROM golang:1.25.5-alpine AS deps
WORKDIR /src
ENV CGO_ENABLED=0 GOFLAGS=-buildvcs=false

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

FROM deps AS build
COPY cmd ./cmd
COPY internal ./internal
COPY assets ./assets
COPY migrations ./migrations
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    mkdir -p /out && go build -trimpath -ldflags="-s -w" -o /out/main ./cmd/api
    
COPY docs ./docs

FROM scratch AS prod
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/main /main
COPY --from=build /src/docs /docs
EXPOSE 8080
ENTRYPOINT ["/main"]
