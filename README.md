# Blockchain de Faltas - Sistema Distribuído

Sistema de blockchain privada e permissionada para registrar presenças e faltas de alunos durante o semestre. Implementado em Go com Docker, utilizando múltiplos nós com funções específicas.

## Arquitetura

O sistema possui três tipos de nós, cada um com permissões específicas:

### 🔵 Professor
- **Pode**: Registrar presenças e faltas
- **Pode**: Minerar novos blocos (adicionar transações pendentes à blockchain)
- **Porta**: 5001

### 🟢 DAE (Secretaria)
- **Pode**: Consultar toda a cadeia de blocos
- **Pode**: Verificar histórico de qualquer aluno
- **Pode**: Adicionar justificativas de faltas
- **Pode**: Minerar blocos com justificativas
- **Porta**: 5003

### 🟡 Aluno
- **Pode**: Consultar apenas seu próprio histórico de frequência
- **Porta**: 5002

## Estrutura do Projeto

```
.
├── cmd/
│   └── node/
│       └── main.go          # Ponto de entrada da aplicação
├── internal/
│   ├── api/
│   │   └── http.go          # Endpoints HTTP e lógica de permissões
│   ├── blockchain/
│   │   └── blockchain.go    # Estrutura e lógica da blockchain
│   ├── node/
│   │   └── node.go          # Definição de nós e roles
│   └── config/
│       └── config.go        # Configurações
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

## Como Executar

### Pré-requisitos
- Docker e Docker Compose instalados
- Go 1.23+ (para desenvolvimento local)

### Executar com Docker Compose

```bash
# Construir e iniciar todos os nós
docker-compose up --build

# Executar em background
docker-compose up -d --build

# Ver logs
docker-compose logs -f

# Parar os serviços
docker-compose down
```

Os serviços estarão disponíveis em:
- Professor: http://localhost:5001
- Aluno: http://localhost:5002
- DAE: http://localhost:5003

## Endpoints da API

### Endpoints Comuns (todos os nós)

#### `GET /chain`
Retorna a blockchain completa (filtrada por permissões):
- **Professor/DAE**: Toda a cadeia
- **Aluno**: Apenas blocos com suas transações

**Resposta:**
```json
{
  "node_id": "PROFESSOR-1",
  "role": "PROFESSOR",
  "chain": [...]
}
```

#### `GET /alunos/{id}/faltas`
Consulta faltas de um aluno específico:
- **Aluno**: Só pode consultar seu próprio ID
- **DAE**: Pode consultar qualquer aluno

**Resposta:**
```json
{
  "aluno_id": "123",
  "registros": [...]
}
```

### Endpoints do Professor

#### `POST /presencas`
Registra uma presença ou falta.

**Body:**
```json
{
  "aluno_id": "123",
  "aula_id": "AULA-001",
  "status": "presente"  // ou "ausente"
}
```

#### `POST /blocos`
Mina um novo bloco com todas as transações pendentes.

**Resposta:**
```json
{
  "index": 2,
  "timestamp": 1234567890,
  "transactions": [...],
  "prev_hash": "...",
  "hash": "..."
}
```

### Endpoints do DAE

#### `POST /justificativas`
Adiciona uma justificativa para uma falta.

**Body:**
```json
{
  "aluno_id": "123",
  "aula_id": "AULA-001",
  "justificativa": "Atestado médico"
}
```

#### `POST /blocos`
Mina um novo bloco (mesmo endpoint do professor).

#### `GET /alunos`
Retorna histórico completo de todos os alunos.

**Resposta:**
```json
{
  "total_alunos": 5,
  "alunos": [
    {
      "aluno_id": "123",
      "registros": [...]
    },
    ...
  ]
}
```

## Exemplos de Uso

### 1. Professor registra uma falta
```bash
curl -X POST http://localhost:5001/presencas \
  -H "Content-Type: application/json" \
  -d '{
    "aluno_id": "123",
    "aula_id": "AULA-001",
    "status": "ausente"
  }'
```

### 2. Professor minera um bloco
```bash
curl -X POST http://localhost:5001/blocos
```

### 3. Aluno consulta seu histórico
```bash
curl http://localhost:5002/alunos/1/faltas
```

### 4. DAE adiciona justificativa
```bash
curl -X POST http://localhost:5003/justificativas \
  -H "Content-Type: application/json" \
  -d '{
    "aluno_id": "123",
    "aula_id": "AULA-001",
    "justificativa": "Atestado médico válido"
  }'
```

### 5. DAE consulta todos os alunos
```bash
curl http://localhost:5003/alunos
```

## Estrutura de Dados

### Transaction
```go
{
  "aluno_id": "123",
  "aula_id": "AULA-001",
  "status": "presente|ausente|justificada",
  "registrado_por": "PROFESSOR-1",
  "timestamp": 1234567890,
  "justificativa": "..." // opcional
}
```

### Block
```go
{
  "index": 1,
  "timestamp": 1234567890,
  "transactions": [...],
  "prev_hash": "...",
  "hash": "..."
}
```

## Características da Blockchain

- **Imutabilidade**: Blocos uma vez adicionados não podem ser alterados
- **Integridade**: Cada bloco contém hash do bloco anterior
- **Transparência**: DAE e Professor podem auditar toda a cadeia
- **Privacidade**: Alunos só veem seus próprios dados
- **Rastreabilidade**: Todas as transações registram quem as criou

## Desenvolvimento

### Executar localmente (sem Docker)

```bash
# Instalar dependências
go mod download

# Executar nó do professor
NODE_ID=PROFESSOR-1 NODE_ROLE=PROFESSOR PORT=8080 go run ./cmd/node

# Executar nó do aluno
NODE_ID=ALUNO-1 NODE_ROLE=ALUNO PORT=8081 go run ./cmd/node

# Executar nó do DAE
NODE_ID=DAE-1 NODE_ROLE=DAE PORT=8082 go run ./cmd/node
```

### Compilar

```bash
go build -o node ./cmd/node
```

## Notas Importantes

- Cada nó mantém sua própria cópia da blockchain em memória
- Para sincronização entre nós em produção, seria necessário implementar comunicação P2P
- O sistema atual é adequado para demonstração e aprendizado
- Em produção, considere adicionar persistência em banco de dados

## Licença

Este projeto é para fins educacionais.

