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

// Manejo de la subida de archivos
func (s *FileService) Upload(stream pb.FileService_UploadServer) error {
	req, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive upload request: %w", err)
	}

	_, err = uploadToNFS(req)
	if err != nil {
		return fmt.Errorf("failed to upload file to NFS: %w", err)
	}

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

	binaryFile, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	encodedContent := base64.StdEncoding.EncodeToString(binaryFile)

	return stream.Send(&pb.FileDownloadResponse{
		FileId:             req.FileId,
		BinaryFileResponse: []byte(encodedContent),
	})
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

func getFilePath(ownerId, fileId string) (string, error) {
	userPath := fmt.Sprintf("./nfs/files/%s", ownerId)
	fileName := fmt.Sprintf("%s", fileId)
	filePath := filepath.Join(userPath, fileName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist")
	}

	return filePath, nil
}
