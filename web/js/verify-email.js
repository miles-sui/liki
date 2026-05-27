function verifyPage() {
  return {
    loading: true, ok: false,
    async verify() {
      try {
        var token = new URLSearchParams(location.search).get('token');
        await api('/api/auth/verify-email?token=' + token);
        this.ok = true;
      } catch (e) { console.error(e); }
      this.loading = false;
    },
  };
}
