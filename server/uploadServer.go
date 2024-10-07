package server

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	pb "github.com/Districorp-UPB/FileServer/proto"
)

type FileService struct {
	pb.UnimplementedFileServiceServer
}

func (s *FileService) Upload(stream pb.FileService_UploadServer) error {
	// Leer el request desde el flujo de datos
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break // Fin del stream
		}
		if err != nil {
			log.Printf("Error al recibir el request de subida de archivo: %v", err)
			return fmt.Errorf("failed to receive upload request: %w", err)
		}

		// Intentar subir el archivo al NFS
		log.Printf("Intentando subir el archivo %s del usuario %s", req.FileName, req.OwnerId)
		_, err = uploadToNFS(req)
		if err != nil {
			log.Printf("Error al subir el archivo al NFS: %v", err)
			return fmt.Errorf("failed to upload file to NFS: %w", err)
		}
	}

	// Responder al cliente con éxito
	err := stream.SendAndClose(&pb.FileUploadResponse{
		FileId: "unique-file-id", // Puedes usar un ID real o generarlo
	})
	if err != nil {
		log.Printf("Error al enviar respuesta de éxito al cliente: %v", err)
		return fmt.Errorf("failed to send upload response: %w", err)
	}

	log.Printf("Archivo subido correctamente")
	return nil
}

func uploadToNFS(req *pb.FileUploadRequest) (string, error) {
	// Crear directorio de usuario si no existe
	userPath := fmt.Sprintf("./nfs/files/%s", req.OwnerId)
	if _, err := os.Stat(userPath); os.IsNotExist(err) {
		err := os.MkdirAll(userPath, 0755)
		if err != nil {
			log.Printf("Error al crear el directorio del usuario %s: %v", req.OwnerId, err)
			return "", fmt.Errorf("failed to create user directory: %w", err)
		}
	}

	// Guardar archivo en el sistema NFS
	fileExtension := filepath.Ext(req.FileName)
	fileName := req.FileId + fileExtension
	filePath := filepath.Join(userPath, fileName)

	err := saveFile(filePath, req.BinaryFile)
	if err != nil {
		log.Printf("Error al guardar el archivo %s: %v", fileName, err)
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	log.Printf("Archivo %s guardado en la ruta %s", fileName, filePath)
	return filePath, nil
}

func saveFile(filePath string, binaryFile []byte) error {
	fileUpload, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error al crear el archivo %s: %v", filePath, err)
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer fileUpload.Close()

	_, err = fileUpload.Write(binaryFile)
	if err != nil {
		log.Printf("Error al escribir el archivo %s: %v", filePath, err)
		return fmt.Errorf("failed to write binary content to file: %w", err)
	}

	log.Printf("Archivo %s guardado exitosamente", filePath)
	return nil
}
