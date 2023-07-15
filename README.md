# Backup & Restore - Apigee X

A ideia simplesmente criar os diretorios dos itens do apigee tais como: 

**_Developers_** \
**_Apps_** \
**_ApiProducts_** \
**_TargetServers_**

## contem o diretorio de backup e restore

O service account precisa ter permissoes no projeto para conseguir efetuar o backup de todos os  recursos como por exemplo key e secret 

No diretorio corrente executar o comando abaixo para criar o pacote \
`go mod init nome_do_pacote` 

Executar os comandos go get nos pacotes da google \
`go get -u golang.org/x/oauth2/google` \
`go get -u google.golang.org/api/apigee/v1` \
`go mod tidy` 

----------------------------------------------------------------------------

# Apps (dir)

## Diretorio: Backup

**Para usar o codigo** 

```sh
Usage: go run backup_apps.go <serviceAccountFile> <organization> <backupDir> 

Description: Este programa faz backup de todos os Apps do Apigee. 

- Options: <serviceAccountFile> - Arquivo json do service account \
- Options: <organization> - Organizacao Apigee \
- Options: <backupDir> - Diretorio que deseja criar. OBS: O script cria no final do diretorio  _mes-dia-ano_hora-min-sec 

Ex: go run backup_apps.go service-account.json my-org backups 
```

## Diretorio: Restore

**Para usar o codigo**

em desenvolvimento

Faz o restore do consumerKey e consumerSecret com os seus respectivos products

* Ja existe o restore do developer, precisa primeiro restaurar o developer para depois restaurar as Keys, vou adicionar todos na proximo futuramente

Inicialmente precisa apontar o arquivo do yaml 

**OBS: Falta terminar o restore dos custom attributes, developers e apps.

----------------------------------------------------------------------------

# Developers

## Diretorio: Backup

**Para usar o codigo**

```sh
Usage: go run backup_developers.go <serviceAccountFile> <organization> <backupDir> 

Description: Este programa faz backup de todos os Apps do Apigee. 

- Options: <serviceAccountFile> - Arquivo json do service account \
- Options: <organization> - Organizacao Apigee \
- Options: <backupDir> - Diretorio que deseja criar. OBS: O script cria no final do diretorio  _mes-dia-ano_hora-min-sec 

Ex: go run backup_developers.go service-account.json my-org backups 
```

## Diretorio: Restore

**Para usar o codigo** 

em fase de teste

----------------------------------------------------------------------------


## Testes 

- O restore do app faz toda a parte do consumerKey, consumerSecret e apiproducts.


