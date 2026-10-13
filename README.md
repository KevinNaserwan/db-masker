# db-masker 🛡️

> **High-throughput SQL stream anonymizer & synthetic data masker CLI with deterministic referential integrity preservation.**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Node.js CI](https://img.shields.io/badge/Node.js-%3E%3D16.0.0-green.svg)](https://nodejs.org/)
[![Status](https://img.shields.io/badge/Release-v1.0.0-blue.svg)](https://github.com/KevinNaserwan/db-masker)

---

## 📌 Background & Motivation
In modern engineering teams, developers and QA engineers frequently require production database dumps to reproduce edge-case bugs and benchmark performance locally or in staging environments. However, strict data privacy regulations (**GDPR, HIPAA, and UU PDP**) strictly prohibit storing unmasked **Personally Identifiable Information (PII)** outside protected production environments.

Traditional tools either:
1. Load entire database dumps into memory (causing out-of-memory errors on multi-gigabyte SQL files).
2. Generate random mock data that destroys relational foreign keys, breaking downstream queries and application logic.

**`db-masker`** solves this with a **zero-memory-footprint streaming architecture** and **deterministic cryptographic HMAC hashing**. It allows you to pipe database dumps on-the-fly directly between database instances or files while maintaining strict 100% referential integrity across relational tables.

---

## ✨ Core Features

* **⚡ Ultra High-Throughput Streaming:** Pipes gigabytes of raw SQL dump (`pg_dump`, `mysqldump`) through a zero-dependency streaming lexer without accumulating memory ($O(1)$ memory consumption).
* **🔗 Referential Integrity Preservation:** Anonymizes Primary Keys (`pk_id`) and Foreign Keys (`fk_id`) deterministically using salt-based HMAC-SHA256, ensuring `orders.user_id` consistently matches sanitized `users.id` across millions of rows.
* **🎭 Realistic Synthetic Masking:**
  * **Emails & Names:** Generates realistic localized personas.
  * **National IDs & Phones:** Masks NIK/KTP and phone numbers with valid formats.
  * **Bcrypt Password Hashes:** Replaces production password hashes with known developer test credentials.
  * **Statistical Numeric Variance:** Retains numeric salary/price data distribution with customizable Gaussian variance (e.g., $\pm 10\%$).
* **🛠️ Cloud & Pipe Native:** Designed as a drop-in UNIX filter (`stdin` $\to$ `stdout`).

---

## 🚀 Installation & Quick Start

### Global Installation (CLI)
```bash
npm install -g db-masker
# Or run directly via npx
npx db-masker --help
```

### Direct Pipe Usage (Production $\to$ Staging)
```bash
# PostgreSQL streaming sanitization:
pg_dump -U postgres prod_db | db-masker -c rules.json -s "my-secret-salt" | psql -U dev dev_db

# MySQL streaming sanitization:
mysqldump -u root -p prod_db | db-masker -c rules.json > sanitized_staging.sql
```

---

## ⚙️ Configuration Schema (`rules.json`)

Define masking rules per table and column:

```json
{
  "tables": {
    "users": {
      "rules": {
        "id": { "type": "pk_id", "domain": "user_id" },
        "full_name": { "type": "name" },
        "email": { "type": "email" },
        "phone_number": { "type": "phone" },
        "nik": { "type": "nik" },
        "home_address": { "type": "address" },
        "password_hash": { "type": "password_hash" },
        "monthly_salary": { "type": "numeric_variance", "variance": 10 }
      }
    },
    "orders": {
      "rules": {
        "id": { "type": "pk_id", "domain": "order_id" },
        "user_id": { "type": "fk_id", "domain": "user_id" },
        "shipping_address": { "type": "address" },
        "total_amount": { "type": "numeric_variance", "variance": 15 },
        "credit_card_masked": { "type": "redact" }
      }
    }
  }
}
```

### Supported Masking Generators

| Generator Type | Description | Example Transformation |
| :--- | :--- | :--- |
| `pk_id` / `fk_id` | Deterministic salted numeric ID | `1` $\to$ `907266` (consistent across tables) |
| `name` | Synthetic realistic persona | `'Kevin Naserwan'` $\to$ `'Chandra Hidayat'` |
| `email` | Salt-hashed synthetic email | `'kevin@example.com'` $\to$ `'user_bc5d0e76@demoapp.dev'` |
| `phone` | Formatted mobile number | `'081298765432'` $\to$ `'081263367222'` |
| `nik` | 16-digit national identifier | `'3271012345678901'` $\to$ `'3271001155600587'` |
| `address` | Realistic street & city address | `'Jl. Sudirman 45'` $\to$ `'Jl. Thamrin No. 54, Medan'` |
| `password_hash` | Standard dev bcrypt hash | Generates valid hash for `Secret123!` |
| `numeric_variance`| Preserves distribution within $\pm X\%$ | `25,000,000.00` $\to$ `24,250,000.00` |
| `redact` | Constant redaction | `'4111-XXXX-1111'` $\to$ `'[REDACTED]'` |
| `null` | Replaces with database NULL | `'sensitive'` $\to$ `NULL` |

---

## 🧪 Testing & Verification

Run the automated test suite:
```bash
npm test
```

Run interactive demo on sample SQL dump:
```bash
npm run demo
```

---

## 📄 License
This project is licensed under the [MIT License](LICENSE).

Authored with ❤️ by [Kevin Naserwan](https://github.com/KevinNaserwan).
