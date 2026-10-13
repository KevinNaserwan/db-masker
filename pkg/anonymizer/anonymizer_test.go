package anonymizer

import (
	"strings"
	"testing"

	"github.com/KevinNaserwan/db-masker/pkg/generators"
	"github.com/KevinNaserwan/db-masker/pkg/parser"
)

func TestSQLTupleTokenizer(t *testing.T) {
	raw := "(1, 'Hello, World!', NULL, 45.20, 'It''s fine')"
	tuples := parser.ParseSQLValuesTuples(raw)
	if len(tuples) != 1 {
		t.Fatalf("expected 1 tuple, got %d", len(tuples))
	}
	if tuples[0][0] != "1" || tuples[0][1] != "'Hello, World!'" || tuples[0][2] != "NULL" {
		t.Errorf("tuple parsed incorrectly: %v", tuples[0])
	}
}

func TestDeterministicHashing(t *testing.T) {
	g1 := generators.NewGenerators("salt-alpha")
	g2 := generators.NewGenerators("salt-alpha")
	gDiff := generators.NewGenerators("salt-beta")

	res1 := g1.MaskKey("42", "user_id")
	res2 := g2.MaskKey("42", "user_id")
	resDiff := gDiff.MaskKey("42", "user_id")

	if res1 != res2 {
		t.Errorf("expected deterministic match, got %s vs %s", res1, res2)
	}
	if res1 == resDiff {
		t.Errorf("expected different salts to produce different IDs, got %s", res1)
	}
}

func TestStreamAnonymizerIntegration(t *testing.T) {
	rules := map[string]TableConfig{
		"users": {
			Rules: map[string]RuleConfig{
				"id":    {Type: "pk_id", Domain: "user_id"},
				"email": {Type: "email"},
				"name":  {Type: "name"},
			},
		},
	}
	anon := &StreamAnonymizer{
		Rules:      rules,
		Generators: generators.NewGenerators("test-salt"),
	}

	line1 := "INSERT INTO users (`id`, `name`, `email`) VALUES"
	line2 := "(1, 'Kevin Naserwan', 'kevin@nasercorp.com');"

	_, count1 := anon.ProcessLine(line1)
	if count1 != 0 {
		t.Errorf("header line should not mask count, got %d", count1)
	}

	out2, count2 := anon.ProcessLine(line2)
	if count2 != 1 {
		t.Errorf("values line should mask 1 row, got %d", count2)
	}

	if strings.Contains(out2, "Kevin Naserwan") || strings.Contains(out2, "kevin@nasercorp.com") {
		t.Errorf("PII leaked in masked output: %s", out2)
	}
}
