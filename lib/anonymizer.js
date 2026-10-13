'use strict';

const { Generators } = require('./generators');
const { parseSqlValuesTuples, serializeSqlTuples } = require('./parser');

class StreamAnonymizer {
  constructor(options = {}) {
    this.config = options.config || { tables: {} };
    this.salt = options.salt || 'default-salt';
    this.generators = new Generators(this.salt);

    // Multi-line insert tracking state
    this.currentContext = null;

    // Normalize table names and rules lookup
    this.tableRules = new Map();
    if (this.config.tables) {
      for (const [rawTable, tableDef] of Object.entries(this.config.tables)) {
        const normTable = rawTable.replace(/[`"\[\]]/g, '').toLowerCase();
        this.tableRules.set(normTable, {
          columns: tableDef.columns || [],
          rules: tableDef.rules || {}
        });
      }
    }
  }

  processLine(line) {
    const trimmed = line.trim();

    // Check if we are continuing a multi-line INSERT statement
    if (this.currentContext) {
      return this._processValuesLine(line, this.currentContext);
    }

    // Check if this line starts an INSERT INTO statement
    const insertHeaderRegex = /^\s*INSERT\s+INTO\s+([^\s(]+)(?:\s*\(([^)]+)\))?\s+VALUES(?:\s*(.+))?$/i;
    const match = line.match(insertHeaderRegex);

    if (!match) {
      return { outputLine: line, maskedCount: 0 };
    }

    const rawTableName = match[1];
    const rawColumnList = match[2];
    const rawValuesPart = match[3]; // might be undefined if newline follows VALUES

    const cleanTable = rawTableName
      .replace(/[`"\[\]]/g, '')
      .split('.')
      .pop()
      .toLowerCase();

    const tableConfig = this.tableRules.get(cleanTable);
    if (!tableConfig) {
      return { outputLine: line, maskedCount: 0 };
    }

    let columns = [];
    if (rawColumnList) {
      columns = rawColumnList
        .split(',')
        .map(c => c.trim().replace(/[`"\[\]]/g, '').toLowerCase());
    } else if (tableConfig.columns && tableConfig.columns.length > 0) {
      columns = tableConfig.columns.map(c => c.toLowerCase());
    }

    const context = {
      tableName: rawTableName,
      columns,
      rules: tableConfig.rules
    };

    if (rawValuesPart && rawValuesPart.trim().length > 0) {
      // Single-line or initial values
      const prefix = line.substring(0, line.toUpperCase().indexOf('VALUES') + 6);
      const res = this._maskValuesChunk(rawValuesPart, context);
      if (rawValuesPart.trim().endsWith(';')) {
        this.currentContext = null;
      } else {
        this.currentContext = context;
      }
      return {
        outputLine: `${prefix} ${res.outputChunk}`,
        maskedCount: res.count
      };
    } else {
      // VALUES is on this line, actual values tuples on subsequent lines
      this.currentContext = context;
      return { outputLine: line, maskedCount: 0 };
    }
  }

  _processValuesLine(line, context) {
    const trimmed = line.trim();
    if (!trimmed.startsWith('(')) {
      // Probably comments or empty line, or statement ended
      if (trimmed.endsWith(';')) {
        this.currentContext = null;
      }
      return { outputLine: line, maskedCount: 0 };
    }

    const res = this._maskValuesChunk(trimmed, context);
    if (trimmed.endsWith(';')) {
      this.currentContext = null;
    }

    // Preserve indentation
    const indentMatch = line.match(/^(\s*)/);
    const indent = indentMatch ? indentMatch[1] : '';

    return {
      outputLine: `${indent}${res.outputChunk}`,
      maskedCount: res.count
    };
  }

  _maskValuesChunk(valuesPart, context) {
    let clean = valuesPart.trim();
    let hasSemicolon = clean.endsWith(';');
    let hasComma = !hasSemicolon && clean.endsWith(',');

    if (hasSemicolon) clean = clean.slice(0, -1).trim();
    if (hasComma) clean = clean.slice(0, -1).trim();

    const tuples = parseSqlValuesTuples(clean);
    if (!tuples || tuples.length === 0) {
      return { outputChunk: valuesPart, count: 0 };
    }

    const { columns, rules } = context;
    for (const tuple of tuples) {
      for (let colIdx = 0; colIdx < columns.length; colIdx++) {
        const colName = columns[colIdx];
        const rule = rules[colName];
        if (rule && colIdx < tuple.length) {
          const rawVal = tuple[colIdx];
          tuple[colIdx] = this.generators.maskGeneric(rule, rawVal);
        }
      }
    }

    const serialized = serializeSqlTuples(tuples);
    const suffix = hasSemicolon ? ';' : (hasComma ? ',' : '');

    return {
      outputChunk: `${serialized}${suffix}`,
      count: tuples.length
    };
  }
}

module.exports = { StreamAnonymizer };
