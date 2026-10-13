package generators

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Generators struct {
	salt       string
	firstNames []string
	lastNames  []string
	domains    []string
	streets    []string
	cities     []string
}

func NewGenerators(salt string) *Generators {
	if salt == "" {
		salt = "default-salt"
	}
	return &Generators{
		salt: salt,
		firstNames: []string{
			"Ahmad", "Budi", "Chandra", "Dewi", "Eko", "Fitri", "Gita", "Hendra",
			"Indra", "Joko", "Kartika", "Lestari", "Mahendra", "Nurul", "Putri",
			"Rian", "Siti", "Taufik", "Utami", "Wahyu",
		},
		lastNames: []string{
			"Pratama", "Santoso", "Wijaya", "Kusuma", "Saputra", "Siregar",
			"Hidayat", "Wibowo", "Nugroho", "Setiawan", "Permana", "Lestari",
			"Purnomo", "Syahputra",
		},
		domains: []string{"example.com", "testcorp.io", "stagingmail.org", "demoapp.dev", "sandbox.net"},
		streets: []string{"Jl. Merdeka", "Jl. Sudirman", "Jl. Thamrin", "Jl. Gatot Subroto", "Jl. Diponegoro", "Jl. Ahmad Yani", "Jl. Pahlawan", "Jl. Gajah Mada"},
		cities:  []string{"Jakarta Selatan", "Bandung", "Surabaya", "Semarang", "Medan", "Palembang", "Yogyakarta", "Denpasar", "Balikpapan", "Makassar"},
	}
}

func (g *Generators) Hash(val string) string {
	mac := hmac.New(sha256.New, []byte(g.salt))
	mac.Write([]byte(val))
	return hex.EncodeToString(mac.Sum(nil))
}

func (g *Generators) HashInt(val string, max int64) int64 {
	h := g.Hash(val)
	if len(h) > 8 {
		h = h[:8]
	}
	parsed, err := strconv.ParseInt(h, 16, 64)
	if err != nil {
		return 0
	}
	res := parsed % max
	if res < 0 {
		res = -res
	}
	return res
}

func (g *Generators) MaskEmail(val string) string {
	if val == "" || val == "NULL" {
		return val
	}
	clean := strings.Trim(val, "'")
	h := g.Hash(clean)
	if len(h) > 10 {
		h = h[:10]
	}
	domIdx := g.HashInt(clean, int64(len(g.domains)))
	return fmt.Sprintf("'user_%s@%s'", h, g.domains[domIdx])
}

func (g *Generators) MaskName(val string) string {
	if val == "" || val == "NULL" {
		return val
	}
	clean := strings.Trim(val, "'")
	fnIdx := g.HashInt(clean+"_fn", int64(len(g.firstNames)))
	lnIdx := g.HashInt(clean+"_ln", int64(len(g.lastNames)))
	return fmt.Sprintf("'%s %s'", g.firstNames[fnIdx], g.lastNames[lnIdx])
}

func (g *Generators) MaskPhone(val string) string {
	if val == "" || val == "NULL" {
		return val
	}
	clean := strings.Trim(val, "'")
	num := g.HashInt(clean, 100000000)
	return fmt.Sprintf("'0812%08d'", num)
}

func (g *Generators) MaskNik(val string) string {
	if val == "" || val == "NULL" {
		return val
	}
	clean := strings.Trim(val, "'")
	num := g.HashInt(clean, 1000000000000)
	return fmt.Sprintf("'3271%012d'", num)
}

func (g *Generators) MaskAddress(val string) string {
	if val == "" || val == "NULL" {
		return val
	}
	clean := strings.Trim(val, "'")
	stIdx := g.HashInt(clean+"_st", int64(len(g.streets)))
	ctIdx := g.HashInt(clean+"_ct", int64(len(g.cities)))
	no := g.HashInt(clean+"_no", 150) + 1
	return fmt.Sprintf("'%s No. %d, %s'", g.streets[stIdx], no, g.cities[ctIdx])
}

func (g *Generators) MaskNumericVariance(val string, variancePercent float64) string {
	if val == "" || val == "NULL" {
		return val
	}
	num, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return val
	}
	h := float64(g.HashInt(val, 200) - 100)
	factor := 1.0 + ((h / 100.0) * (variancePercent / 100.0))
	res := math.Round(num*factor*100) / 100
	return fmt.Sprintf("%.2f", res)
}

func (g *Generators) MaskKey(val string, domainScope string) string {
	if val == "" || val == "NULL" {
		return val
	}
	clean := strings.Trim(val, "'")
	isQuoted := strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")
	mapped := g.HashInt(domainScope+"_"+clean, 899999) + 100000
	if isQuoted {
		return fmt.Sprintf("'%d'", mapped)
	}
	return strconv.FormatInt(mapped, 10)
}

func (g *Generators) MaskGeneric(ruleType string, domain string, variance float64, rawVal string) string {
	switch strings.ToLower(ruleType) {
	case "email":
		return g.MaskEmail(rawVal)
	case "name":
		return g.MaskName(rawVal)
	case "phone":
		return g.MaskPhone(rawVal)
	case "nik":
		return g.MaskNik(rawVal)
	case "address":
		return g.MaskAddress(rawVal)
	case "password_hash":
		return "'$2a$10$e8wY8bL3yAevjU3WdFz3..m0N4nQz8F7A1hC3LgJz4Yf/8g9aH6a2'"
	case "numeric_variance":
		if variance == 0 {
			variance = 15.0
		}
		return g.MaskNumericVariance(rawVal, variance)
	case "pk_id", "fk_id":
		if domain == "" {
			domain = "global_id"
		}
		return g.MaskKey(rawVal, domain)
	case "redact":
		return "'[REDACTED]'"
	case "null":
		return "NULL"
	default:
		return rawVal
	}
}
