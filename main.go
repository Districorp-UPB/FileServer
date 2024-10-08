package main

import (
    "log"
    "net"
    "os"
    "time"

    pb "github.com/Districorp-UPB/FileServer/proto"
    "github.com/Districorp-UPB/FileServer/server"
    "google.golang.org/grpc"
    "google.golang.org/grpc/grpclog"
    "google.golang.org/grpc/keepalive"
)

func init() {
    // Configurar el logger de gRPC
    grpclog.SetLoggerV2(grpclog.NewLoggerV2WithVerbosity(os.Stdout, os.Stdout, os.Stderr, 4))
}

func main() {
    log.Println("Iniciando la aplicación...")

    grpc.EnableTracing = true

    // Servidor gRPC
    grpcListener, err := net.Listen("tcp", "0.0.0.0:50051")
    if err != nil {
        log.Fatalf("Error crítico: no se pudo escuchar en el puerto 50051: %v", err)
    }
    log.Println("Listener de gRPC creado, esperando conexiones...")
    defer func() {
        log.Println("Cerrando el listener de gRPC...")
        if err := grpcListener.Close(); err != nil {
            log.Fatalf("Error al cerrar el listener: %v", err)
        }
        log.Println("Listener cerrado correctamente.")
    }()

    // Configuración de keepalive
    kasp := keepalive.ServerParameters{
        MaxConnectionIdle:     15 * time.Minute,
        MaxConnectionAge:      30 * time.Minute,
        MaxConnectionAgeGrace: 5 * time.Minute,
        Time:                  5 * time.Minute,
        Timeout:               1 * time.Minute,
    }

    // Iniciar servidor gRPC con timeout de 10 minutos y límite de 1GB
    grpcServer := grpc.NewServer(
        grpc.MaxRecvMsgSize(1024*1024*1024),
        grpc.ConnectionTimeout(time.Minute*10),
        grpc.KeepaliveParams(kasp),
    )

    pb.RegisterFileServiceServer(grpcServer, &server.FileService{})
    log.Println("Servidor gRPC iniciado en el puerto 50051.")

    // Ejecutar el servidor
    log.Println("Ejecutando el servidor gRPC...")
    if err := grpcServer.Serve(grpcListener); err != nil {
        log.Fatalf("Error crítico: fallo al servir gRPC: %v", err)
    }

    log.Println("Servidor gRPC detenido.")
}
