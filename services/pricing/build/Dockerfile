FROM golang:1.26.1-alpine AS builder
WORKDIR /app
ENV GOWORK=off
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o /pricing ./cmd

FROM scratch AS release
WORKDIR /app
COPY --from=builder /pricing /pricing
COPY config/config.yaml /app/config/config.yaml
COPY tariffs.json /app/tariffs.json
USER 65532:65532
EXPOSE 8081 9090
ENTRYPOINT ["/pricing"]
