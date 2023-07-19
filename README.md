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

O que faz ? 

- Faz o restore do Apps do arquivo yaml gerado no backup
- Faz o restore do consumerKey e consumerSecret com os seus respectivos products do arquivo yaml gerado no backup
- Faz o restore dos custom attributes e apps do arquivo yaml gerado no backup

O que falta fazer ?

- lista dos arquivos de um diretorio e nao apenas de um arquivo
- validar e se nao existir criar o developer com base no arquivo de app do yaml


Informacoes uteis.

* Ja existe o restore do developer entao se nao existir o developer precisa criar inicialmente para depois criar o restore do apps.

* Inicialmente precisa apontar o arquivo do yaml 

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


