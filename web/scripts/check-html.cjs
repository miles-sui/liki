#!/usr/bin/env node
// check-html.js — Validate HTML files for framework rule violations.
// Vue:  v-text / v-html must not have child content (compiler error #57).
// Alpine: every page with Alpine directives must have x-data.
// Mixing: a page using Vue must not use Alpine directives.

const fs = require('fs');
const path = require('path');

const WEB_DIR = path.resolve(__dirname, '..');
const errors = [];

// Detect framework from <script> tags.
function detectFramework(html, filename) {
  const hasVue = /vue\.global/.test(html);
  const hasAlpine = /alpine\.min\.js/.test(html);
  if (hasVue && hasAlpine) return 'mixed';
  if (hasVue) return 'vue';
  if (hasAlpine) return 'alpine';
  return 'none';
}

// Check Vue rules: v-text / v-html must not have child content (compiler error #57).
function checkVue(html, filename) {
  const lines = html.split('\n');
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const matches = line.matchAll(/<(\w+)[^>]*\s(v-text|v-html)="[^"]*"[^>]*>/g);
    for (const m of matches) {
      const tagName = m[1];
      const dir = m[2];
      const afterTag = line.substring(m.index + m[0].length);

      // If self-closed on the same line (></tag> or />), fine.
      if (/^\s*<\/\w+/.test(afterTag) || /^\s*\/>/.test(afterTag)) continue;

      // If there's non-whitespace inline text after the tag, that's child content.
      const trimmed = afterTag.trimStart();
      if (trimmed && !trimmed.startsWith('<')) {
        errors.push(`${filename}:${i + 1}: ${dir} on <${tagName}> has inline child content — use {{ }} interpolation instead`);
        continue;
      }

      // If the closing tag is not on this line, there's child content (elements or text).
      if (!new RegExp(`<\/${tagName}>`).test(line.substring(m.index + m[0].length))) {
        errors.push(`${filename}:${i + 1}: ${dir} on <${tagName}> has child content on following lines — use {{ }} interpolation instead`);
      }
    }
  }
}

// Check Alpine rules.
function checkAlpine(html, filename) {
  const alpineDirectives = /\bx-(text|show|html|cloak|for|if|model|bind|on|data|ref|effect|init|ignore|transition)\b/;
  const eventDirectives = /@(click|keydown|submit|input|change|mouseenter|mouseleave|focus|blur)\b/;

  const hasAlpineDirectives = alpineDirectives.test(html) || eventDirectives.test(html);

  if (!hasAlpineDirectives) return;

  // Check x-data exists on body or another element.
  if (!/\bx-data\b/.test(html)) {
    errors.push(`${filename}: Alpine directives found but no x-data on any element (Alpine will not initialize)`);
    return;
  }

  // Check x-data is on a container ancestor (body or root div).
  // Find all x-data elements and check they wrap Alpine directive elements.
  const xdataRe = /<(\w+)[^>]*\sx-data[^>]*>/g;
  const lines = html.split('\n');
  const xdataLines = [];
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const matches = line.matchAll(xdataRe);
    for (const m of matches) {
      xdataLines.push(i + 1);
    }
  }

  // Check Alpine load order: alpine.min.js must be before api.js, report.js, demo-utils.js
  const scripts = [...html.matchAll(/<script[^>]*src="([^"]*)"[^>]*>/g)];
  let alpineIdx = -1;
  let firstDepIdx = Infinity;
  for (let i = 0; i < scripts.length; i++) {
    const src = scripts[i][1];
    if (src.includes('alpine')) alpineIdx = i;
    if (/api\.js|report\.js|demo-utils\.js/.test(src) && !src.includes('alpine')) {
      if (firstDepIdx === Infinity) firstDepIdx = i;
    }
  }
  if (alpineIdx >= 0 && firstDepIdx < alpineIdx) {
    // Scripts that use alpine:init event listener register components before
    // Alpine initializes, so Alpine must load last — reverse of normal order.
    let hasAlpineInitListener = false;
    for (let i = firstDepIdx; i <= alpineIdx && i < scripts.length; i++) {
      const src = scripts[i][1];
      const jsPath = path.join(WEB_DIR, src.replace(/^\//, ''));
      if (fs.existsSync(jsPath)) {
        const jsContent = fs.readFileSync(jsPath, 'utf-8');
        if (/addEventListener\s*\(\s*['"]alpine:init['"]/.test(jsContent)) {
          hasAlpineInitListener = true;
          break;
        }
      }
    }
    if (!hasAlpineInitListener) {
      const line = html.substring(0, scripts[firstDepIdx].index).split('\n').length;
      errors.push(`${filename}:${line}: Alpine.js must load before scripts that depend on Alpine (api.js, report.js)`);
    }
  }
}

// Check framework mixing.
function checkMixing(html, filename, framework) {
  if (framework === 'vue') {
    // Vue page should not have Alpine directives
    if (/\bx-(text|show|html|for|if|model|bind|on)\b/.test(html)) {
      errors.push(`${filename}: Vue page has Alpine directives (x-text, x-show, etc.) — pick one framework`);
    }
  }
}

// Check absolute paths in <script src> and <link href>.
function checkPaths(html, filename) {
  const lines = html.split('\n');
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    // Check script src
    const scriptMatch = line.match(/<script[^>]*src="(js\/[^"]+)"/);
    if (scriptMatch) {
      errors.push(`${filename}:${i + 1}: relative script path "${scriptMatch[1]}" — use absolute "/${scriptMatch[1]}"`);
    }
    // Check link href
    const linkMatch = line.match(/<link[^>]*href="(css\/[^"]+)"/);
    if (linkMatch) {
      errors.push(`${filename}:${i + 1}: relative link path "${linkMatch[1]}" — use absolute "/${linkMatch[1]}"`);
    }
  }
}

// Check JS files for sourceMappingURL referencing non-existent .map files.
function checkSourceMaps() {
  const jsDir = path.join(WEB_DIR, 'js');
  if (!fs.existsSync(jsDir)) return;
  const files = fs.readdirSync(jsDir).filter(f => f.endsWith('.js'));
  for (const file of files) {
    const filepath = path.join(jsDir, file);
    const content = fs.readFileSync(filepath, 'utf-8');
    const lines = content.split('\n');
    for (let i = 0; i < lines.length; i++) {
      const m = lines[i].match(/^\/\/#\s*sourceMappingURL=(.+)$/);
      if (m) {
        const mapFile = m[1].trim();
        const mapPath = path.join(jsDir, mapFile);
        if (!fs.existsSync(mapPath)) {
          errors.push(`${file}:${i + 1}: sourceMappingURL references "${mapFile}" which does not exist — remove this line`);
        }
      }
    }
  }
}

// Main.
function main() {
  const files = fs.readdirSync(WEB_DIR).filter(f => f.endsWith('.html'));

  if (files.length === 0) {
    console.log('No HTML files found.');
    process.exit(0);
  }

  for (const file of files) {
    const filepath = path.join(WEB_DIR, file);
    const html = fs.readFileSync(filepath, 'utf-8');
    const framework = detectFramework(html, file);

    if (framework === 'vue') checkVue(html, file);
    if (framework === 'alpine') checkAlpine(html, file);
    checkMixing(html, file, framework);
    checkPaths(html, file);
  }
  checkSourceMaps();

  if (errors.length > 0) {
    console.error(`\n${errors.length} violation(s):\n`);
    for (const e of errors) console.error(`  ✗ ${e}`);
    console.error();
    process.exit(1);
  }

  console.log(`✓ ${files.length} HTML file(s) checked, no violations.`);
}

main();
