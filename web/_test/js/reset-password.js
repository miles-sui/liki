function resetPage() {
  return {
    password: '', loading: false, error: '', success: false,
    async submit() {
      this.loading = true; this.error = '';
      try {
        var token = new URLSearchParams(location.search).get('token');
        await api('/api/auth/reset-password', {
          method: 'POST',
          body: JSON.stringify({ token: token, password: this.password }),
        });
        this.success = true;
      } catch (e) {
        this.error = e.message || Alpine.store('locale').t('error.failed');
      } finally { this.loading = false; }
    },
  };
}
