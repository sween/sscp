package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	onceFlag := flag.Bool("once", false, "Run file copy once and exit")
	flag.Parse()

	log.Println("==================================================")
	log.Println("  InterSystems IRIS Go SuperServer File Copier")
	log.Println("  Powered by caretdev/go-irisnative")
	log.Println("==================================================")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("[SSCP] Configuration error: %v", err)
	}

	if *onceFlag {
		cfg.Once = true
	}

	copier := NewCopier(cfg)

	// One-shot execution
	if cfg.Once {
		log.Println("[SSCP] Running in single-execution (one-shot) mode")
		if err := copier.ExecuteCopy(); err != nil {
			log.Fatalf("[SSCP] Copy failed: %v", err)
		}
		log.Println("[SSCP] Done.")
		return
	}

	// Continuous loop mode
	log.Printf("[SSCP] Running in continuous sync mode (Interval: %v)", cfg.Interval)
	log.Println("[SSCP] Press Ctrl+C or send SIGTERM to terminate")

	// Perform initial sync immediately
	if err := copier.ExecuteCopy(); err != nil {
		log.Printf("[SSCP ERROR] Copy failed: %v (will retry on next tick)", err)
	}

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			if err := copier.ExecuteCopy(); err != nil {
				log.Printf("[SSCP ERROR] Copy failed: %v (will retry on next tick)", err)
			}
		case sig := <-sigChan:
			log.Printf("[SSCP] Received signal %v. Shutting down gracefully...", sig)
			log.Println("[SSCP] Bye!")
			return
		}
	}
}
