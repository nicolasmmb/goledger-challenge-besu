FROM golang:1.22 AS builder
WORKDIR /src
COPY . .
RUN go mod download && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app ./cmd/api

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /out/app /app/app
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
