function forgotPasswordPage() {
  return {
    email: '',
    loading: false,
    sent: false,
    async handleSubmit() {
      this.loading = true;
      try {
        await api('/api/auth/forgot-password', {
          method: 'POST',
          body: JSON.stringify({ email: this.email }),
        });
        this.sent = true;
      } catch (e) { console.error(e); } finally { this.loading = false; }
    },
  };
}
