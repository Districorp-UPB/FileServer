package main

import (
	"log"
	"net"
	"time"

	pb "github.com/Districorp-UPB/FileServer/proto"
	"github.com/Districorp-UPB/FileServer/server"
	"google.golang.org/grpc"
)

func main() {
	// Escuchar en el puerto 50051
	grpcListener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer grpcListener.Close()

	// Iniciar servidor gRPC con límite de 1GB y tiempo de espera de 5 minutos
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(1024*1024*1024),
		grpc.ConnectionTimeout(time.Minute*5),
	)

	// Registrar el servicio de archivos
	pb.RegisterFileServiceServer(grpcServer, &server.FileService{})
	log.Println("gRPC server started, listening on port 50051")

	// Mantener el servidor ejecutándose y escuchando peticiones
	if err := grpcServer.Serve(grpcListener); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
