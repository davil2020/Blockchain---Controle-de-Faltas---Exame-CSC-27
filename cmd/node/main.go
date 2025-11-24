package main

import (
	"log"
	"net/http"
	"os"

	"blockchain-faltas/internal/api"
	"blockchain-faltas/internal/blockchain"
	"blockchain-faltas/internal/node"
)

func main() {
	nodeID := getEnv("NODE_ID", "NODE-1")
	nodeRole := getEnv("NODE_ROLE", "ALUNO")
	port := getEnv("PORT", "8080")

	n := &node.Node{
		ID:   nodeID,
		Role: nodeRole,
	}

	bc := blockchain.NewBlockchain()

	server := api.NewServer(n, bc)

	log.Printf("🚀 Starting node %s with role %s on port %s", nodeID, nodeRole, port)
	log.Printf("📊 Blockchain initialized with %d blocks", len(bc.Chain))

	if err := http.ListenAndServe(":"+port, server.Router()); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue

}
