package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

type Transaction struct {
	AlunoID       string `json:"aluno_id"`
	AulaID        string `json:"aula_id"`
	Status        string `json:"status"`         // presente / ausente / justificada
	RegistradoPor string `json:"registrado_por"` // nó que lançou
	Timestamp     int64  `json:"timestamp"`
	Justificativa string `json:"justificativa,omitempty"` // opcional, usado por DAE
}

type Block struct {
	Index        int           `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	PrevHash     string        `json:"prev_hash"`
	Hash         string        `json:"hash"`
}

type Blockchain struct {
	Chain               []Block
	PendingTransactions []Transaction
}

func NewBlockchain() *Blockchain {
	bc := &Blockchain{}
	// bloco gênese
	bc.newBlock("genesis")
	return bc
}

func (bc *Blockchain) newBlock(prevHash string) Block {
	block := Block{
		Index:        len(bc.Chain) + 1,
		Timestamp:    time.Now().Unix(),
		Transactions: bc.PendingTransactions,
		PrevHash:     prevHash,
	}
	block.Hash = calcHash(block)
	bc.PendingTransactions = []Transaction{}
	bc.Chain = append(bc.Chain, block)
	return block
}

func (bc *Blockchain) AddTransaction(tx Transaction) {
	bc.PendingTransactions = append(bc.PendingTransactions, tx)
}

func (bc *Blockchain) LastBlock() Block {
	return bc.Chain[len(bc.Chain)-1]
}

func (bc *Blockchain) NewBlock(prevHash string) Block {
	return bc.newBlock(prevHash)
}

func calcHash(b Block) string {
	blockWithoutHash := struct {
		Index        int
		Timestamp    int64
		Transactions []Transaction
		PrevHash     string
	}{
		Index:        b.Index,
		Timestamp:    b.Timestamp,
		Transactions: b.Transactions,
		PrevHash:     b.PrevHash,
	}
	data, _ := json.Marshal(blockWithoutHash)
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// IsValid verifica a integridade da blockchain
func (bc *Blockchain) IsValid() bool {
	if len(bc.Chain) == 0 {
		return false
	}

	// Verifica o bloco gênese
	if bc.Chain[0].PrevHash != "genesis" {
		return false
	}

	// Verifica cada bloco subsequente
	for i := 1; i < len(bc.Chain); i++ {
		prevBlock := bc.Chain[i-1]
		currentBlock := bc.Chain[i]

		// Verifica se o índice está correto
		if currentBlock.Index != prevBlock.Index+1 {
			return false
		}

		// Verifica se o hash anterior está correto
		if currentBlock.PrevHash != prevBlock.Hash {
			return false
		}

		// Verifica se o hash do bloco atual está correto
		if calcHash(currentBlock) != currentBlock.Hash {
			return false
		}
	}

	return true
}

// GetTransactionsByAluno retorna todas as transações de um aluno
func (bc *Blockchain) GetTransactionsByAluno(alunoID string) []Transaction {
	var transactions []Transaction
	for _, block := range bc.Chain {
		for _, tx := range block.Transactions {
			if tx.AlunoID == alunoID {
				transactions = append(transactions, tx)
			}
		}
	}
	return transactions
}

// ReplaceChain substitui a blockchain atual se a nova for válida e maior
func (bc *Blockchain) ReplaceChain(newChain []Block) bool {
	if len(newChain) <= len(bc.Chain) {
		return false
	}

	newBC := &Blockchain{Chain: newChain}
	if !newBC.IsValid() {
		return false
	}

	bc.Chain = newChain
	bc.PendingTransactions = []Transaction{}
	return true
}
