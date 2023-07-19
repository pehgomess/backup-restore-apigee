package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/apigee/v1"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v2"
)

type CustomAttribute struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

type Credential struct {
	APIProducts []struct {
		APIProduct string `json:"apiproduct" yaml:"apiproduct"`
		Status     string `json:"status" yaml:"status"`
	} `json:"apiProducts" yaml:"apiProducts"`
	ConsumerKey    string `json:"consumerKey" yaml:"consumerKey"`
	ConsumerSecret string `json:"consumerSecret" yaml:"consumerSecret"`
	ExpiresAt      int64  `json:"expiresAt" yaml:"expiresAt"`
	IssuedAt       int64  `json:"issuedAt" yaml:"issuedAt"`
	Status         string `json:"status" yaml:"status"`
}

type AppBackup struct {
	AppID      string            `json:"appId" yaml:"appId"`
	Attributes []CustomAttribute `json:"attributes" yaml:"attributes"`
	//Attributes     map[string]string `json:"attributes" yaml:"attributes"`
	CreatedAt      int64        `json:"createdAt" yaml:"createdAt"`
	Credentials    []Credential `json:"credentials" yaml:"credentials"`
	DeveloperID    string       `json:"developerId" yaml:"developerId"`
	LastModifiedAt int64        `json:"lastModifiedAt" yaml:"lastModifiedAt"`
	Name           string       `json:"name" yaml:"name"`
	Status         string       `json:"status" yaml:"status"`
	AppFamily      string       `json:"appFamily" yaml:"appFamily"`
}

type Config struct {
	ServiceAccountFile string
	Organization       string
	BackupFile         string
}

func help() {
	fmt.Println("Usage: go run main.go <serviceAccountFile> <organization> <backupFile>")
	fmt.Println("\nDescription: Este programa faz o restore de um App do Apigee a partir de um arquivo de backup no formato YAML.")
	fmt.Println("\n- Options: <serviceAccountFile> - Arquivo json do service account")
	fmt.Println("- Options: <organization> - Organizacao Apigee")
	fmt.Println("- Options: <backupFile> - Arquivo de backup no formato YAML")
	fmt.Println("\nEx: go run main.go service-account.json my-org backup.yaml")
}

func main() {
	if len(os.Args) < 4 {
		help()
		return
	}

	config := Config{
		ServiceAccountFile: os.Args[1],
		Organization:       os.Args[2],
		BackupFile:         os.Args[3],
	}

	ctx := context.Background()

	serviceAccountJSON, err := os.ReadFile(config.ServiceAccountFile)
	if err != nil {
		log.Fatalf("Erro ao carregar as credenciais de Service Account %v", err)
	}

	token, err := google.CredentialsFromJSON(ctx, serviceAccountJSON, apigee.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Erro ao carregar as credenciais da Service Account: %v", err)
	}

	service, err := apigee.NewService(ctx, option.WithCredentials(token))
	if err != nil {
		log.Fatalf("Erro ao criar o cliente do Apigee: %v", err)
	}

	httpClient := oauth2.NewClient(ctx, token.TokenSource)

	data, err := os.ReadFile(config.BackupFile)
	if err != nil {
		log.Fatalf("Erro ao ler o arquivo de backup: %v", err)
	}

	var appBackup AppBackup
	err = yaml.Unmarshal(data, &appBackup)
	if err != nil {
		log.Fatalf("Erro ao fazer a desserializacao do arquivo de backup: %v", err)
	}

	credentials := appBackup.Credentials
	if len(credentials) == 0 {
		log.Fatal("Nenhuma credencial encontrada no arquivo de backup")
	}

	err = createApp(service, appBackup, config)
	if err != nil {
		log.Fatalf("Erro ao criar o aplicativo %s: %v", appBackup.Name, err)
	}

	err = createConsumerKeys(httpClient, service, appBackup.DeveloperID, config.Organization, appBackup.Name, credentials)
	if err != nil {
		log.Fatalf("Erro ao criar as chaves do aplicativo para o App %s: %v", appBackup.Name, err)
	}

	fmt.Println("App restaurado com sucesso!")
}

func createApp(client *apigee.Service, appBackup AppBackup, config Config) error {
	app := &apigee.GoogleCloudApigeeV1DeveloperApp{
		Name:       appBackup.Name,
		Attributes: convertAttributes(appBackup.Attributes),
	}

	createAppCall := client.Organizations.Developers.Apps.Create("organizations/"+config.Organization+"/developers/"+appBackup.DeveloperID, app)
	_, err := createAppCall.Do()
	if err != nil {
		return fmt.Errorf("erro ao criar o app: %v", err)
	}

	return nil
}

func convertAttributes(attributes []CustomAttribute) []*apigee.GoogleCloudApigeeV1Attribute {
	apiAttributes := make([]*apigee.GoogleCloudApigeeV1Attribute, 0, len(attributes))
	for _, attr := range attributes {
		apiAttributes = append(apiAttributes, &apigee.GoogleCloudApigeeV1Attribute{
			Name:  attr.Name,
			Value: attr.Value,
		})
	}
	return apiAttributes
}

func createConsumerKeys(httpClient *http.Client, client *apigee.Service, developerID, org, appName string, credentials []Credential) error {
	if len(credentials) == 0 {
		return fmt.Errorf("nenhuma credencial encontrada no arquivo de backup")
	}

	for _, credential := range credentials {
		consumerKey := credential.ConsumerKey
		consumerSecret := credential.ConsumerSecret

		req := &apigee.GoogleCloudApigeeV1DeveloperAppKey{
			ConsumerKey:    consumerKey,
			ConsumerSecret: consumerSecret,
		}

		keyCreateCall := client.Organizations.Developers.Apps.Keys.Create("organizations/"+org+"/developers/"+developerID+"/apps/"+appName, req)
		keyCreateCall.Context(context.Background())

		key, err := keyCreateCall.Do()
		if err != nil {
			log.Printf("Erro ao criar a chave para o App: %v", err)
			continue
		}

		fmt.Printf("Chave do app %s criada: %s\n", appName, key.ConsumerKey)

		err = associateKeyToProduct(httpClient, client, org, developerID, appName, key.ConsumerKey, credential.APIProducts[0].APIProduct)
		if err != nil {
			log.Printf("Erro ao associar a chave ao produto para o App: %v", err)
		}

	}

	return nil
}

func associateKeyToProduct(httpClient *http.Client, client *apigee.Service, org, developerID, appName, consumerKey, productID string) error {
	url := fmt.Sprintf("https://apigee.googleapis.com/v1/organizations/%s/developers/%s/apps/%s/keys/%s", org, developerID, appName, consumerKey)

	requestBody := fmt.Sprintf(`{"apiProducts": ["%s"]}`, productID)

	req, err := http.NewRequest("POST", url, strings.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("erro ao criar a requisição HTTP: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao fazer a requisição HTTP: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("erro ao ler o corpo da resposta HTTP: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("erro ao associar a chave ao produto: %s", string(body))
	}

	return nil
}
