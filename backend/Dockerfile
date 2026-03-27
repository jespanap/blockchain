FROM golang:1.21-alpine AS builder

WORKDIR /app

# 1. Copiamos los archivos de dependencias
COPY go.mod go.sum* ./

# 2. Descargamos las dependencias existentes
RUN go mod download

# 3. Copiamos el resto del código fuente
COPY . .

# 4. ¡PASO CLAVE!: Limpiamos y actualizamos el registro de módulos
# Esto agregará github.com/gorilla/mux al go.mod/go.sum si falta
RUN go mod tidy

# 5. Compilamos
RUN go build -o blockchain-api main.go

# --- Etapa Final ---
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/blockchain-api .
EXPOSE 8080
CMD ["./blockchain-api"]