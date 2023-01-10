FROM docker.io/golang:1.19 AS builder
WORKDIR /gosnappass
COPY . .
RUN make tidy && make build

FROM registry.access.redhat.com/ubi9/ubi-micro:latest
WORKDIR /gosnappass
COPY --from=builder /gosnappass/build/gosnappass ./
CMD ["./gosnappass"]