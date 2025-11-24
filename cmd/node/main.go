package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"blockchain-faltas/internal/api"
	"blockchain-faltas/internal/blockchain"
	"blockchain-faltas/internal/node"
)

func main() {
	nodeID := getEnv("NODE_ID", "NODE-1")
	nodeRole := getEnv("NODE_ROLE", "ALUNO")
	port := getEnv("PORT", "8080")
	peersStr := getEnv("PEERS", "")

	var peers []string
	if peersStr != "" {
		peers = strings.Split(peersStr, ",")
		for i, peer := range peers {
			peers[i] = strings.TrimSpace(peer)
		}
	}

	n := &node.Node{
		ID:    nodeID,
		Role:  nodeRole,
		Peers: peers,
	}

	bc := blockchain.NewBlockchain()

	server := api.NewServer(n, bc)

	log.Printf("🚀 Starting node %s with role %s on port %s", nodeID, nodeRole, port)
	log.Printf("📊 Blockchain initialized with %d blocks", len(bc.Chain))
	if len(peers) > 0 {
		log.Printf("🔗 Connected to %d peer(s): %v", len(peers), peers)
	}

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
