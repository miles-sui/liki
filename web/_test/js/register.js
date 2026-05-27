function registerPage() {
  return {
    name: '',
    password: '',
    loading: false,
    error: '',
    async submit() {
      this.loading = true; this.error = '';
      try {
        var body = { name: this.name, password: this.password };
        var anonToken = sessionStorage.getItem('anon_token');
        if (anonToken) body.anonymous_token = anonToken;
        var resp = await api('/api/auth/register', {
          method: 'POST',
          body: JSON.stringify(body),
        });
        sessionStorage.removeItem('anon_token');
        Alpine.store('auth').login(resp.data.token);
        // Redirect to profile page (user now has an account but may not have assessment)
        window.location = localePath('/' + encodeURIComponent(resp.data.user.name));
      } catch (e) {
        this.error = e.message || Alpine.store('locale').t('error.registerFailed');
      } finally { this.loading = false; }
    },
  };
}
