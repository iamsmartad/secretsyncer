# syntax=docker/dockerfile:1
FROM golang:1.17 as builder
WORKDIR /app/
COPY src/go.mod src/go.sum ./
RUN go mod download
COPY src/*.go .
COPY src/helpers helpers
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -o app .

FROM alpine:latest as runner
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/app .
CMD ["./app"]
