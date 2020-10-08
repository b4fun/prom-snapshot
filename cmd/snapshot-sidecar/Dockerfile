ARG nonroot_image=gcr.io/distroless/static:nonroot

FROM ${nonroot_image}
WORKDIR /
COPY snapshot-sidecar snapshot-sidecar
USER nonroot:nonroot

ENTRYPOINT ["/snapshot-sidecar"]