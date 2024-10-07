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

func main() {
	// Configurar logging básico
	log.Println("Iniciando el servidor...")

	// Configurar logging avanzado para gRPC
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(os.Stdout, os.Stdout, os.Stderr))

	// Configurar listener en el puerto 50051
	grpcListener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Error al iniciar el listener: %v", err)
	}
	defer grpcListener.Close()

	log.Println("gRPC listener iniciado en :50051")

	// Configurar el servidor gRPC con límite de tamaño y tiempo de espera
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(1024*1024*1024),   // Máximo 1GB de datos
		grpc.ConnectionTimeout(time.Minute*5), // Tiempo de espera de 5 minutos
	)

	// Registrar el servicio de archivos
	pb.RegisterFileServiceServer(grpcServer, &server.FileService{})

	// Iniciar el servidor gRPC en una goroutine
	log.Println("Servidor gRPC iniciado y esperando conexiones...")
	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("Error al servir gRPC: %v", err)
		}
	}()

	// Mantener el servidor en ejecución indefinidamente
	select {}
}
