'use strict';

/**
 * High-performance SQL Value Tuple Tokenizer
 * Parses tuples like (1, 'John Doe', 'john@example.com', NULL, 120.5)
 * Handles escaped quotes ('It\'s', 'It''s'), nested parentheses, and binary data.
 */
function parseSqlValuesTuples(valuesString) {
  const tuples = [];
  let currentTuple = [];
  let currentVal = '';
  let inString = false;
  let escapeNext = false;
  let depth = 0;

  const len = valuesString.length;
  for (let i = 0; i < len; i++) {
    const char = valuesString[i];

    if (escapeNext) {
      currentVal += char;
      escapeNext = false;
      continue;
    }

    if (char === '\\' && inString) {
      currentVal += char;
      escapeNext = true;
      continue;
    }

    if (char === "'") {
      // Check for SQL standard escaped quote ''
      if (inString && i + 1 < len && valuesString[i + 1] === "'") {
        currentVal += "''";
        i++;
        continue;
      }
      inString = !inString;
      currentVal += char;
      continue;
    }

    if (!inString) {
      if (char === '(') {
        depth++;
        if (depth === 1) {
          currentTuple = [];
          currentVal = '';
          continue;
        }
      } else if (char === ')') {
        depth--;
        if (depth === 0) {
          currentTuple.push(currentVal.trim());
          currentVal = '';
          tuples.push(currentTuple);
          continue;
        }
      } else if (char === ',' && depth === 1) {
        currentTuple.push(currentVal.trim());
        currentVal = '';
        continue;
      }
    }

    if (depth > 0) {
      currentVal += char;
    }
  }

  return tuples;
}

function serializeSqlTuples(tuples) {
  return tuples.map(t => `(${t.join(', ')})`).join(', ');
}

module.exports = {
  parseSqlValuesTuples,
  serializeSqlTuples
};
