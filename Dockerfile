FROM docker.io/golang:1.19 AS builder
WORKDIR /gosnappass
COPY . .
RUN make tidy && make build

FROM registry.access.redhat.com/ubi9/ubi-micro:latest

# this label automatically links this to a repository in GHCR.
LABEL org.opencontainers.image.source https://github.com/concerthall/gosnappass

WORKDIR /gosnappass
COPY --from=builder /gosnappass/build/gosnappass ./
CMD ["./gosnappass"]