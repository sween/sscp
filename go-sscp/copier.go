package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/caretdev/go-irisnative/src/connection"
	"github.com/caretdev/go-irisnative/src/iris"
)

type Copier struct {
	cfg *Config
}

func NewCopier(cfg *Config) *Copier {
	return &Copier{cfg: cfg}
}

func (c *Copier) ExecuteCopy() error {
	startTime := time.Now()
	log.Printf("[SSCP] Starting copy operation (Mode: %s)...", c.cfg.Mode)
	log.Printf("[SSCP] Source: %s (%s)", c.cfg.Source.ToSSCPString(), c.cfg.Source.Addr())
	log.Printf("[SSCP] Target: %s (%s)", c.cfg.Target.ToSSCPString(), c.cfg.Target.Addr())

	switch c.cfg.Mode {
	case "direct", "out-of-band", "oob":
		err := c.copyOutOfBand()
		if err != nil {
			return fmt.Errorf("out-of-band copy failed: %w", err)
		}
	case "initiator", "iris":
		err := c.copyInitiator()
		if err != nil {
			return fmt.Errorf("iris initiator copy failed: %w", err)
		}
	default:
		// Default to out-of-band streaming if unspecified
		err := c.copyOutOfBand()
		if err != nil {
			return fmt.Errorf("out-of-band copy failed: %w", err)
		}
	}

	log.Printf("[SSCP] Copy completed successfully in %v", time.Since(startTime))
	return nil
}

// copyOutOfBand streams files using ONLY stock built-in IRIS system classes (%File & %Stream.FileBinary).
// No custom ObjectScript classes (like ZMSP.SuperServer) are required on either IRIS instance.
func (c *Copier) copyOutOfBand() error {
	targetFile := c.cfg.Target.File
	if targetFile == "" {
		return fmt.Errorf("SSCP_TARGET_FILE is required")
	}

	log.Printf("[SSCP Out-Of-Band] Connecting to Target IRIS SuperServer at %s...", c.cfg.Target.Addr())
	targetConn, err := connection.Connect(
		c.cfg.Target.Addr(),
		c.cfg.Target.Namespace,
		c.cfg.Target.User,
		c.cfg.Target.Pass,
	)
	if err != nil {
		return fmt.Errorf("failed to connect to target IRIS SuperServer at %s: %w", c.cfg.Target.Addr(), err)
	}
	defer targetConn.Disconnect()

	// 1. Delete target file if exists using built-in %File class
	var deleteStatus interface{}
	_ = targetConn.ClassMethod("%File", "Delete", &deleteStatus, targetFile)

	// 2. Instantiate stock %Stream.FileBinary stream object on Target IRIS over SuperServer
	var targetStream iris.Oref
	err = targetConn.ClassMethod("%Stream.FileBinary", "%New", &targetStream)
	if err != nil {
		return fmt.Errorf("failed to instantiate stock %%Stream.FileBinary on Target IRIS: %w", err)
	}

	err = targetConn.MethodVoid(string(targetStream), "LinkToFile", targetFile)
	if err != nil {
		return fmt.Errorf("failed to link stream to file %s: %w", targetFile, err)
	}

	// 3. Read source chunks (either from local container file or remote Source IRIS via %Stream.FileBinary)
	sourceFile := c.cfg.Source.File
	if sourceFile == "" {
		return fmt.Errorf("SSCP_SOURCE_FILE is required")
	}

	var totalTransferred int64

	// Check if local source file exists on container runner
	if file, oerr := os.Open(sourceFile); oerr == nil {
		defer file.Close()
		fileInfo, _ := file.Stat()
		log.Printf("[SSCP Out-Of-Band] Streaming local container file %s (%d bytes) to Target IRIS...", sourceFile, fileInfo.Size())

		buffer := make([]byte, 32768)
		for {
			n, rerr := file.Read(buffer)
			if n > 0 {
				chunkStr := string(buffer[:n])
				err := targetConn.MethodVoid(string(targetStream), "Write", chunkStr)
				if err != nil {
					return fmt.Errorf("failed to write chunk to Target IRIS stream: %w", err)
				}
				totalTransferred += int64(n)
			}
			if rerr == io.EOF {
				break
			}
			if rerr != nil {
				return fmt.Errorf("error reading source file: %w", rerr)
			}
		}
	} else {
		// Remote Source IRIS reading over SuperServer using stock %Stream.FileBinary
		log.Printf("[SSCP Out-Of-Band] Connecting to Source IRIS SuperServer at %s...", c.cfg.Source.Addr())
		sourceConn, err := connection.Connect(
			c.cfg.Source.Addr(),
			c.cfg.Source.Namespace,
			c.cfg.Source.User,
			c.cfg.Source.Pass,
		)
		if err != nil {
			return fmt.Errorf("failed to connect to source IRIS SuperServer at %s: %w", c.cfg.Source.Addr(), err)
		}
		defer sourceConn.Disconnect()

		var sourceStream iris.Oref
		err = sourceConn.ClassMethod("%Stream.FileBinary", "%New", &sourceStream)
		if err != nil {
			return fmt.Errorf("failed to instantiate stock %%Stream.FileBinary on Source IRIS: %w", err)
		}

		err = sourceConn.MethodVoid(string(sourceStream), "LinkToFile", sourceFile)
		if err != nil {
			return fmt.Errorf("failed to link source stream to file %s: %w", sourceFile, err)
		}

		for {
			var atEnd int
			_ = sourceConn.PropertyGet(sourceStream, "AtEnd", &atEnd)
			if atEnd == 1 {
				break
			}

			var chunkStr string
			// Read 32KB chunk from Source IRIS
			err := sourceConn.Method(sourceStream, "Read", &chunkStr, 32768)
			if err != nil {
				return fmt.Errorf("failed to read stream chunk from Source IRIS: %w", err)
			}
			if len(chunkStr) == 0 {
				break
			}

			err = targetConn.MethodVoid(string(targetStream), "Write", chunkStr)
			if err != nil {
				return fmt.Errorf("failed to write stream chunk to Target IRIS: %w", err)
			}
			totalTransferred += int64(len(chunkStr))
		}
	}

	// 4. Save stream on Target IRIS
	var saveStatus interface{}
	err = targetConn.Method(targetStream, "%Save", &saveStatus)
	if err != nil {
		return fmt.Errorf("failed to save stream on Target IRIS: %w", err)
	}

	log.Printf("[SSCP Out-Of-Band] Successfully streamed %d bytes to %s on Target IRIS (Zero custom IRIS classes required!)", totalTransferred, targetFile)
	return nil
}

// copyInitiator invokes legacy ZMSP.SuperServer:Copy if installed.
func (c *Copier) copyInitiator() error {
	log.Printf("[SSCP] Connecting to Source IRIS SuperServer at %s...", c.cfg.Source.Addr())
	conn, err := connection.Connect(
		c.cfg.Source.Addr(),
		c.cfg.Source.Namespace,
		c.cfg.Source.User,
		c.cfg.Source.Pass,
	)
	if err != nil {
		return fmt.Errorf("failed to connect to source IRIS SuperServer at %s: %w", c.cfg.Source.Addr(), err)
	}
	defer conn.Disconnect()

	sourceStr := c.cfg.Source.ToSSCPString()
	targetStr := c.cfg.Target.ToSSCPString()

	log.Printf("[SSCP] Invoking ZMSP.SuperServer:Copy via SuperServer RPC...")
	var status string
	err = conn.ClassMethod("ZMSP.SuperServer", "Copy", &status, sourceStr, targetStr)
	if err != nil {
		return fmt.Errorf("ClassMethod ZMSP.SuperServer:Copy execution error: %w", err)
	}

	log.Printf("[SSCP] ZMSP.SuperServer:Copy RPC returned status: %v", status)
	return nil
}
