// result.js — Assessment result page
function resultPage() {
  return {
    identity: null,
    pValues: [],
    loading: true,
    revealed: false,
    closestTypes: [],
    typeDescription: '',

    init() {
      var auth = Alpine.store('auth');
      if (auth && auth.id) {
        window.location.href = '/app#/overview';
        return;
      }
      var params = new URLSearchParams(location.search);
      var aid = params.get('aid');
      if (aid) {
        this.loadFromAPI(aid);
      } else {
        this.loadFromParams(params);
      }
    },

    async loadFromAPI(aid) {
      try {
        var r = await api('/api/assessments/' + aid);
        var data = r.data;
        if (data) {
          this.pValues = namedToArray(data.profile ? data.profile.p : null);
          if (data.identity) {
            this.identity = {
              label: data.identity.label || data.identity.id,
              id: data.identity.id,
              category: data.identity.category,
            };
          }
        }
      } catch (e) { console.error(e); }
      this.loading = false;
      this.afterLoad();
    },

    loadFromParams(params) {
      try {
        this.pValues = namedToArray(params.get('p') ? JSON.parse(params.get('p')) : null);
        if (params.get('label')) {
          this.identity = {
            label: params.get('label'),
            id: params.get('id'),
            category: params.get('category'),
          };
        }
      } catch (e) { console.error(e); }
      this.loading = false;
      this.afterLoad();
    },

    afterLoad() {
      if (this.pValues.length) {
        this.$nextTick(() => {
          this.renderRadar();
          setTimeout(() => { this.revealed = true; }, 600);
        });
        this.loadTypeDescription();
        this.computeDistances();
      } else {
        this.revealed = true;
      }
    },

    loadTypeDescription() {
      if (!this.identity || !window.TYPES) return;
      var t = window.TYPES.find(function(ty) { return ty.id === this.identity.id; }.bind(this));
      if (t && t.desc) this.typeDescription = t.desc;
      else if (t && t.tagline) this.typeDescription = t.tagline;
    },

    async computeDistances() {
      if (!this.pValues.length) return;
      try {
        var resp = await fetch('/content/prototypes.json');
        var protoMap = await resp.json();
        var self = this;
        var scored = [];
        Object.keys(protoMap).forEach(function(id) {
          var d = protoMap[id];
          // Convert deviation vector to proportion for comparison
          var p = d.map(function(v) { return v + 0.2; });
          var score = cosineSimilarity(self.pValues, p);
          var typeInfo = window.TYPES ? window.TYPES.find(function(t) { return t.id === id; }) : null;
          scored.push({
            id: id,
            label: typeInfo ? typeInfo.label : id,
            score: score,
          });
        });
        scored.sort(function(a, b) { return b.score - a.score; });
        // Remove exact self-match if present, then take top 3
        var filtered = scored.filter(function(t) { return t.id !== (self.identity && self.identity.id); });
        this.closestTypes = filtered.slice(0, 3);
      } catch (e) { console.error('Failed to compute distances:', e); }
    },

    async renderRadar() {
      var el = document.getElementById('result-radar');
      if (!el || !this.pValues.length || !window.Charts) return;
      var pVals = this.pValues.slice();
      if (window._loadECharts) await window._loadECharts();
      this.radarInst = Charts.renderElementRadar(el, {
        series: [{
          value: pVals, name: this.identity ? this.identity.label : 'p',
          symbolSize: 8,
          animationDuration: 600, animationEasing: 'cubicOut',
        }],
      });
    },

    copyShareLink() {
      copyText(location.href, { successKey: 'toast.linkCopied' });
    },

    async generateShareCard() {
      await Charts.generateAndSaveShareCard(this.radarInst, this.identity, this.pValues.slice());
    },
  };
}

// Cosine similarity between two arrays
function cosineSimilarity(a, b) {
  var dot = 0, normA = 0, normB = 0;
  for (var i = 0; i < a.length; i++) {
    dot += a[i] * b[i];
    normA += a[i] * a[i];
    normB += b[i] * b[i];
  }
  if (normA === 0 || normB === 0) return 0;
  return dot / (Math.sqrt(normA) * Math.sqrt(normB));
}
