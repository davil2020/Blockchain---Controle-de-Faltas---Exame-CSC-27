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
- Python 3 (para os comandos de teste formatados)

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

---

## 🧪 Guia Completo de Testes - Passo a Passo

Este guia demonstra o funcionamento completo do sistema, incluindo sincronização automática entre os nós.

### 📋 Preparação do Ambiente

#### 1. Clone o repositório (se ainda não fez)
```bash
git clone <url-do-repositorio>
cd Blockchain---Controle-de-Faltas---Exame-CSC-27
```

#### 2. Limpe containers e imagens anteriores (começar do zero)
```bash
# Parar e remover containers existentes
docker-compose down

# Remover volumes (limpa dados persistentes, se houver)
docker-compose down -v

# (Opcional) Remover imagens Docker antigas do projeto
docker-compose down --rmi all

# (Opcional) Limpar sistema Docker completo (cuidado: remove TUDO)
# docker system prune -a --volumes
```

**⚠️ Nota:** Execute esses comandos apenas se quiser começar completamente do zero. O comando `docker system prune` é opcional e remove TODOS os recursos Docker não utilizados no sistema.

#### 3. Inicie os containers Docker
```bash
# Construir as imagens e iniciar os serviços em background
docker-compose up -d --build
```

**Esperado:** Docker irá:
1. Construir as imagens Go (pode levar ~30s na primeira vez)
2. Criar a rede `blockchain-network`
3. Iniciar 3 containers (node_professor, node_aluno, node_dae)

#### 4. Verifique se os containers estão rodando
```bash
docker-compose ps
```
**Esperado:** 3 containers rodando (node_professor, node_aluno, node_dae)

#### 4. Verifique os logs de inicialização
```bash
docker-compose logs --tail=20
```
**Esperado:** Cada nó deve mostrar:
- `🚀 Starting node X with role Y`
- `📊 Blockchain initialized with 1 blocks`
- `🔗 Connected to 2 peer(s)`

---

### 🎬 Testes Práticos - 3 Terminais

Abra **3 terminais** lado a lado e identifique:
- **Terminal 1** = Professor (porta 5001)
- **Terminal 2** = DAE (porta 5003)
- **Terminal 3** = Aluno (porta 5002)

---

### ▶️ PASSO 1: Verificar Estado Inicial

#### 📱 Terminal 1 (Professor):
```bash
curl -s http://localhost:5001/chain | python3 -c "import sys,json; d=json.load(sys.stdin); print('PROFESSOR'); print('Blocos:', len(d['chain']))"
```
**Esperado:** `Blocos: 1` (bloco genesis)

#### 📱 Terminal 2 (DAE):
```bash
curl -s http://localhost:5003/chain | python3 -c "import sys,json; d=json.load(sys.stdin); print('DAE'); print('Blocos:', len(d['chain']))"
```
**Esperado:** `Blocos: 1`

#### 📱 Terminal 3 (Aluno):
```bash
curl -s http://localhost:5002/chain | python3 -c "import sys,json; d=json.load(sys.stdin); print('ALUNO'); print('Blocos visiveis:', len(d['chain']) if d['chain'] else 0)"
```
**Esperado:** `Blocos visiveis: 0` ou `1`

---

### ▶️ PASSO 2: Professor Registra Presenças

#### 📱 Terminal 1 (Professor):

```bash
# Registrar aluno 123 como presente
curl -X POST http://localhost:5001/presencas \
  -H "Content-Type: application/json" \
  -d '{"aluno_id": "123", "aula_id": "AULA-001", "status": "presente"}'
```
**Esperado:** `{"mensagem":"Transação adicionada"}`

```bash
# Registrar aluno 456 como ausente
curl -X POST http://localhost:5001/presencas \
  -H "Content-Type: application/json" \
  -d '{"aluno_id": "456", "aula_id": "AULA-001", "status": "ausente"}'
```
**Esperado:** `{"mensagem":"Transação adicionada"}`

```bash
# Registrar aluno 1 como presente
curl -X POST http://localhost:5001/presencas \
  -H "Content-Type: application/json" \
  -d '{"aluno_id": "1", "aula_id": "AULA-001", "status": "presente"}'
```
**Esperado:** `{"mensagem":"Transação adicionada"}`

**✅ 3 transações pendentes adicionadas**

---

### ▶️ PASSO 3: Professor Minera Bloco

#### 📱 Terminal 1 (Professor):

```bash
# Minerar bloco com as transações pendentes
curl -X POST http://localhost:5001/blocos
```
**Esperado:** Mensagem `"Bloco minerado com sucesso"` com detalhes do bloco

```bash
# Aguardar propagação (2 segundos)
sleep 2
```

**🔄 Neste momento, a blockchain é PROPAGADA automaticamente para Aluno e DAE**

---

### ▶️ PASSO 4: Verificar Sincronização (TODOS os nós)

#### 📱 Terminal 1 (Professor):
```bash
curl -s http://localhost:5001/chain | python3 -c "import sys,json; d=json.load(sys.stdin); print('PROFESSOR'); print('  Blocos:', len(d['chain'])); print('  Hash:', d['chain'][-1]['hash'][:16] + '...')"
```
**Esperado:** `Blocos: 2`

#### 📱 Terminal 2 (DAE):
```bash
curl -s http://localhost:5003/chain | python3 -c "import sys,json; d=json.load(sys.stdin); print('DAE'); print('  Blocos:', len(d['chain'])); print('  Hash:', d['chain'][-1]['hash'][:16] + '...')"
```
**Esperado:** `Blocos: 2` ✅ **SINCRONIZADO COM PROFESSOR!**

#### 📱 Terminal 3 (Aluno):
```bash
curl -s http://localhost:5002/chain | python3 -c "import sys,json; d=json.load(sys.stdin); print('ALUNO'); print('  Blocos visiveis:', len(d['chain']) if d['chain'] else 0)"
```
**Esperado:** `Blocos visiveis: 1` (filtra apenas transações do aluno "1")

---

### ▶️ PASSO 5: DAE Consulta Dados (Evidência de Sincronização)

#### 📱 Terminal 2 (DAE):

```bash
# Consultar todos os alunos registrados
curl -s http://localhost:5003/alunos | python3 -c "import sys,json; d=json.load(sys.stdin); print('Total de alunos:', d['total_alunos']); print('IDs:', [a['aluno_id'] for a in d['alunos']])"
```
**Esperado:** `Total de alunos: 3` e lista `['123', '456', '1']`

```bash
# Consultar detalhes do aluno 456 (que tem falta)
curl -s http://localhost:5003/alunos/456/faltas | python3 -c "import sys,json; d=json.load(sys.stdin); print('Aluno 456 - Registros:'); [print(f'  - {r[\"status\"]} na {r[\"aula_id\"]} (por {r[\"registrado_por\"]})') for r in d['registros']]"
```
**Esperado:** Mostra a falta registrada pelo Professor

**✅ DAE consegue ver dados registrados pelo Professor = SINCRONIZAÇÃO FUNCIONANDO**

---

### ▶️ PASSO 6: Aluno Consulta Seu Histórico

#### 📱 Terminal 3 (Aluno):

```bash
# Consultar próprio histórico (aluno ID "1")
curl -s http://localhost:5002/alunos/1/faltas | python3 -c "import sys,json; d=json.load(sys.stdin); print('Aluno 1 - Registros:', len(d['registros']) if d['registros'] else 0); [print(f'  - {r[\"status\"]} na {r[\"aula_id\"]} (por {r[\"registrado_por\"]})') for r in (d['registros'] or [])]"
```
**Esperado:** Mostra a presença registrada pelo Professor

**✅ Aluno vê dados sincronizados do Professor**

---

### ▶️ PASSO 7: Aluno Tenta Acessar Dados de Outro (Teste de Permissão)

#### 📱 Terminal 3 (Aluno):

```bash
# Tentar consultar aluno 456 (não permitido)
curl http://localhost:5002/alunos/456/faltas
```
**Esperado:** `Você só pode consultar seu próprio histórico`

**✅ Sistema de permissões funcionando corretamente**

---

### ▶️ PASSO 8: DAE Adiciona Justificativa

#### 📱 Terminal 2 (DAE):

```bash
# Adicionar justificativa para a falta do aluno 456
curl -X POST http://localhost:5003/justificativas \
  -H "Content-Type: application/json" \
  -d '{"aluno_id": "456", "aula_id": "AULA-001", "justificativa": "Atestado médico válido"}'
```
**Esperado:** `{"mensagem":"Justificativa adicionada"}`

**✅ Transação pendente adicionada no DAE**

---

### ▶️ PASSO 9: DAE Minera Bloco

#### 📱 Terminal 2 (DAE):

```bash
# Minerar bloco com a justificativa
curl -X POST http://localhost:5003/blocos
```
**Esperado:** Mensagem de sucesso

```bash
# Aguardar propagação
sleep 2
```

**🔄 Blockchain propagada do DAE para Professor e Aluno**

---

### ▶️ PASSO 10: Verificar Sincronização Final (TODOS os nós)

#### 📱 Terminal 1 (Professor):
```bash
curl -s http://localhost:5001/chain | python3 -c "import sys,json; d=json.load(sys.stdin); print('PROFESSOR'); print('  Blocos:', len(d['chain'])); print('  Hash:', d['chain'][-1]['hash'][:16] + '...')"
```
**Esperado:** `Blocos: 3` ✅ **RECEBEU BLOCO DO DAE!**

#### 📱 Terminal 2 (DAE):
```bash
curl -s http://localhost:5003/chain | python3 -c "import sys,json; d=json.load(sys.stdin); print('DAE'); print('  Blocos:', len(d['chain'])); print('  Hash:', d['chain'][-1]['hash'][:16] + '...')"
```
**Esperado:** `Blocos: 3`

#### 📱 Terminal 3 (Aluno):
```bash
curl -s http://localhost:5002/chain | python3 -c "import sys,json; d=json.load(sys.stdin); print('ALUNO'); print('  Blocos visiveis:', len(d['chain']) if d['chain'] else 0)"
```
**Esperado:** `Blocos visiveis: 1`

---

### ▶️ PASSO 11: DAE Consulta Histórico Completo (Evidência Final)

#### 📱 Terminal 2 (DAE):

```bash
# Consultar histórico completo do aluno 456
curl -s http://localhost:5003/alunos/456/faltas | python3 -c "import sys,json; d=json.load(sys.stdin); print('Aluno 456 - Total de registros:', len(d['registros'])); print('\n'.join([f'  {i+1}. {r[\"status\"]} (por {r[\"registrado_por\"]}) - Justificativa: {r.get(\"justificativa\", \"N/A\")}' for i,r in enumerate(d['registros'])]))"
```

**Esperado:**
```
Aluno 456 - Total de registros: 2
  1. ausente (por PROFESSOR-1) - Justificativa: N/A
  2. justificada (por DAE-1) - Justificativa: Atestado médico válido
```

**✅ DAE VÊ TANTO O REGISTRO DO PROFESSOR QUANTO O SEU PRÓPRIO!**
**✅ HISTÓRICO COMPLETO E UNIFICADO!**

---

### ▶️ PASSO 12: Comparar Hashes (Prova de Integridade)

#### 📱 Terminal 1 (Professor):
```bash
curl -s http://localhost:5001/chain | python3 -c "import sys,json; d=json.load(sys.stdin); print('Hash do último bloco (Professor):', d['chain'][-1]['hash'])"
```

#### 📱 Terminal 2 (DAE):
```bash
curl -s http://localhost:5003/chain | python3 -c "import sys,json; d=json.load(sys.stdin); print('Hash do último bloco (DAE):', d['chain'][-1]['hash'])"
```

**Esperado:** Hashes **IDÊNTICOS** entre Professor e DAE

**✅ INTEGRIDADE GARANTIDA - Blockchains sincronizadas perfeitamente!**

---

### ▶️ PASSO 13: Ver Logs de Sincronização

```bash
# Ver logs de propagação
docker-compose logs | grep "Blockchain"
```

**Esperado:** Logs mostrando:
- `📤 Blockchain propagada com sucesso para...`
- `✅ Blockchain atualizada via sync. Novos blocos: X`

---

### 📊 Resumo das Evidências

Ao completar todos os passos, você terá comprovado:

| Funcionalidade | Evidência | Status |
|----------------|-----------|--------|
| **Sincronização Bidirecional** | Professor minera → DAE/Aluno recebem<br>DAE minera → Professor/Aluno recebem | ✅ |
| **Integridade** | Hashes idênticos entre nós<br>Blockchain validada antes de aceitar | ✅ |
| **Permissões** | Aluno só vê seus dados<br>DAE vê tudo<br>Acesso negado ao tentar ver outros | ✅ |
| **Histórico Unificado** | DAE vê registros do Professor + dele mesmo<br>Todos mantêm a mesma blockchain | ✅ |
| **Propagação Automática** | Logs mostram sincronização após mineração<br>Não precisa sincronização manual | ✅ |

---

### 🧹 Limpeza Após os Testes

```bash
# Parar e remover containers
docker-compose down

# (Opcional) Remover imagens
docker-compose down --rmi all

# Reiniciar do zero
docker-compose up -d --build
```

---

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

#### `POST /sync`
Recebe uma blockchain de outro nó e sincroniza (se válida e maior).

**Body:**
```json
{
  "chain": [
    {
      "index": 1,
      "timestamp": 1234567890,
      "transactions": [],
      "prev_hash": "genesis",
      "hash": "..."
    }
  ]
}
```

**Resposta (sucesso):**
```json
{
  "mensagem": "Blockchain atualizada com sucesso",
  "novos_blocos": 3,
  "blocos_locais": 3
}
```

**Resposta (não atualizado):**
```json
{
  "mensagem": "Blockchain não atualizada (local é maior ou igual)",
  "blocos_locais": 2
}
```

**Nota**: Este endpoint é chamado automaticamente após mineração. Não é necessário chamá-lo manualmente em operação normal.

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

# Executar nó do professor (sem peers para teste local)
NODE_ID=PROFESSOR-1 NODE_ROLE=PROFESSOR PORT=8080 go run ./cmd/node

# Executar nó do aluno (em outro terminal)
NODE_ID=ALUNO-1 NODE_ROLE=ALUNO PORT=8081 go run ./cmd/node

# Executar nó do DAE (em outro terminal)
NODE_ID=DAE-1 NODE_ROLE=DAE PORT=8082 go run ./cmd/node

# Para testar sincronização local, adicione PEERS:
NODE_ID=PROFESSOR-1 NODE_ROLE=PROFESSOR PORT=8080 \
  PEERS=http://localhost:8081,http://localhost:8082 \
  go run ./cmd/node
```

**Variáveis de Ambiente:**
- `NODE_ID`: Identificador do nó (ex: PROFESSOR-1)
- `NODE_ROLE`: Papel do nó (PROFESSOR, ALUNO ou DAE)
- `PORT`: Porta do servidor HTTP
- `PEERS`: Lista de URLs dos outros nós separados por vírgula (opcional)

### Compilar

```bash
go build -o node ./cmd/node
```

## 📍 Checkpoint - Estado Atual do Projeto

### Comportamento Atual - Com Sincronização ✅

**✨ NOVA FUNCIONALIDADE**: Os nós agora sincronizam automaticamente suas blockchains! Quando Professor ou DAE mineram um bloco, ele é **propagado para todos os outros nós**.

#### Como Funciona a Sincronização

1. **Professor ou DAE mineram** um novo bloco
2. **Propagação automática**: O bloco é enviado para todos os peers configurados
3. **Validação**: Cada nó recebe e valida a blockchain
4. **Substituição**: Se a blockchain recebida for válida e maior, substitui a local
5. **Consistência**: Todos os nós mantêm a mesma blockchain

**Característica**: Sincronização simples sem consenso - ideal para ambiente sem falhas ou conflitos simultâneos.

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
- ❌ Blockchain **NÃO** é atualizada ainda (transação fica pendente)
- ❌ **Nenhuma** sincronização ocorre (transações pendentes não são propagadas)

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
- 🔄 **Sincronização automática**: Blockchain é propagada para Aluno e DAE
- ✅ **Aluno e DAE recebem** e atualizam suas blockchains

**Estado das Blockchains**:
- 🔵 **Professor**: Novo bloco adicionado
- 🟡 **Aluno**: ✅ **Sincronizado** (recebe o bloco do Professor)
- 🟢 **DAE**: ✅ **Sincronizado** (recebe o bloco do Professor)

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
- ❌ Blockchain **NÃO** é atualizada ainda (transação fica pendente)
- ❌ **Nenhuma** sincronização ocorre (transações pendentes não são propagadas)
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
- 🔄 **Sincronização automática**: Blockchain é propagada para Professor e Aluno
- ✅ **Professor e Aluno recebem** e atualizam suas blockchains

**Estado das Blockchains**:
- 🔵 **Professor**: ✅ **Sincronizado** (recebe o bloco do DAE)
- 🟡 **Aluno**: ✅ **Sincronizado** (recebe o bloco do DAE)
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
- ✅ **Professor/DAE**: Veem todos os blocos e transações  
- ✅ **Aluno**: Vê apenas blocos que contêm transações do seu ID (filtragem de privacidade)
- ✅ **Após sincronização**: Todos os nós têm a **mesma blockchain** internamente

**Estado das Blockchains**: Nenhuma alteração (operação de leitura)

---

#### 6️⃣ Aluno Consulta Suas Faltas

**Endpoint**: `GET /alunos/{id}/faltas` (porta 5002)

**Exemplo**:
```bash
curl http://localhost:5002/alunos/1/faltas
```

**O que acontece**:
- ✅ Busca na blockchain **local do Aluno** (sincronizada) todas as transações do ID especificado
- ✅ Sistema de **permissão**: Aluno só pode consultar seu próprio ID
  - ID do aluno é extraído do `NODE_ID` (ex: `ALUNO-1` → ID = `1`)
  - Se tentar consultar outro ID: retorna **403 Forbidden**
- ✅ **Com sincronização**: O aluno tem acesso a todas as transações mineradas por Professor/DAE

**Nota**: Após sincronização, o aluno pode consultar seu histórico completo mesmo sem poder minerar.

**Estado das Blockchains**: Nenhuma alteração (operação de leitura)

---

#### 7️⃣ DAE Consulta Faltas de Qualquer Aluno

**Endpoint**: `GET /alunos/{id}/faltas` (porta 5003)

**Exemplo**:
```bash
curl http://localhost:5003/alunos/123/faltas
```

**O que acontece**:
- ✅ Busca na blockchain **local do DAE** (sincronizada) todas as transações do aluno especificado
- ✅ **Sem restrição de ID**: DAE pode consultar qualquer aluno
- ✅ Retorna todas as transações (presenças, faltas e justificativas) do aluno
- ✅ **Com sincronização**: DAE tem acesso a registros criados tanto por ele quanto pelo Professor

**Estado das Blockchains**: Nenhuma alteração (operação de leitura)

---

#### 8️⃣ DAE Consulta Todos os Alunos

**Endpoint**: `GET /alunos` (porta 5003)

**Exemplo**:
```bash
curl http://localhost:5003/alunos
```

**O que acontece**:
- ✅ Percorre toda a blockchain **local do DAE** (sincronizada)
- ✅ Agrupa todas as transações por `aluno_id`
- ✅ Retorna um mapa com todos os alunos e seus respectivos registros
- ✅ **Com sincronização**: Inclui registros de todos os nós (Professor, DAE)

**Estado das Blockchains**: Nenhuma alteração (operação de leitura)

---

### Fluxo Completo de Teste com Sincronização

Para entender o comportamento sincronizado das blockchains:

#### Cenário 1: Professor Registra e Minera (com propagação)

```bash
# 1. Professor registra 2 presenças
curl -X POST http://localhost:5001/presencas -H "Content-Type: application/json" \
  -d '{"aluno_id": "123", "aula_id": "AULA-001", "status": "presente"}'

curl -X POST http://localhost:5001/presencas -H "Content-Type: application/json" \
  -d '{"aluno_id": "456", "aula_id": "AULA-001", "status": "ausente"}'

# 2. Professor minera (propaga automaticamente)
curl -X POST http://localhost:5001/blocos

# 3. Verificar blockchains (aguarde 1-2s para propagação)
curl http://localhost:5001/chain  # ✅ Tem 2 blocos
curl http://localhost:5002/chain  # ✅ Tem 2 blocos (sincronizado!)
curl http://localhost:5003/chain  # ✅ Tem 2 blocos (sincronizado!)
```

**Resultado com Sincronização**: 
- 🔵 **Professor**: 2 blocos (minerou)
- 🟡 **Aluno**: 2 blocos ✅ (recebeu via sync)
- 🟢 **DAE**: 2 blocos ✅ (recebeu via sync)
- 🔗 **Todos sincronizados com hash idêntico!**

---

#### Cenário 2: DAE Adiciona Justificativa (com propagação)

```bash
# 1. DAE adiciona justificativa
curl -X POST http://localhost:5003/justificativas -H "Content-Type: application/json" \
  -d '{"aluno_id": "456", "aula_id": "AULA-001", "justificativa": "Atestado médico"}'

# 2. DAE minera (propaga automaticamente)
curl -X POST http://localhost:5003/blocos

# 3. Verificar blockchains
curl http://localhost:5001/chain  # ✅ Agora tem 3 blocos (sincronizado!)
curl http://localhost:5003/chain  # ✅ Tem 3 blocos (minerou)
```

**Resultado com Sincronização**:
- 🔵 **Professor**: 3 blocos ✅ (recebeu bloco do DAE)
- 🟡 **Aluno**: 3 blocos ✅ (recebeu bloco do DAE)
- 🟢 **DAE**: 3 blocos (minerou)
- 🔗 **Blockchains unificadas com histórico completo!**

---

#### Cenário 3: Aluno e DAE Consultam Dados

```bash
# Aluno com NODE_ID=ALUNO-1 tenta consultar
curl http://localhost:5002/alunos/1/faltas     # ✅ Permitido (seu próprio ID)
curl http://localhost:5002/alunos/123/faltas   # ❌ 403 Forbidden (ID diferente)

# DAE consulta aluno 456 (que tem falta + justificativa)
curl http://localhost:5003/alunos/456/faltas
# Resposta mostra:
# - Falta registrada pelo Professor
# - Justificativa registrada pelo DAE
```

**Resultado**:
- ✅ Aluno pode consultar seus dados (se existirem)
- ✅ DAE vê **histórico completo** incluindo ações de ambos os nós
- ✅ Sistema de permissões funcionando corretamente

---

### Limitações Conhecidas

1. **✅ Sincronização Simples Implementada**
   - ✅ Blockchains são sincronizadas automaticamente após mineração
   - ✅ Todos os nós mantêm a mesma blockchain
   - ⚠️ **Sem consenso**: Aceita blockchain maior sem votação
   - ⚠️ **Sem tolerância a falhas**: Assume rede confiável
   - ⚠️ **Sem resolução de conflitos**: Não suporta mineração simultânea

2. **Armazenamento em Memória**
   - Blockchain é perdida ao reiniciar o container
   - Não há persistência em banco de dados
   - Para produção, implemente persistência

3. **Sistema de Permissões na Visualização**
   - Aluno vê apenas blocos com suas transações (filtro de privacidade)
   - Mesmo com blockchain sincronizada, aluno tem visão limitada
   - DAE e Professor veem toda a cadeia

4. **Validações de Negócio Limitadas**
   - DAE pode criar justificativas sem verificar falta prévia
   - Não há verificação de duplicatas de transações
   - Status podem ser inconsistentes (ex: ausente + justificada na mesma aula)

---

### ✅ Funcionalidades Implementadas

- ✅ **Sincronização Automática**: Blockchains propagadas após mineração
- ✅ **Endpoint `/sync`**: Recebe e valida blockchains de outros nós
- ✅ **Configuração de Peers**: Cada nó conhece seus pares via `PEERS`
- ✅ **Validação de Integridade**: Verifica blockchain antes de substituir

### Próximos Passos (Sugestões de Melhorias)

Para evoluir o projeto para produção, considere implementar:

1. **Algoritmo de Consenso**: PBFT, Raft ou Proof of Authority para resolver conflitos
2. **Persistência**: Salvar blockchain em banco de dados (PostgreSQL, MongoDB)
3. **Resolução de Conflitos**: Lidar com mineração simultânea em múltiplos nós
4. **Descoberta de Peers**: Protocolo para adicionar/remover nós dinamicamente
5. **Validações de Negócio**:
   - Verificar se falta existe antes de justificar
   - Prevenir duplicatas de transações
   - Validar sequência de eventos (presença → falta → justificativa)
6. **Tolerância a Falhas**: Retry de propagação, detecção de nós offline
7. **Monitoramento**: Logs estruturados, métricas de sincronização
8. **API de Status**: Endpoint para verificar saúde e sincronização dos nós

---

## Notas Importantes

- ✅ **Sincronização implementada**: Blockchains são automaticamente sincronizadas após mineração
- 📋 **Armazenamento em memória**: Cada nó mantém sua cópia da blockchain em RAM
- 🔄 **Propagação automática**: Professor e DAE propagam blocos para peers após mineração
- ⚠️ **Simplicidade**: Sistema sem consenso complexo, ideal para demonstração e aprendizado
- 🏭 **Produção**: Para ambiente real, adicione persistência, consenso e tolerância a falhas
- 🔒 **Segurança**: Em produção, implemente autenticação entre nós e criptografia de transporte

## Licença

Este projeto é para fins educacionais!

