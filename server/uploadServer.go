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
	var fullBinaryFile []byte
	var fileId, fileName string

	// Leer todas las partes del archivo desde el stream
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// Fin del archivo, proceder a guardar y responder con el FileId
			_, saveErr := uploadToNFS(&pb.FileUploadRequest{
				FileId:     fileId,
				BinaryFile: fullBinaryFile,
				FileName:   fileName,
			})
			if saveErr != nil {
				log.Printf("Error al guardar el archivo en NFS: %v", saveErr)
				return fmt.Errorf("failed to save file: %w", saveErr)
			}

			err = stream.SendAndClose(&pb.FileUploadResponse{
				FileId: fileId,
			})
			if err != nil {
				log.Printf("Error al enviar respuesta de éxito al cliente: %v", err)
				return fmt.Errorf("failed to send upload response: %w", err)
			}

			log.Printf("Archivo %s subido correctamente", fileName)
			return nil
		}
		if err != nil {
			log.Printf("Error al recibir el request de subida de archivo: %v", err)
			return fmt.Errorf("failed to receive upload request: %w", err)
		}

		// Almacenar la información del archivo
		fullBinaryFile = append(fullBinaryFile, req.BinaryFile...)
		fileId = req.FileId
		fileName = req.FileName
	}
}

func uploadToNFS(req *pb.FileUploadRequest) (string, error) {
	// Crear directorio de usuario si no existe
	userPath := fmt.Sprintf("./nfs/files/%s", req.OwnerId)
	if _, err := os.Stat(userPath); os.IsNotExist(err) {
		err := os.Mkdir(userPath, 0755)
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
