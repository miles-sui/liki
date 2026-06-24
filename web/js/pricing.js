// Prepend currency symbol based on geo-IP: ¥ for CN, $ otherwise.
// Waits for i18n:loaded; falls back to 3s timeout if i18n never fires.
(function(){
  var applied = false;

  function show(sym) {
    document.querySelectorAll('.pricing').forEach(function(el){
      el.textContent = sym + el.textContent.trim();
    });
  }

  function revealAll() {
    document.querySelectorAll('.pricing').forEach(function(el){ el.style.visibility = 'visible'; });
  }

  function doApply() {
    if (applied) return;
    applied = true;
    fetch('/api/location').then(function(r){ return r.json(); }).then(function(d){
      show((d.data && d.data.currency === 'CNY') ? '¥' : '$');
    }).finally(revealAll);
  }

  function apply() {
    if (typeof i18next !== 'undefined' && i18next.isInitialized) return doApply();
    document.addEventListener('i18n:loaded', doApply, { once: true });
    setTimeout(function(){ doApply(); }, 3000);
  }

  apply();
})();
