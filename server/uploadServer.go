package server

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	pb "../proto"
)

type FileService struct {
	pb.UnimplementedFileServiceServer
}

// Método para subir un archivo (ya existente)
func (s *FileService) Upload(stream pb.FileService_UploadServer) error {
	req, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive upload request: %w", err)
	}

	// Subir archivo
	_, err = uploadToNFS(req)
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

// Método para descargar un archivo
func (s *FileService) Download(req *pb.FileDownloadRequest, stream pb.FileService_DownloadServer) error {
	// Construir la ruta del archivo según el OwnerId y FileId
	userPath := fmt.Sprintf("./nfs/files/%s", req.OwnerId)
	filePath := filepath.Join(userPath, req.FileId+filepath.Ext(req.FileName))

	// Verificar si el archivo existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found")
	}

	// Leer el archivo desde el sistema de archivos
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Codificar el archivo en base64 para enviar al cliente
	encodedFile := base64.StdEncoding.EncodeToString(fileData)

	// Enviar la respuesta con el archivo
	err = stream.Send(&pb.FileDownloadResponse{
		FileName:    req.FileName,
		BinaryFile:  []byte(encodedFile),
	})
	if err != nil {
		return fmt.Errorf("failed to send file: %w", err)
	}

	return nil
}

// Función para guardar archivos en el NFS (ya existente)
func uploadToNFS(req *pb.FileUploadRequest) (string, error) {
	userPath := fmt.Sprintf("./nfs/files/%s", req.OwnerId)
	if _, err := os.Stat(userPath); os.IsNotExist(err) {
		err := os.Mkdir(userPath, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create user directory: %w", err)
		}
	}

	fileExtension := filepath.Ext(req.FileName)
	fileName := req.FileId + fileExtension
	filePath := filepath.Join(userPath, fileName)

	err := saveFile(filePath, req.BinaryFile)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return filePath, nil
}

func saveFile(filePath string, binaryFile []byte) error {
	fileUpload, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer fileUpload.Close()

	decodedContent, err := base64.StdEncoding.DecodeString(string(binaryFile))
	if err != nil {
		return fmt.Errorf("failed to decode binary content: %w", err)
	}

	_, err = fileUpload.Write(decodedContent)
	if err != nil {
		return fmt.Errorf("failed to write binary content to file: %w", err)
	}

	return nil
}
