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
	// Servidor gRPC
	grpcListener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer grpcListener.Close()


	// Iniciar servidor gRPC con timeout de 5 minutos y limite de 1GB
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(1024*1024*1024),
		grpc.ConnectionTimeout(time.Minute*5),
	)
	pb.RegisterFileServiceServer(grpcServer, &server.FileService{})
	log.Println("gRPC server started")
	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	
}
