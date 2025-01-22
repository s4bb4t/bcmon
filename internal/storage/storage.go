package storage

import (
	"fmt"
	"log/slog"
	"os"
)

type storage struct {
	log *slog.Logger
}

func NewStorage(log *slog.Logger) *storage {
	return &storage{log: log}
}

func (s *storage) SaveContract(address string) error {
	file, err := os.OpenFile("nft_contracts.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(address + "\n"); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}
