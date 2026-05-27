const yaml = require('js-yaml');
const fs = require('fs');
const path = require('path');

const locale = process.env.LOCALE || 'en';
const contentDir = path.join(__dirname, '..', 'content', locale);

module.exports = yaml.load(fs.readFileSync(
  path.join(contentDir, 'mingli-templates.yaml'), 'utf8'));
