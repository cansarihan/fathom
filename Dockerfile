FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
ARG VERSION=dev
ARG COMMIT=none
RUN CGO_ENABLED=0 go build -trimpath \
      -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
      -o /out/fathom ./cmd/fathom

# The runtime image keeps the Go toolchain because fathom drives `go test -fuzz`.
FROM golang:1.25-alpine
RUN adduser -D -u 10001 fathom
ENV HOME=/tmp GOCACHE=/tmp/go-cache GOPATH=/tmp/go GOFLAGS=-mod=mod
COPY --from=build /out/fathom /usr/local/bin/fathom
COPY --from=build /src/deploy/sample /work/sample
USER fathom
WORKDIR /work
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/fathom"]
CMD ["serve", "-dir", "/work/sample", "-listen", "0.0.0.0:8080", "-t", "8s"]
