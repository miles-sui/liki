// result.js — Assessment result page with radar + river charts
function resultPage() {
  return {
    identity: null,
    dValues: [],
    pValues: [],
    loading: true,
    revealed: false,

    init() {
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
          this.dValues = namedToArray(data.profile ? data.profile.d : null);
          this.pValues = namedToArray(data.profile ? data.profile.p : null);
          if (data.identity) {
            this.identity = {
              label: data.identity.label || data.identity.id,
              id: data.identity.id,
              category: data.identity.category,
            };
          }
        }
      } catch (e) {}
      this.loading = false;
      this.afterLoad();
    },

    loadFromParams(params) {
      try {
        this.dValues = namedToArray(params.get('d') ? JSON.parse(params.get('d')) : null);
        this.pValues = namedToArray(params.get('p') ? JSON.parse(params.get('p')) : null);
        if (params.get('label')) {
          this.identity = {
            label: params.get('label'),
            id: params.get('id'),
            category: params.get('category'),
          };
        }
      } catch (e) {}
      this.loading = false;
      this.afterLoad();
    },

    afterLoad() {
      if (this.pValues.length) {
        this.$nextTick(() => {
          this.renderRadar();
          this.renderRiver();
          setTimeout(() => { this.revealed = true; }, 600);
        });
      } else {
        this.revealed = true;
      }
    },

    renderRadar() {
      var el = document.getElementById('result-radar');
      if (!el || !this.pValues.length || !window.Charts) return;
      var pVals = this.pValues.slice();
      Charts.renderElementRadar(el, {
        series: [{
          value: pVals, name: this.identity ? this.identity.label : 'p',
          symbolSize: 8,
          animationDuration: 600, animationEasing: 'cubicOut',
        }],
      });
    },

    renderRiver() {
      var el = document.getElementById('result-river');
      if (!el || !this.pValues.length || !window.Charts) return;

      var pVals = this.pValues.slice();
      var displayNames = ELEMENT_DISPLAY_ORDER.map(function(idx) {
        return window.ELEMENT_NAMES[window.ELEMENT_CODES[idx]];
      });

      Charts.renderElementLine(el, {
        categories: displayNames,
        series: ELEMENT_DISPLAY_ORDER.map(function(idx) {
          return { colorIdx: idx, data: [pVals[idx]], symbolSize: 10 };
        }),
      }, {
        xAxis: { data: displayNames, boundaryGap: true, axisLabel: { fontSize: 12 } },
        yAxis: { name: 'p', nameTextStyle: { fontSize: 10 } },
        legend: { show: false },
        grid: { left: 36, right: 16, top: 20, bottom: 24 },
      });
    },

    copyShareLink() {
      var url = location.href;
      navigator.clipboard.writeText(url).then(() => {
        Alpine.store('toast').success(Alpine.store('locale').t('toast.linkCopied'));
      }).catch(() => {
        Alpine.store('toast').error(Alpine.store('locale').t('toast.copyFailed'));
      });
    },
  };
}
