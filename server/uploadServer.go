package server

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	pb "github.com/Districorp-UPB/FileServer/proto"
)

type FileService struct {
	pb.UnimplementedFileServiceServer
}

// Manejo de la subida de archivos
func (s *FileService) Upload(stream pb.FileService_UploadServer) error {
	req, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive upload request: %w", err)
	}

	// Subir el archivo al NFS
	_, err = uploadToNFS(req)
	if err != nil {
		return fmt.Errorf("failed to upload file to NFS: %w", err)
	}

	// Enviar la respuesta al cliente
	err = stream.SendAndClose(&pb.FileUploadResponse{
		FileId: req.FileId,
	})
	if err != nil {
		return fmt.Errorf("failed to send upload response: %w", err)
	}

	return nil
}

// Manejo de la descarga de archivos
func (s *FileService) Download(req *pb.FileDownloadRequest, stream pb.FileService_DownloadServer) error {
	filePath, err := getFilePath(req.OwnerId, req.FileId)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	const bufferSize = 1024 * 1024 // 1 MB
	buffer := make([]byte, bufferSize)

	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read file: %w", err)
		}
		if n == 0 { // Si no hay más datos para leer
			break
		}

		// Enviar fragmento
		if err := stream.Send(&pb.FileDownloadResponse{
			FileId:             req.FileId,
			BinaryFileResponse: buffer[:n],
		}); err != nil {
			return fmt.Errorf("failed to send file chunk: %w", err)
		}
	}

	return nil
}

func uploadToNFS(req *pb.FileUploadRequest) (string, error) {
	userPath := fmt.Sprintf("./nfs/files/%s", req.OwnerId)
	if _, err := os.Stat(userPath); os.IsNotExist(err) {
		err := os.MkdirAll(userPath, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create user directory: %w", err)
		}
	}

	fileExtension := filepath.Ext(req.FileName)
	fileName := req.FileId + fileExtension
	filePath := filepath.Join(userPath, fileName)

	// Guardar el archivo recibido
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

	// No es necesario decodificar, solo escribir el contenido binario directamente
	_, err = fileUpload.Write(binaryFile)
	if err != nil {
		return fmt.Errorf("failed to write binary content to file: %w", err)
	}

	return nil
}

func getFilePath(ownerId, fileId string) (string, error) {
	// Directorio donde se almacenan los archivos
	userPath := fmt.Sprintf("./nfs/files/%s", ownerId)

	// Buscar archivos con el fileId seguido de cualquier extensión
	filePattern := fmt.Sprintf("%s*", fileId)
	matchedFiles, err := filepath.Glob(filepath.Join(userPath, filePattern))
	if err != nil || len(matchedFiles) == 0 {
		return "", fmt.Errorf("file does not exist")
	}

	// Retornar el primer archivo encontrado
	return matchedFiles[0], nil
}
