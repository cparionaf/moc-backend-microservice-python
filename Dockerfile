# Usar la imagen oficial de Go como base
FROM golang:1.23 AS builder

# Establecer el directorio de trabajo
WORKDIR /app

# Copiar los archivos go.mod y go.sum
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar la aplicación
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Imagen final
FROM alpine:latest

# Copiar el ejecutable compilado
COPY --from=builder /app/main /app/main

# Exponer el puerto 8080
EXPOSE 8080

# Ejecutar la aplicación
CMD ["/app/main"]