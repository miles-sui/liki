const yaml = require('js-yaml');
const fs = require('fs');
const path = require('path');

const locale = process.env.LOCALE || 'en';
const contentDir = path.join(__dirname, '..', 'content', locale);

const summary = yaml.load(fs.readFileSync(
  path.join(contentDir, 'types.yaml'), 'utf8'));

const typesDir = path.join(contentDir, 'types');

module.exports = summary.types.map((t) => {
  const detailPath = path.join(typesDir, `${t.id}.yaml`);
  if (fs.existsSync(detailPath)) {
    const detail = yaml.load(fs.readFileSync(detailPath, 'utf8'));
    return { ...t, ...detail };
  }
  return t;
});
