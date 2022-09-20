FROM golang:1.19-alpine3.16 AS builder

WORKDIR /go/src/github.com/sapcc/kuberntes-oomkill-exporter
ADD go.mod go.sum ./
RUN go mod download
ADD cache/main.go .
RUN CGO_ENABLED=0 go build -v -o /dev/null
ADD . .
RUN CGO_ENABLED=0 go build -v -o /kubernetes-oomkill-exporter

FROM alpine:3.16
LABEL maintainer="jan.knipper@sap.com"
LABEL source_repository="https://github.com/sapcc/kubernetes-oomkill-exporter"

RUN apk --no-cache add ca-certificates
COPY --from=builder /kubernetes-oomkill-exporter /kubernetes-oomkill-exporter

ENTRYPOINT ["/kubernetes-oomkill-exporter"]
CMD ["-logtostderr"]
