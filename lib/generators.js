'use strict';

const crypto = require('crypto');

class Generators {
  constructor(salt) {
    this.salt = salt || 'default-salt';
    this.firstNames = ['Ahmad', 'Budi', 'Chandra', 'Dewi', 'Eko', 'Fitri', 'Gita', 'Hendra', 'Indra', 'Joko', 'Kartika', 'Lestari', 'Mahendra', 'Nurul', 'Putri', 'Rian', 'Siti', 'Taufik', 'Utami', 'Wahyu'];
    this.lastNames = ['Pratama', 'Santoso', 'Wijaya', 'Kusuma', 'Saputra', 'Siregar', 'Hidayat', 'Wibowo', 'Nugroho', 'Setiawan', 'Permana', 'Lestari', 'Purnomo', 'Syahputra'];
    this.domains = ['example.com', 'testcorp.io', 'stagingmail.org', 'demoapp.dev', 'sandbox.net'];
    this.streets = ['Jl. Merdeka', 'Jl. Sudirman', 'Jl. Thamrin', 'Jl. Gatot Subroto', 'Jl. Diponegoro', 'Jl. Ahmad Yani', 'Jl. Pahlawan', 'Jl. Gajah Mada'];
    this.cities = ['Jakarta Selatan', 'Bandung', 'Surabaya', 'Semarang', 'Medan', 'Palembang', 'Yogyakarta', 'Denpasar', 'Balikpapan', 'Makassar'];
  }

  hash(val) {
    return crypto.createHmac('sha256', this.salt).update(String(val)).digest('hex');
  }

  hashInt(val, max = 1000000000) {
    const h = this.hash(val).substring(0, 8);
    return Math.abs(parseInt(h, 16) % max);
  }

  maskEmail(val) {
    if (!val || val === 'NULL') return val;
    const clean = val.replace(/^'|'$/g, '');
    const h = this.hash(clean).substring(0, 10);
    const domainIdx = this.hashInt(clean, this.domains.length);
    return `'user_${h}@${this.domains[domainIdx]}'`;
  }

  maskName(val) {
    if (!val || val === 'NULL') return val;
    const clean = val.replace(/^'|'$/g, '');
    const fnIdx = this.hashInt(clean + '_fn', this.firstNames.length);
    const lnIdx = this.hashInt(clean + '_ln', this.lastNames.length);
    return `'${this.firstNames[fnIdx]} ${this.lastNames[lnIdx]}'`;
  }

  maskPhone(val) {
    if (!val || val === 'NULL') return val;
    const clean = val.replace(/^'|'$/g, '');
    const numPart = String(this.hashInt(clean, 100000000)).padStart(8, '0');
    return `'0812${numPart}'`;
  }

  maskNik(val) {
    if (!val || val === 'NULL') return val;
    const clean = val.replace(/^'|'$/g, '');
    const numPart = String(this.hashInt(clean, 1000000000000)).padStart(12, '0');
    return `'3271${numPart}'`;
  }

  maskAddress(val) {
    if (!val || val === 'NULL') return val;
    const clean = val.replace(/^'|'$/g, '');
    const streetIdx = this.hashInt(clean + '_st', this.streets.length);
    const cityIdx = this.hashInt(clean + '_ct', this.cities.length);
    const no = (this.hashInt(clean + '_no', 150) + 1);
    return `'${this.streets[streetIdx]} No. ${no}, ${this.cities[cityIdx]}'`;
  }

  maskPasswordHash() {
    // Standard fixed bcrypt hash for 'Secret123!'
    return `'$2a$10$e8wY8bL3yAevjU3WdFz3..m0N4nQz8F7A1hC3LgJz4Yf/8g9aH6a2'`;
  }

  maskNumericVariance(val, variancePercent = 15) {
    if (!val || val === 'NULL' || isNaN(Number(val))) return val;
    const num = parseFloat(val);
    const h = this.hashInt(val, 200) - 100; // -100 to +100
    const factor = 1 + ((h / 100) * (variancePercent / 100));
    const result = Math.round(num * factor * 100) / 100;
    return String(result);
  }

  maskForeignOrPrimaryKey(val, domainScope) {
    if (!val || val === 'NULL') return val;
    const clean = val.replace(/^'|'$/g, '');
    const isQuoted = val.startsWith("'") && val.endsWith("'");
    const mapped = (this.hashInt(`${domainScope}_${clean}`, 899999) + 100000);
    return isQuoted ? `'${mapped}'` : String(mapped);
  }

  maskUuid(val, domainScope) {
    if (!val || val === 'NULL') return val;
    const clean = val.replace(/^'|'$/g, '');
    const h = this.hash(`${domainScope}_${clean}`);
    const uuid = `${h.substr(0,8)}-${h.substr(8,4)}-4${h.substr(13,3)}-a${h.substr(17,3)}-${h.substr(20,12)}`;
    return `'${uuid}'`;
  }

  maskGeneric(rule, rawValue) {
    if (!rawValue || rawValue === 'NULL') return rawValue;
    const type = rule.type.toLowerCase();

    switch (type) {
      case 'email':
        return this.maskEmail(rawValue);
      case 'name':
        return this.maskName(rawValue);
      case 'phone':
        return this.maskPhone(rawValue);
      case 'nik':
        return this.maskNik(rawValue);
      case 'address':
        return this.maskAddress(rawValue);
      case 'password_hash':
        return this.maskPasswordHash();
      case 'numeric_variance':
        return this.maskNumericVariance(rawValue, rule.variance || 15);
      case 'fk_id':
      case 'pk_id':
        return this.maskForeignOrPrimaryKey(rawValue, rule.domain || 'global_id');
      case 'uuid':
        return this.maskUuid(rawValue, rule.domain || 'global_uuid');
      case 'redact':
        return `'[REDACTED]'`;
      case 'null':
        return 'NULL';
      default:
        return rawValue;
    }
  }
}

module.exports = { Generators };
