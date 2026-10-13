'use strict';

const assert = require('assert');
const { StreamAnonymizer } = require('../lib/anonymizer');
const { parseSqlValuesTuples, serializeSqlTuples } = require('../lib/parser');
const { Generators } = require('../lib/generators');

console.log('Running db-masker Test Suite...');

// Test 1: Parser unit tests
{
  const raw = "(1, 'Hello, World!', NULL, 45.2, 'It''s fine', 'Escaped \\' quote')";
  const tuples = parseSqlValuesTuples(raw);
  assert.strictEqual(tuples.length, 1);
  assert.strictEqual(tuples[0][0], '1');
  assert.strictEqual(tuples[0][1], "'Hello, World!'");
  assert.strictEqual(tuples[0][2], 'NULL');
  assert.strictEqual(tuples[0][3], '45.2');
  console.log('✔ Test 1: SQL Tuple Tokenizer passed');
}

// Test 2: Deterministic Referencing (Referential Integrity)
{
  const gen1 = new Generators('secret-salt-1');
  const gen2 = new Generators('secret-salt-1');
  const genDifferentSalt = new Generators('secret-salt-2');

  const idA = gen1.maskForeignOrPrimaryKey('42', 'user_id');
  const idB = gen2.maskForeignOrPrimaryKey('42', 'user_id');
  const idDiff = genDifferentSalt.maskForeignOrPrimaryKey('42', 'user_id');

  assert.strictEqual(idA, idB, 'Deterministic hashing must produce identical IDs with same salt');
  assert.notStrictEqual(idA, idDiff, 'Different salt must produce different masked IDs');
  console.log('✔ Test 2: Referential Integrity & Deterministic Hashing passed');
}

// Test 3: StreamAnonymizer Multi-line & Column Mapping
{
  const config = {
    tables: {
      users: {
        rules: {
          id: { type: 'pk_id', domain: 'user_id' },
          email: { type: 'email' },
          name: { type: 'name' }
        }
      }
    }
  };

  const anon = new StreamAnonymizer({ config, salt: 'test-salt' });
  const line1 = "INSERT INTO users (id, name, email) VALUES";
  const line2 = "(1, 'Kevin Naserwan', 'kevin@test.com');";

  const res1 = anon.processLine(line1);
  assert.strictEqual(res1.maskedCount, 0);

  const res2 = anon.processLine(line2);
  assert.strictEqual(res2.maskedCount, 1);
  assert(!res2.outputLine.includes('Kevin Naserwan'), 'PII Name must be masked');
  assert(!res2.outputLine.includes('kevin@test.com'), 'PII Email must be masked');
  assert(res2.outputLine.endsWith(';'), 'Must preserve trailing semicolon');
  console.log('✔ Test 3: StreamAnonymizer Multi-line processing passed');
}

console.log('All tests passed successfully (3/3)!');
