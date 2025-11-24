# Exemplos de Uso da API

Este arquivo contém exemplos práticos de como usar a API da blockchain de faltas.

## Pré-requisitos

Certifique-se de que os serviços estão rodando:
```bash
docker-compose up -d
```

## 1. Professor Registra Presenças/Faltas

### Registrar uma presença
```bash
curl -X POST http://localhost:5001/presencas \
  -H "Content-Type: application/json" \
  -d '{
    "aluno_id": "123",
    "aula_id": "AULA-001",
    "status": "presente"
  }'
```

### Registrar uma falta
```bash
curl -X POST http://localhost:5001/presencas \
  -H "Content-Type: application/json" \
  -d '{
    "aluno_id": "123",
    "aula_id": "AULA-002",
    "status": "ausente"
  }'
```

### Registrar múltiplas presenças
```bash
# Aluno 123 presente
curl -X POST http://localhost:5001/presencas \
  -H "Content-Type: application/json" \
  -d '{"aluno_id": "123", "aula_id": "AULA-003", "status": "presente"}'

# Aluno 456 ausente
curl -X POST http://localhost:5001/presencas \
  -H "Content-Type: application/json" \
  -d '{"aluno_id": "456", "aula_id": "AULA-003", "status": "ausente"}'

# Aluno 789 presente
curl -X POST http://localhost:5001/presencas \
  -H "Content-Type: application/json" \
  -d '{"aluno_id": "789", "aula_id": "AULA-003", "status": "presente"}'
```

## 2. Professor Minera Blocos

Após registrar várias transações, o professor pode minerar um bloco:

```bash
curl -X POST http://localhost:5001/blocos
```

**Resposta esperada:**
```json
{
  "mensagem": "Bloco minerado com sucesso",
  "bloco": {
    "index": 2,
    "timestamp": 1234567890,
    "transactions": [...],
    "prev_hash": "...",
    "hash": "..."
  },
  "total_transacoes": 3
}
```

## 3. Aluno Consulta Seu Histórico

### Consultar faltas (aluno com ID "1")
```bash
curl http://localhost:5002/alunos/1/faltas
```

**Resposta:**
```json
{
  "aluno_id": "1",
  "registros": [
    {
      "aluno_id": "1",
      "aula_id": "AULA-001",
      "status": "presente",
      "registrado_por": "PROFESSOR-1",
      "timestamp": 1234567890
    },
    ...
  ]
}
```

### Tentar consultar outro aluno (será bloqueado)
```bash
curl http://localhost:5002/alunos/123/faltas
```

**Resposta (403 Forbidden):**
```json
Você só pode consultar seu próprio histórico
```

## 4. DAE Adiciona Justificativas

### Adicionar justificativa para uma falta
```bash
curl -X POST http://localhost:5003/justificativas \
  -H "Content-Type: application/json" \
  -d '{
    "aluno_id": "123",
    "aula_id": "AULA-002",
    "justificativa": "Atestado médico válido - gripe"
  }'
```

### Minerar bloco com justificativa
```bash
curl -X POST http://localhost:5003/blocos
```

## 5. DAE Consulta Dados

### Consultar histórico de um aluno específico
```bash
curl http://localhost:5003/alunos/123/faltas
```

### Consultar histórico de TODOS os alunos
```bash
curl http://localhost:5003/alunos
```

**Resposta:**
```json
{
  "total_alunos": 3,
  "alunos": [
    {
      "aluno_id": "123",
      "registros": [...]
    },
    {
      "aluno_id": "456",
      "registros": [...]
    },
    {
      "aluno_id": "789",
      "registros": [...]
    }
  ]
}
```

## 6. Consultar Blockchain Completa

### Professor vê toda a cadeia
```bash
curl http://localhost:5001/chain
```

### DAE vê toda a cadeia
```bash
curl http://localhost:5003/chain
```

### Aluno vê apenas seus blocos
```bash
curl http://localhost:5002/chain
```

## Fluxo Completo de Exemplo

### 1. Professor registra várias presenças/faltas
```bash
# Aula 1
curl -X POST http://localhost:5001/presencas -H "Content-Type: application/json" -d '{"aluno_id": "123", "aula_id": "AULA-001", "status": "presente"}'
curl -X POST http://localhost:5001/presencas -H "Content-Type: application/json" -d '{"aluno_id": "456", "aula_id": "AULA-001", "status": "ausente"}'

# Aula 2
curl -X POST http://localhost:5001/presencas -H "Content-Type: application/json" -d '{"aluno_id": "123", "aula_id": "AULA-002", "status": "presente"}'
curl -X POST http://localhost:5001/presencas -H "Content-Type: application/json" -d '{"aluno_id": "456", "aula_id": "AULA-002", "status": "ausente"}'
```

### 2. Professor minera o bloco
```bash
curl -X POST http://localhost:5001/blocos
```

### 3. Aluno consulta seu histórico
```bash
curl http://localhost:5002/alunos/1/faltas
```

### 4. DAE verifica todas as faltas
```bash
curl http://localhost:5003/alunos
```

### 5. DAE adiciona justificativa
```bash
curl -X POST http://localhost:5003/justificativas -H "Content-Type: application/json" -d '{"aluno_id": "456", "aula_id": "AULA-001", "justificativa": "Atestado médico"}'
```

### 6. DAE minera bloco com justificativa
```bash
curl -X POST http://localhost:5003/blocos
```

### 7. Verificar blockchain completa
```bash
curl http://localhost:5003/chain | jq
```

## Usando com PowerShell (Windows)

### Registrar presença
```powershell
Invoke-RestMethod -Uri "http://localhost:5001/presencas" -Method POST -ContentType "application/json" -Body '{"aluno_id": "123", "aula_id": "AULA-001", "status": "presente"}'
```

### Consultar faltas
```powershell
Invoke-RestMethod -Uri "http://localhost:5002/alunos/1/faltas" -Method GET
```

## Notas

- Cada nó mantém sua própria cópia da blockchain em memória
- As transações ficam pendentes até serem mineradas em um bloco
- Após mineração, as transações são adicionadas permanentemente à blockchain
- O hash de cada bloco garante a integridade da cadeia


