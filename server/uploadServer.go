package server

import (
    "encoding/base64"
    "fmt"
    "log" // Importar el paquete log
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

    // Acumulador para el FileId
    var fileId string

    for {
        req, err := stream.Recv()
        if err != nil {
            if err == io.EOF {
                log.Println("Fin del flujo recibido.")
                break
            }
            log.Printf("Error al recibir solicitud de carga: %v", err)
            return fmt.Errorf("failed to receive upload request: %w", err)
        }

        log.Printf("Recibido chunk del archivo con ID: %s de tamaño %d bytes\n", req.FileId, len(req.BinaryFile))

        // Almacena el fileId
        fileId = req.FileId

        // Subir archivo
        log.Printf("Iniciando carga de archivo con ID: %s al NFS...", fileId)
        _, err = uploadToNFS(req)
        if err != nil {
            log.Printf("Error al subir archivo con ID %s: %v", fileId, err)
            return fmt.Errorf("failed to upload file to NFS: %w", err)
        }
        log.Printf("Chunk del archivo con ID %s cargado exitosamente.", fileId)
    }

    // Respuesta al cliente
    log.Printf("Enviando respuesta de éxito al cliente con FileId: %s", fileId)
    err := stream.SendAndClose(&pb.FileUploadResponse{
        FileId: fileId,
    })
    if err != nil {
        log.Printf("Error al enviar respuesta de carga: %v", err)
        return fmt.Errorf("failed to send upload response: %w", err)
    }

    log.Println("Archivo subido exitosamente.")
    return nil
}

func uploadToNFS(req *pb.FileUploadRequest) (string, error) {
    log.Printf("Verificando si el directorio del usuario '%s' existe...", req.OwnerId)
    
    userPath := fmt.Sprintf("./nfs/files/%s", req.OwnerId)
    if _, err := os.Stat(userPath); os.IsNotExist(err) {
        log.Printf("Directorio no existe. Creando directorio para el usuario '%s'...", req.OwnerId)
        err := os.Mkdir(userPath, 0755)
        if err != nil {
            log.Printf("Error al crear el directorio del usuario '%s': %v", req.OwnerId, err)
            return "", fmt.Errorf("failed to create user directory: %w", err)
        }
        log.Printf("Directorio creado exitosamente para el usuario '%s'.", req.OwnerId)
    } else {
        log.Printf("Directorio para el usuario '%s' ya existe.", req.OwnerId)
    }

    // Extraer la extensión y el nombre del archivo
    fileExtension := filepath.Ext(req.FileName)
    fileName := req.FileId + fileExtension
    log.Printf("Preparando para guardar el archivo como '%s' en '%s'...", fileName, userPath)

    // Guardar el archivo en la ruta principal
    filePath := filepath.Join(userPath, fileName)
    err := saveFile(filePath, req.BinaryFile)
    if err != nil {
        log.Printf("Error al guardar el archivo en '%s': %v", filePath, err)
        return "", fmt.Errorf("failed to upload file: %w", err)
    }

    log.Printf("Archivo guardado exitosamente en '%s'.", filePath)
    return filePath, nil
}

func saveFile(filePath string, binaryFile []byte) error {
    log.Printf("Creando el archivo en '%s'...", filePath)
    fileUpload, err := os.Create(filePath)
    if err != nil {
        log.Printf("Error al crear el archivo: %v", err)
        return fmt.Errorf("failed to create file: %w", err)
    }
    defer func() {
        if err := fileUpload.Close(); err != nil {
            log.Printf("Error al cerrar el archivo: %v", err)
        }
        log.Println("Archivo cerrado correctamente.")
    }()

    // Decodificar el contenido del binario que viene en base64
    log.Println("Decodificando el contenido del archivo...")
    decodedContent, err := base64.StdEncoding.DecodeString(string(binaryFile))
    if err != nil {
        log.Printf("Error al decodificar el contenido binario: %v", err)
        return fmt.Errorf("failed to decode binary content: %w", err)
    }
    log.Printf("Contenido decodificado. Longitud: %d bytes.", len(decodedContent))

    // Escribir el contenido decodificado en el archivo
    if _, err = fileUpload.Write(decodedContent); err != nil {
        log.Printf("Error al escribir contenido binario en el archivo: %v", err)
        return fmt.Errorf("failed to write binary content to file: %w", err)
    }

    log.Println("Contenido escrito en el archivo correctamente.")
    return nil
}
