function loginPage() {
  return {
    name: '',
    password: '',
    loading: false,
    error: '',
    async submit() {
      this.loading = true; this.error = '';
      try {
        var body = { name: this.name, password: this.password };
        var resp = await api('/api/auth/login', {
          method: 'POST',
          body: JSON.stringify(body),
        });
        Alpine.store('auth').login(resp.data.token);
        var anonToken = sessionStorage.getItem('anon_token');
        if (anonToken) {
          try {
            await api('/api/assessments/claim', {
              method: 'POST',
              body: JSON.stringify({ anonymous_token: anonToken }),
            });
          } catch (e) {}
          sessionStorage.removeItem('anon_token');
        }
        var redirect = new URLSearchParams(location.search).get('redirect');
        if (!redirect) {
          redirect = localePath('/' + encodeURIComponent(resp.data.user.name));
        }
        if (!/^\/(en|zh-CN)\//.test(redirect)) redirect = localePath(redirect);
        window.location = redirect;
      } catch (e) {
        this.error = e.message || Alpine.store('locale').t('error.loginFailed');
      } finally { this.loading = false; }
    },
  };
}
