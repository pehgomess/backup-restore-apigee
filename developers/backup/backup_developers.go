package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/apigee/v1"
	"google.golang.org/api/option"
)

type DeveloperBackup struct {
	Email            string   `json:"email"`
	FirstName        string   `json:"firstName"`
	LastName         string   `json:"lastName"`
	UserName         string   `json:"userName"`
	Apps             []string `json:"apps"`
	DeveloperID      string   `json:"developerId"`
	OrganizationName string   `json:"organizationName"`
	Status           string   `json:"status"`
	CreatedAt        int64    `json:"createdAt"`
	LastModifiedAt   int64    `json:"lastModifiedAt"`
}

func help() {
	fmt.Println("Usage: go run main.go <serviceAccountFile> <organization> <backupDir>")
	fmt.Println("\nDescription: Este programa faz backup de todos os developers do Apigee.")
	fmt.Println("\n- Options: <serviceAccountFile> - Arquivo json do service account")
	fmt.Println("- Options: <organization> - Organizacao Apigee")
	fmt.Println("- Options: <backupDir> - Diretorio que deseja criar. OBS: O script cria no final do diretorio _dia-mes-ano_hora_min_segundos")
	fmt.Println("\nEx: go run main.go service-account.json my-org backups")
}

func main() {

	if len(os.Args) < 4 {
		help()
		return

	}

	serviceAccountFile := os.Args[1]
	org := os.Args[2]
	backupDir := os.Args[3]

	//timestamp := time.Now().Unix()
	timestamp := time.Now().Format("02-01-2006_15-04-05")

	//dirBackup := backupDir + "_" + strconv.FormatInt(timestamp, 10)
	dirBackup := backupDir + "_" + string(timestamp)

	err := os.Mkdir(dirBackup, 0755)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Diretorio '%s' criado com sucesso.", backupDir)

	//serviceAccountFile := "../../credentials/admapi.json"

	ctx := context.Background()

	serviceAccountJSON, err := os.ReadFile(serviceAccountFile)
	if err != nil {
		log.Fatalf("Erro ao carregar as credenciais de Service Account %v", err)
	}

	credentials, err := google.CredentialsFromJSON(ctx, serviceAccountJSON, apigee.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Erro ao carregar as credenciais da Service Account: %v", err)
	}

	service, err := apigee.NewService(ctx, option.WithCredentials(credentials))
	if err != nil {
		log.Fatalf("Erro ao criar o cliente do Apigee: %v", err)
	}

	developers, err := service.Organizations.Developers.List("organizations/" + org).Do()
	if err != nil {
		log.Fatalf("Erro ao obter a lista de developers: %v", err)
	}

	// Percorre a lista de apps que pertencem ao developers
	for _, developer := range developers.Developer {
		developerDetails, err := service.Organizations.Developers.Get(fmt.Sprintf("organizations/%s/developers/%s", org, developer.Email)).Do()
		if err != nil {
			log.Fatalf("Erro ao obter a lista de developer %s: %v", developer.Email, err)
			continue
		}

		developerBackup := DeveloperBackup{
			Email:            developerDetails.Email,
			FirstName:        developerDetails.FirstName,
			LastName:         developerDetails.LastName,
			UserName:         developerDetails.UserName,
			Apps:             developerDetails.Apps,
			DeveloperID:      developerDetails.DeveloperId,
			OrganizationName: developerDetails.OrganizationName,
			Status:           developerDetails.Status,
			CreatedAt:        developerDetails.CreatedAt,
			LastModifiedAt:   developerDetails.LastModifiedAt,
		}

		backupData, err := json.MarshalIndent(developerBackup, "", "  ")
		if err != nil {
			log.Printf("Erro ao converter o developer para JSON: %v", err)
			continue
		}

		filename := fmt.Sprintf(dirBackup+"/%s.json", developerBackup.Email)
		err = saveToFile(filename, backupData)
		if err != nil {
			log.Printf("Erro ao salvar o arquivo de backup para o developers %s: %v", developer.Email, err)
		}
	}
}

func saveToFile(filename string, data []byte) error {
	// Abre o arquivo no modo de escrita
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("erro ao abrir o arquivo: %v", err)
	}
	defer file.Close()

	// Escreve os dados no arquivo
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("erro ao escrever no arquivo: %v", err)
	}

	return nil
}
