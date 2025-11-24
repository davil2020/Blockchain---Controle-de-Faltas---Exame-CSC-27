package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"blockchain-faltas/internal/blockchain"
	"blockchain-faltas/internal/node"

	"github.com/gorilla/mux"
)

type Server struct {
	Node       *node.Node
	Blockchain *blockchain.Blockchain
}

func NewServer(n *node.Node, bc *blockchain.Blockchain) *Server {
	return &Server{
		Node:       n,
		Blockchain: bc,
	}
}

func (s *Server) Router() http.Handler {
	r := mux.NewRouter()

	// comum a todos
	r.HandleFunc("/chain", s.getChain).Methods("GET")
	r.HandleFunc("/sync", s.syncChain).Methods("POST")

	// só professor pode registrar presença e minerar
	if s.Node.Role == node.RoleProfessor {
		r.HandleFunc("/presencas", s.registrarPresenca).Methods("POST")
		r.HandleFunc("/blocos", s.minerarBloco).Methods("POST")
	}

	// DAE pode adicionar justificativas e minerar blocos
	if s.Node.Role == node.RoleDAE {
		r.HandleFunc("/justificativas", s.adicionarJustificativa).Methods("POST")
		r.HandleFunc("/blocos", s.minerarBloco).Methods("POST")
	}

	// aluno/DAE podem consultar faltas por aluno
	r.HandleFunc("/alunos/{id}/faltas", s.getFaltasAluno).Methods("GET")

	// DAE pode consultar histórico completo de todos os alunos
	if s.Node.Role == node.RoleDAE {
		r.HandleFunc("/alunos", s.getAllAlunos).Methods("GET")
	}

	return r
}

func (s *Server) getChain(w http.ResponseWriter, r *http.Request) {
	// DAE e Professor podem ver toda a cadeia
	// Aluno só pode ver blocos relacionados a ele
	var chainToShow []blockchain.Block

	if s.Node.Role == node.RoleDAE || s.Node.Role == node.RoleProfessor {
		chainToShow = s.Blockchain.Chain
	} else if s.Node.Role == node.RoleAluno {
		// Aluno só vê blocos com suas transações
		alunoID := extractAlunoIDFromNodeID(s.Node.ID)
		for _, block := range s.Blockchain.Chain {
			var filteredTxs []blockchain.Transaction
			for _, tx := range block.Transactions {
				if tx.AlunoID == alunoID {
					filteredTxs = append(filteredTxs, tx)
				}
			}
			if len(filteredTxs) > 0 {
				filteredBlock := block
				filteredBlock.Transactions = filteredTxs
				chainToShow = append(chainToShow, filteredBlock)
			}
		}
	} else {
		chainToShow = s.Blockchain.Chain
	}

	resp := map[string]interface{}{
		"node_id": s.Node.ID,
		"role":    s.Node.Role,
		"chain":   chainToShow,
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) registrarPresenca(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AlunoID string `json:"aluno_id"`
		AulaID  string `json:"aula_id"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	tx := blockchain.Transaction{
		AlunoID:       body.AlunoID,
		AulaID:        body.AulaID,
		Status:        body.Status,
		RegistradoPor: s.Node.ID,
		Timestamp:     time.Now().Unix(),
	}
	s.Blockchain.AddTransaction(tx)

	json.NewEncoder(w).Encode(map[string]string{
		"mensagem": "Transação adicionada",
	})
}

func (s *Server) minerarBloco(w http.ResponseWriter, r *http.Request) {
	if len(s.Blockchain.PendingTransactions) == 0 {
		http.Error(w, "Não há transações pendentes para minerar", http.StatusBadRequest)
		return
	}

	last := s.Blockchain.LastBlock()
	block := s.Blockchain.NewBlock(last.Hash)

	// Verificar integridade após mineração
	if !s.Blockchain.IsValid() {
		http.Error(w, "Erro: Blockchain inválida após mineração", http.StatusInternalServerError)
		return
	}

	// Propagar blockchain para todos os peers
	go s.propagateChain()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensagem":         "Bloco minerado com sucesso",
		"bloco":            block,
		"total_transacoes": len(block.Transactions),
	})
}

func (s *Server) getFaltasAluno(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alunoID := vars["id"]

	// Aluno só pode ver seu próprio histórico
	if s.Node.Role == node.RoleAluno {
		// Extrair aluno_id do NODE_ID (formato: ALUNO-{id})
		nodeAlunoID := extractAlunoIDFromNodeID(s.Node.ID)
		if nodeAlunoID != alunoID {
			http.Error(w, "Você só pode consultar seu próprio histórico", http.StatusForbidden)
			return
		}
	}

	var faltas []blockchain.Transaction
	for _, b := range s.Blockchain.Chain {
		for _, tx := range b.Transactions {
			if tx.AlunoID == alunoID {
				faltas = append(faltas, tx)
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"aluno_id":  alunoID,
		"registros": faltas,
	})
}

func (s *Server) adicionarJustificativa(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AlunoID       string `json:"aluno_id"`
		AulaID        string `json:"aula_id"`
		Justificativa string `json:"justificativa"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Criar transação de justificativa
	tx := blockchain.Transaction{
		AlunoID:       body.AlunoID,
		AulaID:        body.AulaID,
		Status:        "justificada",
		RegistradoPor: s.Node.ID,
		Timestamp:     time.Now().Unix(),
		Justificativa: body.Justificativa,
	}
	s.Blockchain.AddTransaction(tx)

	json.NewEncoder(w).Encode(map[string]string{
		"mensagem": "Justificativa adicionada",
	})
}

func (s *Server) getAllAlunos(w http.ResponseWriter, r *http.Request) {
	// Mapear todos os alunos e seus registros
	alunosMap := make(map[string][]blockchain.Transaction)

	for _, b := range s.Blockchain.Chain {
		for _, tx := range b.Transactions {
			alunosMap[tx.AlunoID] = append(alunosMap[tx.AlunoID], tx)
		}
	}

	// Converter para formato de resposta
	result := make([]map[string]interface{}, 0, len(alunosMap))
	for alunoID, registros := range alunosMap {
		result = append(result, map[string]interface{}{
			"aluno_id":  alunoID,
			"registros": registros,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_alunos": len(result),
		"alunos":       result,
	})
}

// extractAlunoIDFromNodeID extrai o ID do aluno do NODE_ID
// Exemplo: "ALUNO-123" -> "123"
func extractAlunoIDFromNodeID(nodeID string) string {
	if len(nodeID) > 6 && nodeID[:6] == "ALUNO-" {
		return nodeID[6:]
	}
	return nodeID
}

// syncChain recebe uma blockchain de outro nó e tenta substituir a local
func (s *Server) syncChain(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Chain []blockchain.Block `json:"chain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if s.Blockchain.ReplaceChain(body.Chain) {
		log.Printf("✅ [%s] Blockchain atualizada via sync. Novos blocos: %d", s.Node.ID, len(body.Chain))
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mensagem":      "Blockchain atualizada com sucesso",
			"novos_blocos":  len(body.Chain),
			"blocos_locais": len(s.Blockchain.Chain),
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mensagem":      "Blockchain não atualizada (local é maior ou igual)",
			"blocos_locais": len(s.Blockchain.Chain),
		})
	}
}

// propagateChain envia a blockchain atual para todos os peers
func (s *Server) propagateChain() {
	if len(s.Node.Peers) == 0 {
		return
	}

	payload := map[string]interface{}{
		"chain": s.Blockchain.Chain,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("❌ [%s] Erro ao serializar blockchain: %v", s.Node.ID, err)
		return
	}

	for _, peer := range s.Node.Peers {
		go func(peerURL string) {
			resp, err := http.Post(peerURL+"/sync", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				log.Printf("⚠️ [%s] Erro ao propagar para %s: %v", s.Node.ID, peerURL, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				log.Printf("📤 [%s] Blockchain propagada com sucesso para %s", s.Node.ID, peerURL)
			}
		}(peer)
	}
}
