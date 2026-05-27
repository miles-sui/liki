const crypto = require('crypto');
const fs = require('fs');
const path = require('path');

const distDir = path.join(__dirname, '..', 'dist');
const originals = new Set(['alpine.min.js', 'common.js', 'tailwind.min.css']);
const hashes = {};

function fingerprintDir(subdir, prefix) {
  const dir = path.join(distDir, subdir);
  if (!fs.existsSync(dir)) return;

  for (const fn of fs.readdirSync(dir)) {
    if (!originals.has(fn)) continue;
    const ext = path.extname(fn);
    const stem = fn.slice(0, -ext.length);
    const fpath = path.join(dir, fn);
    const hash = crypto.createHash('sha256').update(fs.readFileSync(fpath)).digest('hex').slice(0, 8);
    const newName = `${stem}.${hash}${ext}`;
    if (newName === fn) continue;

    fs.renameSync(fpath, path.join(dir, newName));
    hashes[`${prefix}${fn}`] = `${prefix}${newName}`;
    console.log(`  Fingerprinted: ${fn} -> ${newName}`);
  }
}

function updateHtml() {
  for (const root of [distDir, ...subdirs(distDir)]) {
    for (const fn of fs.readdirSync(root)) {
      if (!fn.endsWith('.html')) continue;
      const fpath = path.join(root, fn);
      let html = fs.readFileSync(fpath, 'utf8');
      let changed = false;
      for (const [old, new_] of Object.entries(hashes)) {
        if (html.includes(old)) {
          html = html.replaceAll(old, new_);
          changed = true;
        }
      }
      if (changed) fs.writeFileSync(fpath, html);
    }
  }
}

function subdirs(dir) {
  const result = [];
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    if (entry.isDirectory()) {
      result.push(path.join(dir, entry.name));
      result.push(...subdirs(path.join(dir, entry.name)));
    }
  }
  return result;
}

fingerprintDir('js', '/js/');
fingerprintDir('js/vendor', '/js/vendor/');
fingerprintDir('css', '/css/');
updateHtml();
