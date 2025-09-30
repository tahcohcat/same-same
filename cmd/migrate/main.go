package main

import (
	"flag"
	"log"

	"github/tahcohcat/same-same/internal/storage/local"
	"github/tahcohcat/same-same/internal/storage/memory"
)

func main() {
	var (
		mode       = flag.String("mode", "backup", "Migration mode: backup, restore, export")
		sourcePath = flag.String("source", "./data", "Source path for migration")
		targetPath = flag.String("target", "./backup", "Target path for migration")
		collection = flag.String("collection", "vectors", "Collection name")
	)
	flag.Parse()

	migrator := local.NewMigrationManager()

	switch *mode {
	case "backup":
		log.Println("Creating backup of memory storage...")
		memStorage := memory.NewStorage()

		// TODO: Load data into memory storage from running instance

		if err := migrator.BackupMemoryStorage(memStorage, *targetPath); err != nil {
			log.Fatalf("Backup failed: %v", err)
		}
		log.Println("Backup completed successfully")

	case "restore":
		log.Println("Restoring from backup...")
		memStorage := memory.NewStorage()

		if err := migrator.RestoreMemoryStorage(*sourcePath, memStorage); err != nil {
			log.Fatalf("Restore failed: %v", err)
		}
		log.Println("Restore completed successfully")

	case "export":
		log.Println("Exporting collection...")
		localStorage, err := local.NewLocalStorage(*sourcePath)
		if err != nil {
			log.Fatalf("Failed to open storage: %v", err)
		}

		outputFile := *targetPath + "/" + *collection + ".json"
		if err := localStorage.Export(*collection, outputFile); err != nil {
			log.Fatalf("Export failed: %v", err)
		}
		log.Printf("Collection exported to %s\n", outputFile)

	case "import":
		log.Println("Importing collection...")
		localStorage, err := local.NewLocalStorage(*targetPath)
		if err != nil {
			log.Fatalf("Failed to open storage: %v", err)
		}

		inputFile := *sourcePath + "/" + *collection + ".json"
		if err := localStorage.Import(*collection, inputFile); err != nil {
			log.Fatalf("Import failed: %v", err)
		}
		log.Printf("Collection imported from %s\n", inputFile)

	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}
}
