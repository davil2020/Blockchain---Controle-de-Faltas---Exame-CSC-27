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

## 📍 Checkpoint - Estado Atual do Projeto

### Comportamento Atual

**⚠️ IMPORTANTE**: Atualmente, cada nó mantém sua própria blockchain **independente** em memória. Não há sincronização entre os nós. Isso significa que:

- Operações realizadas no nó Professor afetam **APENAS** a blockchain do Professor
- Operações realizadas no nó DAE afetam **APENAS** a blockchain do DAE
- O nó Aluno tem **apenas leitura** e não pode modificar sua blockchain

### Operações Disponíveis e Impacto

#### 1️⃣ Professor Registra Presença/Falta

**Endpoint**: `POST /presencas` (porta 5001)

**Exemplo**:
```bash
curl -X POST http://localhost:5001/presencas \
  -H "Content-Type: application/json" \
  -d '{"aluno_id": "123", "aula_id": "AULA-001", "status": "presente"}'
```

**O que acontece**:
- ✅ Transação é adicionada ao `PendingTransactions` do **nó Professor**
- ❌ Blockchain do Professor **NÃO** é atualizada ainda (transação fica pendente)
- ❌ Blockchains do Aluno e DAE **NÃO** são afetadas

**Estado das Blockchains**:
- 🔵 **Professor**: Transação pendente (não minerada)
- 🟡 **Aluno**: Sem alterações
- 🟢 **DAE**: Sem alterações

---

#### 2️⃣ Professor Minera Bloco

**Endpoint**: `POST /blocos` (porta 5001)

**Exemplo**:
```bash
curl -X POST http://localhost:5001/blocos
```

**O que acontece**:
- ✅ Todas as transações pendentes são mineradas em um **novo bloco**
- ✅ Blockchain do **Professor** é atualizada (novo bloco adicionado)
- ✅ `PendingTransactions` do Professor é **limpo**
- ✅ Integridade da blockchain é **verificada automaticamente**
- ❌ Blockchains do Aluno e DAE **NÃO** são afetadas

**Estado das Blockchains**:
- 🔵 **Professor**: Novo bloco adicionado com transações
- 🟡 **Aluno**: Sem alterações
- 🟢 **DAE**: Sem alterações

**Retorna**: Informações do bloco minerado incluindo hash e total de transações

---

#### 3️⃣ DAE Adiciona Justificativa

**Endpoint**: `POST /justificativas` (porta 5003)

**Exemplo**:
```bash
curl -X POST http://localhost:5003/justificativas \
  -H "Content-Type: application/json" \
  -d '{"aluno_id": "123", "aula_id": "AULA-001", "justificativa": "Atestado médico"}'
```

**O que acontece**:
- ✅ Transação com status "justificada" é adicionada ao `PendingTransactions` do **nó DAE**
- ❌ Blockchain do DAE **NÃO** é atualizada ainda (transação fica pendente)
- ❌ Blockchains do Professor e Aluno **NÃO** são afetadas
- ⚠️ **Nota**: A justificativa é criada independentemente de existir uma falta prévia

**Estado das Blockchains**:
- 🔵 **Professor**: Sem alterações
- 🟡 **Aluno**: Sem alterações
- 🟢 **DAE**: Transação pendente (não minerada)

---

#### 4️⃣ DAE Minera Bloco

**Endpoint**: `POST /blocos` (porta 5003)

**Exemplo**:
```bash
curl -X POST http://localhost:5003/blocos
```

**O que acontece**:
- ✅ Todas as transações pendentes do DAE são mineradas em um **novo bloco**
- ✅ Blockchain do **DAE** é atualizada (novo bloco adicionado)
- ✅ `PendingTransactions` do DAE é **limpo**
- ✅ Integridade da blockchain é **verificada automaticamente**
- ❌ Blockchains do Professor e Aluno **NÃO** são afetadas

**Estado das Blockchains**:
- 🔵 **Professor**: Sem alterações
- 🟡 **Aluno**: Sem alterações
- 🟢 **DAE**: Novo bloco adicionado com justificativas

---

#### 5️⃣ Consultar Blockchain Completa

**Endpoint**: `GET /chain` (todas as portas)

**Exemplos**:
```bash
curl http://localhost:5001/chain  # Professor - vê toda sua blockchain
curl http://localhost:5002/chain  # Aluno - vê apenas seus dados
curl http://localhost:5003/chain  # DAE - vê toda sua blockchain
```

**O que acontece**:
- ✅ Retorna a blockchain **local** do nó consultado
- ✅ **Professor/DAE**: Veem todos os blocos e transações de sua blockchain
- ✅ **Aluno**: Vê apenas blocos que contêm transações do seu ID
- ❌ **NÃO** há consulta entre nós (cada um retorna sua própria cadeia)

**Estado das Blockchains**: Nenhuma alteração (operação de leitura)

---

#### 6️⃣ Aluno Consulta Suas Faltas

**Endpoint**: `GET /alunos/{id}/faltas` (porta 5002)

**Exemplo**:
```bash
curl http://localhost:5002/alunos/1/faltas
```

**O que acontece**:
- ✅ Busca na blockchain **local do Aluno** todas as transações do ID especificado
- ✅ Sistema de **permissão**: Aluno só pode consultar seu próprio ID
  - ID do aluno é extraído do `NODE_ID` (ex: `ALUNO-1` → ID = `1`)
  - Se tentar consultar outro ID: retorna **403 Forbidden**
- ❌ **NÃO** consulta blockchains de outros nós

**Nota**: Como o aluno não pode minerar, sua blockchain local estará vazia (apenas bloco genesis) a menos que você implemente sincronização.

**Estado das Blockchains**: Nenhuma alteração (operação de leitura)

---

#### 7️⃣ DAE Consulta Faltas de Qualquer Aluno

**Endpoint**: `GET /alunos/{id}/faltas` (porta 5003)

**Exemplo**:
```bash
curl http://localhost:5003/alunos/123/faltas
```

**O que acontece**:
- ✅ Busca na blockchain **local do DAE** todas as transações do aluno especificado
- ✅ **Sem restrição de ID**: DAE pode consultar qualquer aluno
- ✅ Retorna todas as transações (presenças, faltas e justificativas) do aluno
- ❌ **NÃO** consulta blockchains de outros nós

**Estado das Blockchains**: Nenhuma alteração (operação de leitura)

---

#### 8️⃣ DAE Consulta Todos os Alunos

**Endpoint**: `GET /alunos` (porta 5003)

**Exemplo**:
```bash
curl http://localhost:5003/alunos
```

**O que acontece**:
- ✅ Percorre toda a blockchain **local do DAE**
- ✅ Agrupa todas as transações por `aluno_id`
- ✅ Retorna um mapa com todos os alunos e seus respectivos registros
- ❌ **NÃO** consulta blockchains de outros nós

**Estado das Blockchains**: Nenhuma alteração (operação de leitura)

---

### Fluxo Completo de Teste

Para entender o comportamento isolado de cada blockchain:

#### Cenário 1: Professor Registra e Minera

```bash
# 1. Professor registra 2 presenças
curl -X POST http://localhost:5001/presencas -H "Content-Type: application/json" \
  -d '{"aluno_id": "123", "aula_id": "AULA-001", "status": "presente"}'

curl -X POST http://localhost:5001/presencas -H "Content-Type: application/json" \
  -d '{"aluno_id": "456", "aula_id": "AULA-001", "status": "ausente"}'

# 2. Professor minera
curl -X POST http://localhost:5001/blocos

# 3. Verificar blockchains
curl http://localhost:5001/chain  # ✅ Tem 2 blocos (genesis + novo)
curl http://localhost:5002/chain  # ❌ Tem 1 bloco (apenas genesis)
curl http://localhost:5003/chain  # ❌ Tem 1 bloco (apenas genesis)
```

**Resultado**: 
- 🔵 Professor: 2 blocos (genesis + bloco com 2 transações)
- 🟡 Aluno: 1 bloco (genesis)
- 🟢 DAE: 1 bloco (genesis)

---

#### Cenário 2: DAE Adiciona Justificativa

```bash
# 1. DAE adiciona justificativa
curl -X POST http://localhost:5003/justificativas -H "Content-Type: application/json" \
  -d '{"aluno_id": "789", "aula_id": "AULA-002", "justificativa": "Atestado médico"}'

# 2. DAE minera
curl -X POST http://localhost:5003/blocos

# 3. Verificar blockchains
curl http://localhost:5001/chain  # Continua com 2 blocos (do cenário 1)
curl http://localhost:5003/chain  # ✅ Agora tem 2 blocos (genesis + justificativa)
```

**Resultado**:
- 🔵 Professor: 2 blocos (presenças do cenário 1)
- 🟢 DAE: 2 blocos (genesis + justificativa)
- ⚠️ As blockchains do Professor e DAE são **independentes** e contêm dados diferentes

---

#### Cenário 3: Aluno Tenta Consultar

```bash
# Aluno com NODE_ID=ALUNO-1 tenta consultar
curl http://localhost:5002/alunos/1/faltas     # ✅ Permitido (seu próprio ID)
curl http://localhost:5002/alunos/123/faltas   # ❌ 403 Forbidden (ID diferente)

# Como a blockchain do aluno está vazia (não sincronizada):
# Resposta: {"aluno_id":"1","registros":null}
```

---

### Limitações Conhecidas

1. **Sem Sincronização P2P**
   - Cada nó opera de forma independente
   - Transações em um nó não são propagadas para outros
   - Ideal para demonstração, não para produção

2. **Armazenamento em Memória**
   - Blockchain é perdida ao reiniciar o container
   - Não há persistência em banco de dados

3. **Aluno com Blockchain Vazia**
   - O nó Aluno não pode minerar blocos
   - Sem sincronização, ele só terá o bloco genesis
   - Consultas retornarão vazias

4. **Justificativas Independentes**
   - DAE pode criar justificativas sem verificar se existe falta prévia
   - Não há validação cruzada entre blockchains de diferentes nós

---

### Próximos Passos (Sugestões)

Para evoluir o projeto, considere implementar:

1. **Sincronização P2P**: Comunicação entre nós para compartilhar blocos
2. **Consenso**: Algoritmo de consenso (ex: Proof of Work, PBFT)
3. **Persistência**: Salvar blockchain em banco de dados
4. **Validação Cruzada**: Verificar se falta existe antes de justificar
5. **Endpoints de Sincronização**: 
   - `POST /sync` para solicitar blockchain de outro nó
   - `GET /peers` para descobrir outros nós da rede

---

## Notas Importantes

- Cada nó mantém sua própria cópia da blockchain em memória
- Para sincronização entre nós em produção, seria necessário implementar comunicação P2P
- O sistema atual é adequado para demonstração e aprendizado
- Em produção, considere adicionar persistência em banco de dados

## Licença

Este projeto é para fins educacionais!

