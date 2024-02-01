FROM golang:1.21 AS builder

WORKDIR /go/src/github.com/sapcc/kuberntes-oomkill-exporter
ADD go.mod go.sum ./
RUN go mod download
ADD cache/main.go .
RUN CGO_ENABLED=0 go build -v -o /dev/null
ADD . .
RUN go test -v .
RUN CGO_ENABLED=0 go build -v -o /kubernetes-oomkill-exporter

RUN apt update -qqq && \
    apt install -yqqq ca-certificates && \
    update-ca-certificates

FROM gcr.io/distroless/static-debian12
LABEL maintainer="jan.knipper@sap.com"
LABEL source_repository="https://github.com/sapcc/kubernetes-oomkill-exporter"

COPY --from=builder /kubernetes-oomkill-exporter /kubernetes-oomkill-exporter
COPY --from=builder /etc/ssl/certs /etc/ssl/certs

ENTRYPOINT ["/kubernetes-oomkill-exporter"]
CMD ["-logtostderr"]
