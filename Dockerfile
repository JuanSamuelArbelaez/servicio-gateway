FROM golang:1.20-alpine AS builder
WORKDIR /src
COPY go.mod .
COPY *.go .
RUN apk add --no-cache git
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway .

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /src/gateway .
COPY static ./static
ENV PORT=8080
EXPOSE 8080
CMD ["./gateway"]

