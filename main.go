package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
)

// ServerInfo contiene información detallada sobre la instancia y región
type ServerInfo struct {
    Region          string    `json:"region"`
    AvailabilityZone string   `json:"availability_zone"`
    InstanceID      string    `json:"instance_id"`
    InstanceType    string    `json:"instance_type"`
    TimeStamp       time.Time `json:"timestamp"`
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
        // Fallback a la región de la configuración
        info.Region = cfg.Region
    } else {
        info.Region = region.Region
    }

    // Obtener zona de disponibilidad
    az, err := imdsClient.GetMetadata(ctx, &imds.GetMetadataInput{
        Path: "placement/availability-zone",
    })
    if err == nil {
        info.AvailabilityZone = az
    }

    // Obtener ID de instancia
    instanceID, err := imdsClient.GetMetadata(ctx, &imds.GetMetadataInput{
        Path: "instance-id",
    })
    if err == nil {
        info.InstanceID = instanceID
    }

    // Obtener tipo de instancia
    instanceType, err := imdsClient.GetMetadata(ctx, &imds.GetMetadataInput{
        Path: "instance-type",
    })
    if err == nil {
        info.InstanceType = instanceType
    }

    // Enviar respuesta
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(info)
}