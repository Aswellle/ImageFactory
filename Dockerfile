# ImageForge backend image — multi-stage build.
ARG GOLANG_IMAGE=golang:1.26-alpine
ARG ALPINE_IMAGE=alpine:3.21

FROM ${GOLANG_IMAGE} AS builder
WORKDIR /src/backend
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build with CGO disabled for a static binary; embed mode is off here.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /imageforge ./cmd/server

FROM ${ALPINE_IMAGE} AS runtime
RUN apk add --no-cache ca-certificates tzdata postgresql16-client
RUN addgroup -S iforge && adduser -S iforge -G iforge
USER iforge
WORKDIR /app
COPY --from=builder /imageforge /app/imageforge
EXPOSE 8080
ENTRYPOINT ["/app/imageforge"]
