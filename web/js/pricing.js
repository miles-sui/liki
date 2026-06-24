// Show pricing with currency symbol based on IP geo (¥ for CN, $ for others).
fetch('/api/location').then(r => r.json()).then(d => {
  const sym = (d.data && d.data.currency === 'CNY') ? '¥' : '$';
  document.querySelectorAll('.pricing').forEach(el => {
    el.textContent = sym + el.textContent.trim();
    el.style.visibility = 'visible';
  });
}).catch(() => {
  document.querySelectorAll('.pricing').forEach(el => el.style.visibility = 'visible');
});
