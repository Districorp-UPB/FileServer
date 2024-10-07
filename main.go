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
)

func init() {
	// Configurar el logger de gRPC
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(os.Stdout, os.Stdout, os.Stderr))
}

func main() {
	// Servidor gRPC
	grpcListener, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer grpcListener.Close()

	// Iniciar servidor gRPC con timeout de 5 minutos y límite de 1GB
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(1024*1024*1024),
		grpc.ConnectionTimeout(time.Minute*5),
	)
	pb.RegisterFileServiceServer(grpcServer, &server.FileService{})
	log.Println("gRPC server started")

	// Ejecutar el servidor en una goroutine para manejar las conexiones
	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// Mantener el servidor en ejecución indefinidamente
	select {}
}
