# Build the manager binary
FROM europe-docker.pkg.dev/kyma-project/prod/external/golang:1.22.0-alpine3.19 as builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY main.go main.go
COPY converter/ converter/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o crd-conversion-webhook main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/crd-conversion-webhook .

USER 65532:65532

ENTRYPOINT ["/crd-conversion-webhook"]
