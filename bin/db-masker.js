#!/usr/bin/env node
'use strict';

const fs = require('fs');
const readline = require('readline');
const { StreamAnonymizer } = require('../lib/anonymizer');

function printUsage() {
  console.error(`
db-masker v1.0.0
High-throughput SQL stream anonymizer & synthetic data masker CLI

Usage:
  db-masker --config <path-to-rules.json> [options] < dump.sql > masked.sql
  cat dump.sql | db-masker --config <path-to-rules.json> > masked.sql

Options:
  -c, --config <file>     Path to masking rules JSON configuration (required)
  -s, --salt <string>     Secret salt for deterministic HMAC hashing (default: db-masker-default-salt)
  -b, --batch <size>      Batch lines per processing flush (default: 500)
  --silent                Suppress progress stderr logs
  -v, --version           Print version information
  -h, --help              Print this help screen

Examples:
  # Anonymize PostgreSQL or MySQL dump stream on-the-fly:
  pg_dump -U postgres prod_db | db-masker -c rules.json | psql -U dev dev_db
  db-masker -c examples/rules.json < examples/sample-dump.sql > masked.sql
`);
}

async function main() {
  const args = process.argv.slice(2);
  let configPath = null;
  let salt = 'db-masker-default-salt';
  let silent = false;

  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    if (arg === '-h' || arg === '--help') {
      printUsage();
      process.exit(0);
    }
    if (arg === '-v' || arg === '--version') {
      console.log('db-masker v1.0.0');
      process.exit(0);
    }
    if ((arg === '-c' || arg === '--config') && args[i + 1]) {
      configPath = args[++i];
    }
    if ((arg === '-s' || arg === '--salt') && args[i + 1]) {
      salt = args[++i];
    }
    if (arg === '--silent') {
      silent = true;
    }
  }

  if (!configPath) {
    console.error('Error: Masking rules configuration file is required (--config <rules.json>).');
    printUsage();
    process.exit(1);
  }

  if (!fs.existsSync(configPath)) {
    console.error(`Error: Configuration file not found at: ${configPath}`);
    process.exit(1);
  }

  let rulesConfig;
  try {
    const raw = fs.readFileSync(configPath, 'utf8');
    rulesConfig = JSON.parse(raw);
  } catch (err) {
    console.error(`Error parsing JSON configuration file: ${err.message}`);
    process.exit(1);
  }

  const anonymizer = new StreamAnonymizer({
    config: rulesConfig,
    salt,
    silent
  });

  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
    terminal: false
  });

  let linesCount = 0;
  let maskedRows = 0;
  const startTime = Date.now();

  rl.on('line', (line) => {
    linesCount++;
    const result = anonymizer.processLine(line);
    process.stdout.write(result.outputLine + '\n');
    if (result.maskedCount > 0) {
      maskedRows += result.maskedCount;
    }

    if (!silent && linesCount % 5000 === 0) {
      const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);
      process.stderr.write(`[db-masker] Processed ${linesCount} lines (${maskedRows} rows masked) in ${elapsed}s\r`);
    }
  });

  rl.on('close', () => {
    if (!silent) {
      const elapsed = ((Date.now() - startTime) / 1000).toFixed(2);
      process.stderr.write(`\n[db-masker] Complete! Processed ${linesCount} lines (${maskedRows} rows sanitized) in ${elapsed}s.\n`);
    }
  });
}

main().catch((err) => {
  console.error(`[db-masker Fatal Error]: ${err.stack || err.message}`);
  process.exit(1);
});
