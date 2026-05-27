function typesPage(typesData) {
  return {
    types: typesData,
    filter: 'all',

    get filteredTypes() {
      if (this.filter === 'all') return this.types;
      return this.types.filter((t) => t.id[0] === this.filter);
    },

    init() {
      var self = this;
      setTimeout(function () { self.buildMatrix(); }, 0);
    },

    buildMatrix() {
      var el = document.getElementById('types-matrix');
      if (!el) return;

      var codes = window.ELEMENT_CODES;
      var names = window.ELEMENT_NAMES;
      var loc = window.CURRENT_LOCALE || 'en';
      var colors = Alpine.store('theme').chartColors;
      var col = {};
      for (var i = 0; i < 5; i++) col[codes[i]] = colors[i];

      var typeMap = {};
      for (var i = 0; i < this.types.length; i++) {
        typeMap[this.types[i].id] = this.types[i];
      }

      function cellHTML(rid, cid) {
        var isPure = rid === cid;
        var id = isPure ? codes[rid] : codes[rid] + codes[cid];
        var t = typeMap[id];
        var label = t ? (t.label || id) : id;
        var c = col[codes[rid]];
        var bg = isPure ? c : c + '15';
        var fg = isPure ? '#fff' : c;
        return '<a href="/' + loc + '/types/' + id + '"' +
          ' class="block rounded-md px-2 py-1.5 text-center transition hover:brightness-110"' +
          ' style="background:' + bg + ';color:' + fg + '">' +
          '<span class="text-sm font-semibold">' + id + '</span>' +
          '<span class="block text-xs opacity-70 leading-tight truncate">' + label + '</span></a>';
      }

      var html = '<table class="w-full text-center" style="border-collapse:separate;border-spacing:2px">';
      html += '<thead><tr><th class="p-1"></th>';
      for (var ci = 0; ci < 5; ci++) {
        html += '<th class="p-1 text-xs font-medium" style="color:' + col[codes[ci]] + '">' + names[codes[ci]] + '</th>';
      }
      html += '</tr></thead><tbody>';

      for (var ri = 0; ri < 5; ri++) {
        html += '<tr>';
        html += '<th class="p-1 text-xs font-medium text-right" style="color:' + col[codes[ri]] + '">' + names[codes[ri]] + '</th>';
        for (var ci = 0; ci < 5; ci++) {
          html += '<td class="min-w-[72px]">' + cellHTML(ri, ci) + '</td>';
        }
        html += '</tr>';
      }
      html += '</tbody></table>';

      el.innerHTML = html;
    }
  };
}
