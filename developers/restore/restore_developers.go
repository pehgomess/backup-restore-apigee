package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/apigee/v1"
	"google.golang.org/api/option"
)

type DeveloperBackup struct {
	Email            string   `json:"email,omitempty"`
	UserName         string   `json:"userName,omitempty"`
	FirstName        string   `json:"firstName,omitempty"`
	LastName         string   `json:"lastName,omitempty"`
	Apps             []string `json:"apps"`
	DeveloperID      string   `json:"developerId"`
	OrganizationName string   `json:"organizationName,omitempty"`
	Status           string   `json:"status"`
	CreatedAt        int64    `json:"createdAt"`
	LastModifiedAt   int64    `json:"lastModifiedAt"`
}

func help() {
	fmt.Println("Usage: go run main.go <serviceAccountFile> <organization> <restoreDir>")
	fmt.Println("\nDescription: Este programa faz backup de todos os developers do Apigee.")
	fmt.Println("\n- Options: <serviceAccountFile> - Arquivo json do service account")
	fmt.Println("- Options: <organization> - Organizacao Apigee")
	fmt.Println("- Options: <restoreDir> - Diretorio que contem os *json dos apps, OBS: O script lista todos os *.json do diretorio e cria 1 a 1.")
	fmt.Println("\nEx: go run main.go service-account.json my-org restoreDir")
}

func main() {

	if len(os.Args) < 4 {
		help()
		return

	}

	serviceAccountFile := os.Args[1]
	org := os.Args[2]
	backupDir := os.Args[3]

	ctx := context.Background()

	serviceAccountJSON, err := os.ReadFile(serviceAccountFile)
	if err != nil {
		log.Fatalf("Error reading service account file: %v", err)
	}

	credentials, err := google.CredentialsFromJSON(ctx, serviceAccountJSON, apigee.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Error creating credentials: %v", err)
	}

	service, err := apigee.NewService(ctx, option.WithCredentials(credentials))
	if err != nil {
		log.Fatalf("Error creating Apigee service: %v", err)
	}

	// Listar os arquivos de backup
	backupFiles, err := filepath.Glob(filepath.Join(backupDir, "*.json"))
	if err != nil {
		log.Fatalf("Error listing backup files: %v", err)
	}

	// Percorrer os arquivos de backup
	for _, backupFile := range backupFiles {
		// Abrir o arquivo de backup
		file, err := os.Open(backupFile)
		if err != nil {
			log.Printf("Error opening backup file %s: %v", backupFile, err)
			continue
		}
		defer file.Close()

		// Decodificar o arquivo JSON em uma estrutura DeveloperBackup
		var backup DeveloperBackup
		err = json.NewDecoder(file).Decode(&backup)
		if err != nil {
			log.Printf("Error decoding backup file %s: %v", backupFile, err)
			continue
		}

		// Criar ou atualizar o registro do developer no Apigee
		developer := &apigee.GoogleCloudApigeeV1Developer{
			Email:            backup.Email,
			UserName:         backup.UserName,
			FirstName:        backup.FirstName,
			LastName:         backup.LastName,
			Apps:             backup.Apps,
			DeveloperId:      backup.DeveloperID,
			OrganizationName: backup.OrganizationName,
			Status:           backup.Status,
			CreatedAt:        backup.CreatedAt,
			LastModifiedAt:   backup.LastModifiedAt,
		}

		// Imprimir o JSON do objeto developer
		jsonData, err := json.MarshalIndent(developer, "", "  ")
		if err != nil {
			log.Printf("Error marshaling developer to JSON: %v", err)
			continue
		}

		fmt.Println(string(jsonData))

		_, err = service.Organizations.Developers.Create("organizations/"+org, developer).Do()
		if err != nil {
			log.Printf("Error creating or updating developer %s: %v", backup.Email, err)
			continue
		}

		fmt.Printf("Restored developer: %s\n", backup.Email)
	}
}
