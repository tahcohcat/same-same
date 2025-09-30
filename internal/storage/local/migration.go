package local

import (
	"fmt"

	"github.com/tahcohcat/same-same/internal/storage/memory"

	"github.com/sirupsen/logrus"
)

// MigrationManager handles data migration between storage backends
type MigrationManager struct {
	logger *logrus.Logger
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager() *MigrationManager {
	return &MigrationManager{
		logger: logrus.New(),
	}
}

// MigrateMemoryToLocal migrates data from memory storage to local file storage
func (mm *MigrationManager) MigrateMemoryToLocal(memStorage *memory.Storage, localPath, collectionName string) error {
	mm.logger.Info("starting migration from memory to local storage")

	// Create local storage adapter
	adapter, err := NewVectorStorageAdapter(localPath, collectionName)
	if err != nil {
		return fmt.Errorf("failed to create local storage: %w", err)
	}
	defer adapter.Close()

	// Get all vectors from memory
	vectors, err := memStorage.List()
	if err != nil {
		return fmt.Errorf("failed to list vectors: %w", err)
	}

	mm.logger.WithField("count", len(vectors)).Info("found vectors to migrate")

	// Migrate each vector
	migrated := 0
	failed := 0

	for _, vector := range vectors {
		if err := adapter.Store(vector); err != nil {
			mm.logger.WithFields(logrus.Fields{
				"id":    vector.ID,
				"error": err,
			}).Error("failed to migrate vector")
			failed++
			continue
		}
		migrated++

		if migrated%100 == 0 {
			mm.logger.WithField("progress", migrated).Info("migration progress")
		}
	}

	mm.logger.WithFields(logrus.Fields{
		"migrated": migrated,
		"failed":   failed,
		"total":    len(vectors),
	}).Info("migration completed")

	return nil
}

// MigrateLocalToMemory migrates data from local storage to memory
func (mm *MigrationManager) MigrateLocalToMemory(localPath, collectionName string, memStorage *memory.Storage) error {
	mm.logger.Info("starting migration from local to memory storage")

	// Create local storage adapter
	adapter, err := NewVectorStorageAdapter(localPath, collectionName)
	if err != nil {
		return fmt.Errorf("failed to create local storage: %w", err)
	}
	defer adapter.Close()

	// Get all vectors from local storage
	vectors, err := adapter.List()
	if err != nil {
		return fmt.Errorf("failed to list vectors: %w", err)
	}

	mm.logger.WithField("count", len(vectors)).Info("found vectors to migrate")

	// Migrate each vector
	migrated := 0
	failed := 0

	for _, vector := range vectors {
		if err := memStorage.Store(vector); err != nil {
			mm.logger.WithFields(logrus.Fields{
				"id":    vector.ID,
				"error": err,
			}).Error("failed to migrate vector")
			failed++
			continue
		}
		migrated++

		if migrated%100 == 0 {
			mm.logger.WithField("progress", migrated).Info("migration progress")
		}
	}

	mm.logger.WithFields(logrus.Fields{
		"migrated": migrated,
		"failed":   failed,
		"total":    len(vectors),
	}).Info("migration completed")

	return nil
}

// BackupMemoryStorage creates a backup of memory storage
func (mm *MigrationManager) BackupMemoryStorage(memStorage *memory.Storage, backupPath string) error {
	mm.logger.WithField("path", backupPath).Info("creating backup")

	adapter, err := NewVectorStorageAdapter(backupPath, "backup")
	if err != nil {
		return err
	}
	defer adapter.Close()

	return mm.MigrateMemoryToLocal(memStorage, backupPath, "backup")
}

// RestoreMemoryStorage restores memory storage from backup
func (mm *MigrationManager) RestoreMemoryStorage(backupPath string, memStorage *memory.Storage) error {
	mm.logger.WithField("path", backupPath).Info("restoring from backup")

	return mm.MigrateLocalToMemory(backupPath, "backup", memStorage)
}
