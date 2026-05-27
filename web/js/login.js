function loginPage() {
  return {
    name: '',
    password: '',
    showPw: false,
    loading: false,
    error: '',
    async handleSubmit() {
      this.loading = true; this.error = '';
      try {
        var body = { name: this.name, password: this.password };
        var resp = await api('/api/auth/login', {
          method: 'POST',
          body: JSON.stringify(body),
        });
        Alpine.store('auth').login(resp.data.token, resp.data.user);
        var anonToken = sessionStorage.getItem('anon_token');
        if (anonToken) {
          try {
            await api('/api/assessments/claim', {
              method: 'POST',
              body: JSON.stringify({ anonymous_token: anonToken }),
            });
          } catch (e) { console.error(e); }
          sessionStorage.removeItem('anon_token');
        }
        var redirect = new URLSearchParams(location.search).get('redirect');
        if (!redirect) {
          redirect = '/app';
        }
        // Map old SSG paths to SPA hash routes (superseded pages).
        var spaMap = { profile: '#overview', bonds: '#bonds', settings: '#settings' };
        var clean = redirect.replace(/^\/(en|zh-CN)\//, '/').replace(/\.html$/, '');
        if (spaMap[clean.slice(1)]) {
          redirect = '/app' + spaMap[clean.slice(1)];
        }
        if (!/^\/(en|zh-CN)\//.test(redirect) && redirect !== '/app' && !/^\/app#/.test(redirect)) redirect = localePath(redirect);
        window.location = redirect;
      } catch (e) {
        console.error(e);
        this.error = e.message || Alpine.store('locale').t('error.loginFailed');
      } finally { this.loading = false; }
    },
  };
}
