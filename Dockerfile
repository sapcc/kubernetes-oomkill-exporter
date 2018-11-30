FROM alpine:3.8
LABEL maintainer="jan.knipper@sap.com"

RUN apk --no-cache add ca-certificates
COPY kubernetes-oomkill-exporter /kubernetes-oomkill-exporter

ENTRYPOINT ["/kubernetes-oomkill-exporter"]
CMD ["-logtostderr"]
