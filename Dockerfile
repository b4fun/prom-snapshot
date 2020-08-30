ARG nonroot_image=gcr.io/distroless/static:nonroot

FROM ${nonroot_image}
WORKDIR /
COPY ./docker_bin/prom-snapshot-server-amd64 prom-snapshot-server
USER nonroot:nonroot

ENTRYPOINT ["/prom-snapshot-server"]