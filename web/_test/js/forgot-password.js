function forgotPasswordPage() {
  return {
    email: '',
    loading: false,
    sent: false,
    async submit() {
      this.loading = true;
      try {
        await api('/api/auth/forgot-password', {
          method: 'POST',
          body: JSON.stringify({ email: this.email }),
        });
        this.sent = true;
      } catch (_) {} finally { this.loading = false; }
    },
  };
}
