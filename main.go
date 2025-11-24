package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// AttendanceRecord describes a single class session for one student.
type AttendanceRecord struct {
	StudentID   string    `json:"student_id"`
	StudentName string    `json:"student_name"`
	ClassDate   time.Time `json:"class_date"`
	Status      string    `json:"status"` // present, absent, justified
	Notes       string    `json:"notes"`
}

// Block registers one immutable record in the chain.
type Block struct {
	Index     int              `json:"index"`
	Timestamp time.Time        `json:"timestamp"`
	Record    AttendanceRecord `json:"record"`
	PrevHash  string           `json:"prev_hash"`
	Hash      string           `json:"hash"`
}

// Blockchain wraps a simple slice so validation logic stays centralized.
type Blockchain struct {
	Blocks []Block `json:"blocks"`
}

func newGenesisBlock() Block {
	genesisRecord := AttendanceRecord{
		StudentID:   "GENESIS",
		StudentName: "Genesis Block",
		ClassDate:   time.Now().UTC(),
		Status:      "n/a",
		Notes:       "Genesis block para inicializar a blockchain.",
	}
	return newBlock(genesisRecord, Block{})
}

func computeHash(block Block) string {
	payload := fmt.Sprintf("%d%s%s%s%s%s%s",
		block.Index,
		block.Timestamp.UTC().Format(time.RFC3339Nano),
		block.Record.StudentID,
		block.Record.StudentName,
		block.Record.ClassDate.UTC().Format("2006-01-02"),
		block.Record.Status,
		block.PrevHash,
	)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func newBlock(record AttendanceRecord, prev Block) Block {
	block := Block{
		Index:     prev.Index + 1,
		Timestamp: time.Now().UTC(),
		Record:    record,
		PrevHash:  prev.Hash,
	}
	block.Hash = computeHash(block)
	return block
}

func (bc *Blockchain) addRecord(record AttendanceRecord) (Block, error) {
	if strings.TrimSpace(record.StudentID) == "" {
		return Block{}, errors.New("o campo StudentID é obrigatório")
	}
	if record.ClassDate.IsZero() {
		return Block{}, errors.New("ClassDate inválida")
	}
	if record.ClassDate.After(time.Now().Add(24 * time.Hour)) {
		return Block{}, errors.New("ClassDate não pode estar muito à frente")
	}
	status := strings.ToLower(strings.TrimSpace(record.Status))
	switch status {
	case "presente", "ausente", "justificada":
		record.Status = status
	default:
		return Block{}, errors.New("Status deve ser presente, ausente ou justificada")
	}

	if len(bc.Blocks) == 0 {
		bc.Blocks = append(bc.Blocks, newGenesisBlock())
	}

	prev := bc.Blocks[len(bc.Blocks)-1]
	newBlk := newBlock(record, prev)
	bc.Blocks = append(bc.Blocks, newBlk)
	return newBlk, nil
}

func (bc Blockchain) isValid() error {
	if len(bc.Blocks) == 0 {
		return errors.New("blockchain vazia")
	}
	for i := 1; i < len(bc.Blocks); i++ {
		prev := bc.Blocks[i-1]
		current := bc.Blocks[i]
		if current.Index != prev.Index+1 {
			return fmt.Errorf("índice inválido no bloco %d", i)
		}
		if current.PrevHash != prev.Hash {
			return fmt.Errorf("hash anterior não confere no bloco %d", i)
		}
		if computeHash(current) != current.Hash {
			return fmt.Errorf("hash inválido detectado no bloco %d", i)
		}
	}
	return nil
}

func loadChain(path string) (Blockchain, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return Blockchain{}, nil
	}
	if err != nil {
		return Blockchain{}, err
	}
	defer file.Close()

	var chain Blockchain
	if err := json.NewDecoder(file).Decode(&chain); err != nil {
		return Blockchain{}, err
	}
	return chain, nil
}

func (bc Blockchain) save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(bc)
}

func listBlocks(bc Blockchain) {
	if len(bc.Blocks) <= 1 {
		fmt.Println("Nenhum lançamento de falta registrado ainda.")
		return
	}
	for i := 1; i < len(bc.Blocks); i++ {
		block := bc.Blocks[i]
		record := block.Record
		fmt.Printf("[%d] %s | %s | %s | %s\n", block.Index, record.ClassDate.Format("02/01/2006"), record.StudentID, record.Status, record.Notes)
	}
}

func promptInput(prompt string, reader *bufio.Reader) string {
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func parseDate(input string) (time.Time, error) {
	layouts := []string{"02/01/2006", "2006-01-02"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, input); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("data inválida: use dd/mm/aaaa")
}

func main() {
	const ledgerPath = "ledger.json"

	chain, err := loadChain(ledgerPath)
	if err != nil {
		fmt.Println("Erro ao carregar blockchain:", err)
		return
	}
	if len(chain.Blocks) == 0 {
		chain.Blocks = append(chain.Blocks, newGenesisBlock())
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\nBlockchain de Faltas - Escolha uma opção:")
		fmt.Println("1 - Registrar falta/presença")
		fmt.Println("2 - Listar registros")
		fmt.Println("3 - Validar blockchain")
		fmt.Println("4 - Sair")

		choice := promptInput("> ", reader)
		switch choice {
		case "1":
			record, err := collectRecord(reader)
			if err != nil {
				fmt.Println("Erro:", err)
				continue
			}
			block, err := chain.addRecord(record)
			if err != nil {
				fmt.Println("Erro ao adicionar registro:", err)
				continue
			}
			if err := chain.save(ledgerPath); err != nil {
				fmt.Println("Erro ao salvar ledger:", err)
				continue
			}
			fmt.Printf("Bloco %d adicionado com hash %s\n", block.Index, block.Hash[:10])
		case "2":
			listBlocks(chain)
		case "3":
			if err := chain.isValid(); err != nil {
				fmt.Println("Blockchain inválida:", err)
			} else {
				fmt.Println("Blockchain consistente.")
			}
		case "4", "q", "Q":
			fmt.Println("Até a próxima!")
			return
		default:
			fmt.Println("Opção inválida.")
		}
	}
}

func collectRecord(reader *bufio.Reader) (AttendanceRecord, error) {
	studentID := promptInput("Matrícula do aluno: ", reader)
	studentName := promptInput("Nome do aluno: ", reader)
	dateText := promptInput("Data da aula (dd/mm/aaaa): ", reader)
	classDate, err := parseDate(dateText)
	if err != nil {
		return AttendanceRecord{}, err
	}
	status := promptInput("Status (presente/ausente/justificada): ", reader)
	notes := promptInput("Observações: ", reader)

	return AttendanceRecord{
		StudentID:   studentID,
		StudentName: studentName,
		ClassDate:   classDate,
		Status:      status,
		Notes:       notes,
	}, nil
}

