#!/usr/bin/env node
// check-i18n.cjs — Validate i18n consistency across locales.
// Checks: key alignment, price format consistency, untranslated fallbacks.

const fs = require('fs');
const path = require('path');

const I18N_DIR = path.resolve(__dirname, '..', 'i18n');
const errors = [];

function loadJSON(filepath) {
  try {
    return JSON.parse(fs.readFileSync(filepath, 'utf-8'));
  } catch (e) {
    errors.push(`${path.basename(filepath)}: invalid JSON — ${e.message}`);
    return null;
  }
}

function checkKeyAlignment(locales) {
  const names = Object.keys(locales);
  for (let i = 0; i < names.length; i++) {
    for (let j = i + 1; j < names.length; j++) {
      const a = names[i], b = names[j];
      const keysA = new Set(Object.keys(locales[a]));
      const keysB = new Set(Object.keys(locales[b]));

      for (const k of keysA) {
        if (!keysB.has(k)) errors.push(`${b}.json: missing key "${k}" (present in ${a}.json)`);
      }
      for (const k of keysB) {
        if (!keysA.has(k)) errors.push(`${a}.json: missing key "${k}" (present in ${b}.json)`);
      }
    }
  }
}

function checkPrices(locales) {
  // Check that price-format keys (containing "price" or "unlock" or "ctaDesc")
  // use consistent currency markers. All should use USD ($) now.
  const priceKeyRe = /price|unlock|ctaDesc/i;
  const currencyRe = /([¥£€]|HK\$|CNY|HKD|USD)/g;

  for (const [locale, data] of Object.entries(locales)) {
    for (const [key, value] of Object.entries(data)) {
      if (typeof value !== 'string') continue;
      if (!priceKeyRe.test(key)) continue;

      const matches = value.match(currencyRe);
      if (matches) {
        for (const m of matches) {
          if (m !== '$' || (m === '$' && value.includes('HK$'))) {
            // Allow $ only if it's not HK$
            if (m === '$' && !value.includes('HK$') && !value.includes('HKD')) continue;
          }
          errors.push(`${locale}.json: "${key}" contains non-USD currency "${m}" — use "$" only`);
        }
      }
    }
  }
}

function checkHtmlKeys(locales) {
  // Extract all data-i18n keys from HTML files and verify they exist in all locales.
  const allKeys = new Set();
  for (const [name, data] of Object.entries(locales)) {
    for (const k of Object.keys(data)) allKeys.add(k);
  }

  const htmlDir = path.resolve(__dirname, '..');
  const htmlFiles = fs.readdirSync(htmlDir).filter(f => f.endsWith('.html'));
  const keyRe = /data-i18n="([^"]+)"/g;

  for (const file of htmlFiles) {
    const html = fs.readFileSync(path.join(htmlDir, file), 'utf-8');
    let m;
    while ((m = keyRe.exec(html)) !== null) {
      const key = m[1];
      if (!allKeys.has(key)) {
        errors.push(`${file}: data-i18n="${key}" not found in any locale file`);
      }
    }
  }
}

function checkEmptyValues(locales) {
  for (const [locale, data] of Object.entries(locales)) {
    for (const [key, value] of Object.entries(data)) {
      if (value === '' || value === null || value === undefined) {
        errors.push(`${locale}.json: "${key}" is empty`);
      }
    }
  }
}

function main() {
  const files = fs.readdirSync(I18N_DIR).filter(f => f.endsWith('.json'));
  if (files.length < 2) {
    console.log('Need at least 2 locale files to compare.');
    process.exit(0);
  }

  const locales = {};
  for (const file of files) {
    const name = file.replace('.json', '');
    const data = loadJSON(path.join(I18N_DIR, file));
    if (data) locales[name] = data;
  }

  checkKeyAlignment(locales);
  checkPrices(locales);
  checkEmptyValues(locales);
  checkHtmlKeys(locales);

  if (errors.length > 0) {
    console.error(`\n${errors.length} i18n violation(s):\n`);
    for (const e of errors) console.error(`  ✗ ${e}`);
    console.error();
    process.exit(1);
  }

  console.log(`✓ ${Object.keys(locales).length} locale(s) checked, no violations.`);
}

main();
