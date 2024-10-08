package server

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"

	pb "github.com/Districorp-UPB/FileServer/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FileService struct {
	pb.UnimplementedFileServiceServer	
}

func (s *FileService) Upload(stream pb.FileService_UploadServer) error {
	for {
		// Leer los chunks del archivo enviados por el cliente
		req, err := stream.Recv()
		if err == io.EOF {
			// Si ya se recibieron todos los chunks
			break
		}
		if err != nil {
			log.Printf("Error al recibir chunks: %v", err)
			return fmt.Errorf("failed to receive chunks: %w", err)
		}

		// Aquí procesa el chunk recibido, por ejemplo, guarda el archivo en disco
		_, err = uploadToNFS(req)
		if err != nil {
			log.Printf("Error al subir el archivo al NFS: %v", err)
			return fmt.Errorf("failed to upload file to NFS: %w", err)
		}
	}

	// Responder al cliente con éxito cuando todos los chunks se han recibido
	err := stream.SendAndClose(&pb.FileUploadResponse{
		FileId: "some-file-id", // Aquí debería estar el ID del archivo subido
	})
	if err != nil {
		log.Printf("Error al enviar respuesta al cliente: %v", err)
		return fmt.Errorf("failed to send response: %w", err)
	}

	return nil
}


func uploadToNFS(req *pb.FileUploadRequest) (string, error) {
	userPath := fmt.Sprintf("./nfs/files/%s", req.OwnerId)
	if err := os.MkdirAll(userPath, 0755); err != nil {
		log.Printf("Error al crear el directorio del usuario %s: %v", req.OwnerId, err)
		return "", fmt.Errorf("failed to create user directory: %w", err)
	}
	log.Printf("Directorio creado/verificado para el usuario %s: %s", req.OwnerId, userPath)

	fileExtension := filepath.Ext(req.FileName)
	fileName := req.FileId + fileExtension
	filePath := filepath.Join(userPath, fileName)

	// Verificar espacio en disco
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(userPath, &fs)
	if err != nil {
		log.Printf("Error al verificar espacio en disco: %v", err)
		return "", fmt.Errorf("failed to check disk space: %w", err)
	}

	// Calcular espacio libre en bytes
	freeSpace := fs.Bfree * uint64(fs.Bsize)
	if freeSpace < uint64(len(req.BinaryFile)) {
		log.Printf("No hay suficiente espacio en disco. Libre: %d bytes, Necesario: %d bytes", freeSpace, len(req.BinaryFile))
		return "", fmt.Errorf("not enough disk space")
	}

	err = saveFile(filePath, req.BinaryFile)
	if err != nil {
		log.Printf("Error al guardar el archivo %s: %v", fileName, err)
		return "", fmt.Errorf("failed to save file: %w", err)
	}
	log.Printf("Archivo %s guardado en la ruta %s", fileName, filePath)
	return filePath, nil
}

func appendToFile(filePath string, binaryFile []byte) error {
	fileUpload, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error al abrir el archivo %s: %v", filePath, err)
		return fmt.Errorf("failed to open file for appending: %w", err)
	}
	defer fileUpload.Close()

	_, err = fileUpload.Write(binaryFile)
	if err != nil {
		log.Printf("Error al escribir el archivo %s: %v", filePath, err)
		return fmt.Errorf("failed to write binary content to file: %w", err)
	}
	log.Printf("Chunk guardado exitosamente en %s", filePath)
	return nil
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