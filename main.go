package main

import (
    "context"
    "encoding/json"
    "io"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
)

// ServerInfo contiene información detallada sobre la instancia y región
type ServerInfo struct {
    Region           string    `json:"region"`
    AvailabilityZone string    `json:"availability_zone"`
    InstanceID       string    `json:"instance_id"`
    InstanceType     string    `json:"instance_type"`
    TimeStamp        time.Time `json:"timestamp"`
}

// readMetadata es una función auxiliar para leer el contenido del ReadCloser
func readMetadata(content io.ReadCloser) (string, error) {
    // Es importante cerrar el reader cuando terminemos
    defer content.Close()
    
    // Leemos todo el contenido
    bytes, err := io.ReadAll(content)
    if err != nil {
        return "", err
    }
    
    return string(bytes), nil
}

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    http.HandleFunc("/info", getServerInfoHandler)
    
    log.Printf("Servidor iniciado en puerto %s", port)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatal(err)
    }
}

func getServerInfoHandler(w http.ResponseWriter, r *http.Request) {
    ctx := context.TODO()
    
    // Cargar configuración de AWS con soporte para IMDS
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        http.Error(w, "Error al cargar configuración AWS: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Crear cliente IMDS
    imdsClient := imds.NewFromConfig(cfg)

    // Obtener información de la instancia
    info := ServerInfo{
        TimeStamp: time.Now(),
    }

    // Obtener región
    region, err := imdsClient.GetRegion(ctx, &imds.GetRegionInput{})
    if err != nil {
        log.Printf("Error al obtener región desde IMDS: %v", err)
        info.Region = cfg.Region
    } else {
        info.Region = region.Region
    }

    // Obtener zona de disponibilidad
    azOutput, err := imdsClient.GetMetadata(ctx, &imds.GetMetadataInput{
        Path: "placement/availability-zone",
    })
    if err == nil && azOutput != nil {
        if azContent, err := readMetadata(azOutput.Content); err == nil {
            info.AvailabilityZone = azContent
            log.Printf("AZ obtenida: %s", azContent)
        } else {
            log.Printf("Error leyendo AZ: %v", err)
        }
    }

    // Obtener ID de instancia
    instanceIDOutput, err := imdsClient.GetMetadata(ctx, &imds.GetMetadataInput{
        Path: "instance-id",
    })
    if err == nil && instanceIDOutput != nil {
        if idContent, err := readMetadata(instanceIDOutput.Content); err == nil {
            info.InstanceID = idContent
            log.Printf("Instance ID obtenido: %s", idContent)
        } else {
            log.Printf("Error leyendo Instance ID: %v", err)
        }
    }

    // Obtener tipo de instancia
    instanceTypeOutput, err := imdsClient.GetMetadata(ctx, &imds.GetMetadataInput{
        Path: "instance-type",
    })
    if err == nil && instanceTypeOutput != nil {
        if typeContent, err := readMetadata(instanceTypeOutput.Content); err == nil {
            info.InstanceType = typeContent
            log.Printf("Instance Type obtenido: %s", typeContent)
        } else {
            log.Printf("Error leyendo Instance Type: %v", err)
        }
    }

    // Enviar respuesta
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(info)
}