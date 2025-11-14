FROM golang:1.20-alpine AS builder

WORKDIR /src

# 1. Copiar go.mod y go.sum
COPY go.mod go.sum ./

# 2. Descargar dependencias
RUN go mod download

# 3. Copiar TODO el proyecto
COPY . .

# 4. Compilar
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway .

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=builder /src/gateway .
COPY static ./static

ENV PORT=8088
EXPOSE 8088

CMD ["./gateway"]
