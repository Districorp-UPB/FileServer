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
    grpclog.SetLoggerV2(grpclog.NewLoggerV2(os.Stdout, os.Stdout, os.Stderr))
}

func main() {
    grpc.EnableTracing = true

    // Servidor gRPC
    grpcListener, err := net.Listen("tcp", "0.0.0.0:50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    defer grpcListener.Close()

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
    log.Println("gRPC server started on port 50051")

    // Ejecutar el servidor
    if err := grpcServer.Serve(grpcListener); err != nil {
        log.Fatalf("failed to serve gRPC: %v", err)
    }
}