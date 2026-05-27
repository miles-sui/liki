function registerPage() {
  return {
    name: '',
    email: '',
    password: '',
    showPw: false,
    get strength() { return passwordStrength(this.password); },

    barClass(n) { return pwBarClass(this.strength.score, n); },

    loading: false,
    error: '',
    async handleSubmit() {
      this.loading = true; this.error = '';
      try {
        var body = { name: this.name, email: this.email, password: this.password };
        var anonToken = sessionStorage.getItem('anon_token');
        if (anonToken) body.anonymous_token = anonToken;
        var resp = await api('/api/auth/register', {
          method: 'POST',
          body: JSON.stringify(body),
        });
        sessionStorage.removeItem('anon_token');
        Alpine.store('auth').login(resp.data.token, resp.data.user);
        window.location = '/app';
      } catch (e) {
        console.error(e);
        this.error = e.message || Alpine.store('locale').t('error.registerFailed');
      } finally { this.loading = false; }
    },
  };
}
