const fs = require('fs');
const path = require('path');

const rootDir = path.join(__dirname, '..');
const enJson = path.join(rootDir, 'content', 'en', 'translations.json');

if (!fs.existsSync(enJson)) {
  console.log('  (en/translations.json not found, skipping translation check)');
  process.exit(0);
}

const uiData = JSON.parse(fs.readFileSync(enJson, 'utf8'));
const definedKeys = new Set(Object.keys(uiData));

// Scan source files (templates, JS) for t('key') patterns
const usedKeys = new Set();
function scanDir(dir) {
  if (!fs.existsSync(dir)) return;
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) { if (entry.name !== 'node_modules' && entry.name !== 'vendor') scanDir(full); continue; }
    if (!entry.name.endsWith('.html') && !entry.name.endsWith('.njk') && !entry.name.endsWith('.js')) continue;
    const text = fs.readFileSync(full, 'utf8');
    // \$store.locale.t('key') and locale.t('key') and .t('key')
    // \$store.locale.t('key'), Alpine.store('locale').t('key'), locale.t('key')
    for (const m of text.matchAll(/\.t\('([^']+)'\)/g)) {
      // Exclude Nunjucks templates with unresolved {{ var }}
      if (!m[1].includes('{{')) usedKeys.add(m[1]);
    }
    // Nunjucks set variables: {% set labelKey = 'key' %}
    for (const m of text.matchAll(/\{%[-]?\s*set\s+(?:labelKey|messageKey|titleKey)\s*=\s*'([^']+)'\s*[-]?%\}/g)) {
      usedKeys.add(m[1]);
    }
  }
}
scanDir(path.join(rootDir, 'pages'));
scanDir(path.join(rootDir, '_includes'));
scanDir(path.join(rootDir, '_layouts'));
scanDir(path.join(rootDir, 'js'));

const missing = [...usedKeys].filter(k => !definedKeys.has(k)).sort();
const unused = [...definedKeys].filter(k => !usedKeys.has(k)).sort();

if (missing.length) {
  console.log(`  MISSING translation keys (${missing.length}):`);
  for (const k of missing) console.log(`    - ${k}`);
} else {
  console.log('  All used translation keys defined.');
}
if (unused.length) {
  console.log(`  Unused translation keys (${unused.length}):`);
  for (const k of unused) console.log(`    - ${k}`);
}
