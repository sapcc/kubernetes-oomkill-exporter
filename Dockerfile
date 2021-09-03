FROM golang:1.16-alpine3.13 AS builder

WORKDIR /go/src/github.com/sapcc/kuberntes-oomkill-exporter
ADD go.mod go.sum ./
RUN go mod download
ADD cache/main.go .
RUN go build -v -o /dev/null
ADD . .
RUN CGOENABLED=0 go build -v -o /kubernetes-oomkill-exporter

FROM alpine:3.13
LABEL maintainer="jan.knipper@sap.com"
LABEL source_repository="https://github.com/sapcc/kubernetes-oomkill-exporter"

RUN apk --no-cache add ca-certificates
COPY --from=builder /kubernetes-oomkill-exporter /kubernetes-oomkill-exporter

ENTRYPOINT ["/kubernetes-oomkill-exporter"]
CMD ["-logtostderr"]
