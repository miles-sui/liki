const { chromium } = require('playwright-core');
(async () => {
  const browser = await chromium.launch({ executablePath: '/opt/google/chrome/chrome' });
  const page = await browser.newPage();
  const errors = [];
  page.on('console', (msg) => { if (msg.type() === 'error') errors.push(msg.text()); });
  page.on('pageerror', (err) => errors.push(err.message));

  await page.goto('http://localhost:8080/en/', { waitUntil: 'networkidle', timeout: 15000 });
  await page.waitForTimeout(3000);

  console.log('=== JS Errors ===');
  if (errors.length === 0) console.log('(none)');
  else errors.forEach(e => console.log('  ERR:', e.substring(0, 150)));

  // Check all chart canvases
  console.log('\n=== Chart containers (canvas inside div) ===');
  const chartIds = [
    'radar-0', 'radar-1', 'radar-2',          // profile mini radars
    'example-flow-radar',                       // flow dual radar
    'example-river-chart',                      // river line chart
    'peer-compare-radar',                       // peer dual radar
    'bond-influence-radar-a', 'bond-influence-radar-b', // bond radars
  ];
  for (const id of chartIds) {
    const el = await page.$('#' + id);
    if (!el) { console.log('  ' + id + ': container MISSING'); continue; }
    const rect = await el.boundingBox();
    const size = rect ? Math.round(rect.width) + 'x' + Math.round(rect.height) : 'no rect';
    // Check if canvas child exists and has dimensions
    const canvas = await el.$('canvas');
    const canvasInfo = canvas ? 'has canvas' : 'NO CANVAS';
    const visible = await el.isVisible();
    console.log('  ' + id + ': visible=' + visible + ' size=' + size + ' ' + canvasInfo);
  }

  // Check delta bars content
  console.log('\n=== Delta bars ===');
  const deltas = [
    ['flow deltas', await page.locator('#example-flow-radar').locator('..').locator('.space-y-1').first().innerHTML().catch(() => 'MISSING')],
    ['peer deltas', await page.locator('#peer-compare-deltas').innerHTML().catch(() => 'MISSING')],
    ['bond deltas A', await page.locator('#bond-influence-radar-a').locator('..').locator('[x-data]').innerHTML().catch(() => '')], // check via x-data
  ];
  for (const [name, html] of deltas) {
    console.log('  ' + name + ': ' + (typeof html === 'string' ? html.length + ' chars' : 'MISSING'));
  }

  // Try to read chart options from ECharts instances
  console.log('\n=== ECharts instances (series count, data check) ===');
  const chartData = await page.evaluate(() => {
    const result = [];
    const ids = ['radar-0', 'radar-1', 'radar-2', 'example-flow-radar', 'example-river-chart', 'peer-compare-radar', 'bond-influence-radar-a', 'bond-influence-radar-b'];
    for (const id of ids) {
      const el = document.getElementById(id);
      if (!el) { result.push({id, error: 'no element'}); continue; }
      const instance = window.echarts && window.echarts.getInstanceByDom ? window.echarts.getInstanceByDom(el) : null;
      if (!instance) { result.push({id, error: 'no echarts instance'}); continue; }
      const opt = instance.getOption();
      // Extract key info
      const series = (opt.series || []).map(s => ({
        name: s.name,
        type: s.type,
        dataLength: Array.isArray(s.data) ? s.data.length : (s.data ? Object.keys(s.data).length : '?'),
        hasLineStyle: !!s.lineStyle,
        hasAreaStyle: !!s.areaStyle,
      }));
      const legend = opt.legend ? opt.legend[0] : null;
      const legendData = legend ? legend.data : [];
      result.push({id, series, legendData});
    }
    return result;
  });

  for (const cd of chartData) {
    console.log('  ' + cd.id + ':');
    if (cd.error) { console.log('    ERROR: ' + cd.error); continue; }
    if (cd.legendData.length) console.log('    legend: [' + cd.legendData.join(', ') + ']');
    for (const s of cd.series) {
      console.log('    series: type=' + s.type + ' name="' + (s.name||'') + '" dataLen=' + s.dataLength + ' area=' + (s.hasAreaStyle ? 'Y' : 'N'));
    }
  }

  await browser.close();
})();
