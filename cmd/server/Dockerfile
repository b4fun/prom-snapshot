ARG nonroot_image=gcr.io/distroless/static:nonroot

FROM ${nonroot_image}
WORKDIR /
COPY prom-snapshot-server prom-snapshot-server
USER nonroot:nonroot

ENTRYPOINT ["/prom-snapshot-server"]