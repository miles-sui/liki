const fs = require('fs');
const path = require('path');
const MarkdownIt = require('markdown-it');

const md = new MarkdownIt({ html: true });
const locale = process.env.LOCALE || 'en';
const contentDir = path.join(__dirname, '..', 'content', locale);
const pages = ['about', 'privacy', 'terms', 'faq', 'cookies', 'refund'];

const result = {};
for (const page of pages) {
  const filePath = path.join(contentDir, `${page}.md`);
  try {
    result[page] = md.render(fs.readFileSync(filePath, 'utf8'));
  } catch (e) {
    if (e.code !== 'ENOENT') throw e;
  }
}

module.exports = { json: JSON.stringify(result) };
