package anonymizer

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"github.com/KevinNaserwan/db-masker/pkg/generators"
	"github.com/KevinNaserwan/db-masker/pkg/parser"
)

type RuleConfig struct {
	Type     string  `json:"type"`
	Domain   string  `json:"domain,omitempty"`
	Variance float64 `json:"variance,omitempty"`
}

type TableConfig struct {
	Columns []string              `json:"columns,omitempty"`
	Rules   map[string]RuleConfig `json:"rules"`
}

type MaskingRules struct {
	Tables map[string]TableConfig `json:"tables"`
}

type AnonymizerContext struct {
	TableName string
	Columns   []string
	Rules     map[string]RuleConfig
}

type StreamAnonymizer struct {
	Rules          map[string]TableConfig
	Generators     *generators.Generators
	CurrentContext *AnonymizerContext
}

func LoadRulesFromFile(filePath string, salt string) (*StreamAnonymizer, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var config MaskingRules
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	normalized := make(map[string]TableConfig)
	for k, v := range config.Tables {
		clean := cleanIdentifier(k)
		normRules := make(map[string]RuleConfig)
		for rk, rv := range v.Rules {
			normRules[strings.ToLower(cleanIdentifier(rk))] = rv
		}
		v.Rules = normRules
		normalized[clean] = v
	}

	return &StreamAnonymizer{
		Rules:      normalized,
		Generators: generators.NewGenerators(salt),
	}, nil
}

func cleanIdentifier(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "\"", "")
	s = strings.ReplaceAll(s, "[", "")
	s = strings.ReplaceAll(s, "]", "")
	s = strings.TrimSpace(s)
	parts := strings.Split(s, ".")
	return strings.ToLower(parts[len(parts)-1])
}

var insertHeaderRegex = regexp.MustCompile(`(?i)^\s*INSERT\s+INTO\s+([^\s(]+)(?:\s*\(([^)]+)\))?\s+VALUES(?:\s*(.+))?$`)

func (a *StreamAnonymizer) ProcessLine(line string) (string, int) {
	if a.CurrentContext != nil {
		return a.processValuesLine(line, a.CurrentContext)
	}

	match := insertHeaderRegex.FindStringSubmatch(line)
	if match == nil {
		return line, 0
	}

	rawTableName := match[1]
	rawColumnList := match[2]
	rawValuesPart := match[3]

	cleanTable := cleanIdentifier(rawTableName)
	tableConfig, exists := a.Rules[cleanTable]
	if !exists {
		return line, 0
	}

	var columns []string
	if rawColumnList != "" {
		for _, col := range strings.Split(rawColumnList, ",") {
			columns = append(columns, strings.ToLower(cleanIdentifier(col)))
		}
	} else if len(tableConfig.Columns) > 0 {
		for _, col := range tableConfig.Columns {
			columns = append(columns, strings.ToLower(cleanIdentifier(col)))
		}
	}

	ctx := &AnonymizerContext{
		TableName: rawTableName,
		Columns:   columns,
		Rules:     tableConfig.Rules,
	}

	if strings.TrimSpace(rawValuesPart) != "" {
		prefixIdx := strings.Index(strings.ToUpper(line), "VALUES")
		prefix := line[:prefixIdx+6]
		outChunk, count := a.maskValuesChunk(rawValuesPart, ctx)
		if strings.HasSuffix(strings.TrimSpace(rawValuesPart), ";") {
			a.CurrentContext = nil
		} else {
			a.CurrentContext = ctx
		}
		return prefix + " " + outChunk, count
	}

	a.CurrentContext = ctx
	return line, 0
}

func (a *StreamAnonymizer) processValuesLine(line string, ctx *AnonymizerContext) (string, int) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "(") {
		if strings.HasSuffix(trimmed, ";") {
			a.CurrentContext = nil
		}
		return line, 0
	}

	outChunk, count := a.maskValuesChunk(trimmed, ctx)
	if strings.HasSuffix(trimmed, ";") {
		a.CurrentContext = nil
	}

	indent := ""
	for _, ch := range line {
		if ch == ' ' || ch == '\t' {
			indent += string(ch)
		} else {
			break
		}
	}

	return indent + outChunk, count
}

func (a *StreamAnonymizer) maskValuesChunk(chunk string, ctx *AnonymizerContext) (string, int) {
	clean := strings.TrimSpace(chunk)
	hasSemicolon := strings.HasSuffix(clean, ";")
	hasComma := !hasSemicolon && strings.HasSuffix(clean, ",")

	if hasSemicolon {
		clean = strings.TrimSpace(strings.TrimSuffix(clean, ";"))
	}
	if hasComma {
		clean = strings.TrimSpace(strings.TrimSuffix(clean, ","))
	}

	tuples := parser.ParseSQLValuesTuples(clean)
	if len(tuples) == 0 {
		return chunk, 0
	}

	for _, tuple := range tuples {
		for colIdx, colName := range ctx.Columns {
			rule, hasRule := ctx.Rules[colName]
			if hasRule && colIdx < len(tuple) {
				tuple[colIdx] = a.Generators.MaskGeneric(rule.Type, rule.Domain, rule.Variance, tuple[colIdx])
			}
		}
	}

	serialized := parser.SerializeSQLTuples(tuples)
	suffix := ""
	if hasSemicolon {
		suffix = ";"
	} else if hasComma {
		suffix = ","
	}

	return serialized + suffix, len(tuples)
}
