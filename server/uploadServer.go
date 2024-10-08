package server

import (
    "encoding/base64"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "io"

    pb "github.com/Districorp-UPB/FileServer/proto"
)

type FileService struct {
    pb.UnimplementedFileServiceServer
}

func (s *FileService) Upload(stream pb.FileService_UploadServer) error {
    log.Println("Comenzando la carga del archivo...")

    var fileId string

    for {
        req, err := stream.Recv()
        if err != nil {
            if err == io.EOF {
                log.Println("Fin de la transmisión recibido.")
                break
            }
            return fmt.Errorf("failed to receive upload request: %w", err)
        }

        log.Printf("Recibido chunk del archivo con ID: %s de tamaño %d bytes\n", req.FileId, len(req.BinaryFile))

        fileId = req.FileId

        _, err = uploadToNFS(req)
        if err != nil {
            log.Printf("Error al subir archivo a NFS: %v", err)
            return fmt.Errorf("failed to upload file to NFS: %w", err)
        }
    }

    log.Println("Preparando la respuesta para el cliente...")
    err := stream.SendAndClose(&pb.FileUploadResponse{
        FileId: fileId,
    })
    if err != nil {
        log.Printf("Error al enviar la respuesta: %v", err)
        return fmt.Errorf("failed to send upload response: %w", err)
    }

    log.Println("Archivo subido exitosamente.")
    return nil
}

func uploadToNFS(req *pb.FileUploadRequest) (string, error) {
    userPath := fmt.Sprintf("./nfs/files/%s", req.OwnerId)
    if _, err := os.Stat(userPath); os.IsNotExist(err) {
        err := os.Mkdir(userPath, 0755)
        if err != nil {
            return "", fmt.Errorf("failed to create user directory: %w", err)
        }
        log.Printf("Directorio creado para el usuario: %s", req.OwnerId)
    }

    fileExtension := filepath.Ext(req.FileName)
    fileName := req.FileId + fileExtension

    filePath := filepath.Join(userPath, fileName)
    err := saveFile(filePath, req.BinaryFile)
    if err != nil {
        return "", fmt.Errorf("failed to upload file: %w", err)
    }

    log.Printf("Archivo guardado en: %s", filePath)
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

    if _, err = fileUpload.Write(decodedContent); err != nil {
        return fmt.Errorf("failed to write binary content to file: %w", err)
    }

    log.Printf("Contenido decodificado escrito en el archivo: %s", filePath)
    return nil
}
