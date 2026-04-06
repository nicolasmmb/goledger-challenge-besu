# API (Go)

## Pré-requisitos

- Docker + Docker Compose
- Go 1.22+
- `make`

## Subir tudo

1. Na raiz do repositório, suba a rede e faça o deploy do contrato:

```bash
make devnet-deploy
```

2. Entre na pasta da aplicação:

```bash
cd app
```

3. Crie o arquivo de ambiente:

```bash
cp .env.example .env
```

4. Atualize no `.env` os campos do contrato/deploy:

- `CONTRACT_ADDRESS` (endereço mostrado no `make devnet-deploy`)
- `CHAIN_ID` (rede local padrão: `1337`)

5. Suba o PostgreSQL:

```bash
make docker-up
```

6. Aplique as migrações:

```bash
make migrate-up
```

7. Inicie a API localmente:

```bash
make run
```

## Validar execução

```bash
curl -s http://localhost:8080/healthz
curl -s http://localhost:8080/readyz
```

## Encerrar tudo

1. Na pasta `app`, pare o banco e apague os dados:

```bash
make docker-down
```

2. Na raiz, pare a rede:

```bash
make stop-devnet
```

## Comandos úteis (app)

```bash
make test
make lint
make test-integration
```
