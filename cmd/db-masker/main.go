package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/KevinNaserwan/db-masker/pkg/anonymizer"
)

func main() {
	configPath := flag.String("config", "", "Path to masking rules JSON configuration (required)")
	configShort := flag.String("c", "", "Path to masking rules JSON (shorthand)")
	salt := flag.String("salt", "db-masker-default-salt", "HMAC secret salt for deterministic hashing")
	saltShort := flag.String("s", "", "HMAC secret salt (shorthand)")
	silent := flag.Bool("silent", false, "Suppress stderr progress output")
	version := flag.Bool("version", false, "Print version")
	flag.BoolVar(version, "v", false, "Print version (shorthand)")

	flag.Parse()

	if *version {
		fmt.Println("db-masker (Go Edition) v1.0.0")
		os.Exit(0)
	}

	cfg := *configPath
	if cfg == "" && *configShort != "" {
		cfg = *configShort
	}

	secretSalt := *salt
	if *saltShort != "" {
		secretSalt = *saltShort
	}

	if cfg == "" {
		fmt.Fprintf(os.Stderr, "Error: Masking rules configuration file is required (-c / --config <rules.json>)\n")
		os.Exit(1)
	}

	anon, err := anonymizer.LoadRulesFromFile(cfg, secretSalt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	// Increase buffer capacity for large SQL lines (up to 16MB per line)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 16*1024*1024)

	lineCount := 0
	maskedRows := 0
	startTime := time.Now()

	for scanner.Scan() {
		lineCount++
		outLine, count := anon.ProcessLine(scanner.Text())
		writer.WriteString(outLine)
		writer.WriteByte('\n')
		maskedRows += count

		if !*silent && lineCount%10000 == 0 {
			elapsed := time.Since(startTime).Seconds()
			fmt.Fprintf(os.Stderr, "[db-masker-go] Processed %d lines (%d rows sanitized) in %.2fs\r", lineCount, maskedRows, elapsed)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	writer.Flush()
	if !*silent {
		elapsed := time.Since(startTime).Seconds()
		fmt.Fprintf(os.Stderr, "\n[db-masker-go] Completed! Processed %d lines (%d rows sanitized) in %.3fs.\n", lineCount, maskedRows, elapsed)
	}
}
