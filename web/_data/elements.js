const fs = require('fs');
const path = require('path');
const locale = process.env.LOCALE || 'en';

module.exports = JSON.parse(fs.readFileSync(
  path.join(__dirname, '..', 'content', locale, 'elements.json'), 'utf8'));
