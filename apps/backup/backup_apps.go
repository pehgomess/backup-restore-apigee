package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/apigee/v1"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v2"
)

type Attribute struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type Credential struct {
	APIProducts []struct {
		APIProduct string `yaml:"apiproduct"`
		Status     string `yaml:"status"`
	} `yaml:"apiProducts"`
	ConsumerKey    string `yaml:"consumerKey"`
	ConsumerSecret string `yaml:"consumerSecret"`
	ExpiresAt      int64  `yaml:"expiresAt"`
	IssuedAt       int64  `yaml:"issuedAt"`
	Status         string `yaml:"status"`
}

type AppBackup struct {
	AppID          string       `yaml:"appId"`
	Attributes     []Attribute  `yaml:"attributes"`
	CreatedAt      int64        `yaml:"createdAt"`
	Credentials    []Credential `yaml:"credentials"`
	DeveloperID    string       `yaml:"developerId"`
	LastModifiedAt int64        `yaml:"lastModifiedAt"`
	Name           string       `yaml:"name"`
	Status         string       `yaml:"status"`
	AppFamily      string       `yaml:"appFamily"`
}

func help() {
	fmt.Println("Usage: go run main.go <serviceAccountFile> <organization> <backupDir>")
	fmt.Println("\nDescription: Este programa faz backup de todos os Apps do Apigee.")
	fmt.Println("\n- Options: <serviceAccountFile> - Arquivo json do service account")
	fmt.Println("- Options: <organization> - Organizacao Apigee")
	fmt.Println("- Options: <backupDir> - Diretorio que deseja criar. OBS: O script cria no final do diretorio _timestamp")
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

	timestamp := time.Now().Format("02-01-2006_15-04-05")
	dirBackup := backupDir + "_" + string(timestamp)

	err := os.Mkdir(dirBackup, 0755)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Diretorio '%s' criado com sucesso.", backupDir)

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

	var numApps int

	for _, developer := range developers.Developer {
		apps, err := service.Organizations.Developers.Apps.List("organizations/" + org + "/developers/" + developer.Email).Do()
		if err != nil {
			log.Printf("Erro ao obter a lista de Apps do developer %s: %v", developer.Email, err)
			continue
		}

		for _, app := range apps.App {
			appDetails, err := service.Organizations.Developers.Apps.Get("organizations/" + org + "/developers/" + developer.Email + "/apps/" + app.AppId).Do()
			if err != nil {
				log.Printf("Erro ao obter os detalhes do App: %v", err)
				continue
			}
			numApps++

			var attributes []Attribute
			for _, attr := range appDetails.Attributes {
				attributes = append(attributes, Attribute{
					Name:  attr.Name,
					Value: attr.Value,
				})
			}

			var credentials []Credential
			for _, cred := range appDetails.Credentials {
				apiProducts := make([]struct {
					APIProduct string `yaml:"apiproduct"`
					Status     string `yaml:"status"`
				}, len(cred.ApiProducts))

				for i, product := range cred.ApiProducts {
					apiProducts[i] = struct {
						APIProduct string `yaml:"apiproduct"`
						Status     string `yaml:"status"`
					}{
						APIProduct: product.Apiproduct,
						Status:     product.Status,
					}
				}

				credentials = append(credentials, Credential{
					APIProducts:    apiProducts,
					ConsumerKey:    cred.ConsumerKey,
					ConsumerSecret: cred.ConsumerSecret,
					ExpiresAt:      cred.ExpiresAt,
					IssuedAt:       cred.IssuedAt,
					Status:         cred.Status,
				})
			}

			appBackup := AppBackup{
				AppID:          appDetails.AppId,
				Attributes:     attributes,
				CreatedAt:      appDetails.CreatedAt,
				Credentials:    credentials,
				DeveloperID:    developer.Email,
				LastModifiedAt: appDetails.LastModifiedAt,
				Name:           appDetails.Name,
				Status:         appDetails.Status,
				AppFamily:      appDetails.AppFamily,
			}

			yamlData, err := yaml.Marshal(appBackup)
			if err != nil {
				log.Fatalf("Erro ao converter o backup do App em YAML: %v", err)
				continue
			}

			filename := fmt.Sprintf(dirBackup+"/%s.yaml", appDetails.Name)
			err = saveToFile(filename, yamlData)
			if err != nil {
				log.Printf("Erro ao salvar o arquivo YAML: %v", err)
			}
			fmt.Printf(" - Apps consumido: %s\n", appDetails.Name)
		}
	}

	fmt.Printf("Total de Apps: %d\n", numApps)
}

func saveToFile(filename string, data []byte) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("erro ao abrir o arquivo: %v", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("erro ao escrever no arquivo: %v", err)
	}

	return nil
}
