package server

import (
    "encoding/base64"
	"fmt"
	"os"
	"path/filepath"

    pb "github.com/Districorp-UPB/FileServer/proto"
)

type FileService struct {
	pb.UnimplementedFileServiceServer
}

func (s *FileService) Upload(stream pb.FileService_UploadServer) error {
	// Leer el request desde el flujo de datos
	req, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive upload request: %w", err)
	}

	// Subir archivo
	uploadToNFS(req)
	if err != nil {
		return fmt.Errorf("failed to upload file to NFS: %w", err)
	}


	// Respuesta
	err = stream.SendAndClose(&pb.FileUploadResponse{
		FileId: req.FileId,
	})
	if err != nil {
		return fmt.Errorf("failed to send upload response: %w", err)
	}

	return nil
}


func uploadToNFS(req *pb.FileUploadRequest) (string, error) {
	// Si el usuario nunca ha subido un archivo, crear un directorio para el usuario
	userPath := fmt.Sprintf("./nfs/files/%s", req.OwnerId)
	if _, err := os.Stat(userPath); os.IsNotExist(err) {
		err := os.Mkdir(userPath, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create user directory: %w", err)
		}
	}

	// Extraer la extensión y el nombre del archivo
	fileExtension := filepath.Ext(req.FileName)
	fileName := req.FileId + fileExtension

	// Guardar el archivo en la ruta principal
	filePath := filepath.Join(userPath, fileName)
	err := saveFile(filePath, req.BinaryFile)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

}

func saveFile(filePath string, binaryFile []byte) error {
	fileUpload, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer fileUpload.Close()

	// Decodificar el contenido del binario que viene en base64
	decodedContent, err := base64.StdEncoding.DecodeString(string(binaryFile))
	if err != nil {
		return fmt.Errorf("failed to decode binary content: %w", err)
	}

	// Escribir el contenido decodificado en el archivo
	_, err = fileUpload.Write(decodedContent)
	if err != nil {
		return fmt.Errorf("failed to write binary content to file: %w", err)
	}

	return nil
}


