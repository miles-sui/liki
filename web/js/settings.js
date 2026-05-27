// settings.js — Account settings page
function settingsPage() {
  return {
    nameValue: '',
    newEmail: '',
    currentPassword: '',
    newPassword: '',
    confirmPassword: '',
    saving: false,
    resendCooldown: false,
    showChangeEmail: false,
    showDeleteModal: false,
    showPw: false,
    get newPwStrength() { return passwordStrength(this.newPassword); },

    pwBarClass(n) { return pwBarClass(this.newPwStrength.score, n); },

    activeSection: 'section-account',

    get emailState() {
      if (this.auth.emailVerified) return 'verified';
      if (this.auth.pendingEmail) return 'pending';
      if (this.auth.email) return 'unverified';
      return 'none';
    },

    get auth() { return Alpine.store('auth'); },
    get locale() { return Alpine.store('locale'); },
    get toast() { return Alpine.store('toast'); },

    sections: [
      {
        id: 'section-account',
        icon: '<svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="10" cy="7" r="3.5"/><path d="M3 18c0-3.3 3.1-6 7-6s7 2.7 7 6"/></svg>',
      },
      {
        id: 'section-security',
        icon: '<svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="4" y="9" width="12" height="9" rx="1.5"/><path d="M7 9V6a3 3 0 016 0v3"/></svg>',
      },
      {
        id: 'section-privacy',
        icon: '<svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 10s3-6 8-6 8 6 8 6-3 6-8 6-8-6-8-6z"/><circle cx="10" cy="10" r="2.5"/></svg>',
      },
      {
        id: 'section-data',
        icon: '<svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M10 3v11M6 10l4 4 4-4M3 17h14"/></svg>',
      },
      {
        id: 'section-danger',
        icon: '<svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M10 2L2 18h16L10 2zM10 8v4M10 14.5v1"/></svg>',
      },
    ],

    init() {
      var t = this.locale.t;
      this.sections[0].label = t('settings.account');
      this.sections[1].label = t('settings.security');
      this.sections[2].label = t('settings.privacy');
      this.sections[3].label = t('settings.data');
      this.sections[4].label = t('settings.dangerZone');

      this.nameValue = this.auth.name;
      document.title = t('settings.title') + ' — ' + (window.CURRENT_LOCALE === 'zh-CN' ? '真象 25型' : '25 Types');

      var self = this;
      this._beforeUnload = function (e) {
        if (self.nameValue !== self.auth.name || self.newEmail || self.currentPassword || self.newPassword || self.confirmPassword) {
          e.preventDefault();
        }
      };
      window.addEventListener('beforeunload', this._beforeUnload);
      this._observer = new IntersectionObserver(function (entries) {
        var visible = entries.filter(function (e) { return e.isIntersecting; });
        if (visible.length > 0) {
          self.activeSection = visible[0].target.id;
        }
      }, { rootMargin: '-80px 0px -70% 0px', threshold: 0 });

      for (var i = 0; i < this.sections.length; i++) {
        var el = document.getElementById(this.sections[i].id);
        if (el) self._observer.observe(el);
      }
    },

    destroy() {
      clearTimeout(this._cooldownTimer);
      if (this._observer) this._observer.disconnect();
      if (this._beforeUnload) window.removeEventListener('beforeunload', this._beforeUnload);
    },

    scrollToSection(id) {
      var el = document.getElementById(id);
      if (el) {
        el.scrollIntoView({ behavior: 'smooth', block: 'start' });
      }
      this.activeSection = id;
    },

    // --- Account actions ---

    async saveName() {
      if (this.saving) return;
      var name = this.nameValue.trim();
      if (!name || name === this.auth.name) return;
      this.saving = true;
      try {
        await api('/api/users/me', {
          method: 'PATCH',
          body: JSON.stringify({ name: name }),
        });
        this.auth.name = name;
        this.toast.success(this.locale.t('toast.settingsUpdated'));
      } catch (e) {
        var msg = (e && e.code === 'conflict')
          ? this.locale.t('settings.nameTaken')
          : this.locale.t('error.failed');
        this.toast.error(msg);
      } finally {
        this.saving = false;
      }
    },

    async changeEmail() {
      if (this.saving) return;
      var email = this.newEmail.trim();
      if (!email) return;
      this.saving = true;
      try {
        await api('/api/users/me', {
          method: 'PATCH',
          body: JSON.stringify({ email: email }),
        });
        this.auth.pendingEmail = email;
        this.newEmail = '';
        this.showChangeEmail = false;
        this.toast.success(this.locale.t('toast.emailVerificationSent').replace('{email}', email));
      } catch (e) {
        var msg = (e && e.code === 'conflict')
          ? this.locale.t('settings.emailTaken')
          : this.locale.t('error.failed');
        this.toast.error(msg);
      } finally {
        this.saving = false;
      }
    },

    async resendVerification() {
      if (this.saving || this.resendCooldown) return;
      this.saving = true;
      try {
        var res = await api('/api/auth/resend-verification', { method: 'POST' });
        this.toast.success(this.locale.t('toast.emailResent').replace('{email}', res.data.email));
        this.resendCooldown = true;
        clearTimeout(this._cooldownTimer);
        var self = this;
        this._cooldownTimer = setTimeout(function () { self.resendCooldown = false; }, 3000);
      } catch (e) {
        if (e && e.code === 'already_verified') {
          this.auth.fetchMe();
          this.toast.success(this.locale.t('settings.emailAlreadyVerified'));
        } else {
          this.toast.error(this.locale.t('error.failed'));
        }
      } finally {
        this.saving = false;
      }
    },

    async changePassword() {
      if (this.saving) return;
      if (this.newPassword.length < 8) {
        this.toast.error(this.locale.t('register.passwordHint'));
        return;
      }
      if (this.newPassword !== this.confirmPassword) {
        this.toast.error(this.locale.t('settings.passwordMismatch'));
        return;
      }
      this.saving = true;
      try {
        var res = await api('/api/auth/password', {
          method: 'PUT',
          body: JSON.stringify({
            current_password: this.currentPassword,
            new_password: this.newPassword,
          }),
        });
        if (res.data && res.data.token) {
          this.auth.token = res.data.token;
          localStorage.setItem('token', res.data.token);
        }
        this.currentPassword = '';
        this.newPassword = '';
        this.confirmPassword = '';
        this.toast.success(this.locale.t('settings.passwordChanged'));
      } catch (e) {
        var msg = (e && e.code === 'incorrect_password')
          ? this.locale.t('settings.incorrectPassword')
          : this.locale.t('error.failed');
        this.toast.error(msg);
      } finally {
        this.saving = false;
      }
    },

    async togglePrivacy() {
      if (this.saving) return;
      this.saving = true;
      try {
        var next = !this.auth.isPublic;
        await api('/api/users/me', {
          method: 'PATCH',
          body: JSON.stringify({ is_public: next }),
        });
        this.auth.isPublic = next;
        this.toast.success(next ? this.locale.t('profile.madePublic') : this.locale.t('profile.madePrivate'));
      } catch (e) {
        this.toast.error(this.locale.t('error.failed'));
      } finally {
        this.saving = false;
      }
    },

    async exportData() {
      if (this.saving) return;
      this.saving = true;
      try {
        var res = await api('/api/users/me/export', { quiet: true });
        var blob = new Blob([JSON.stringify(res.data || res, null, 2)], { type: 'application/json' });
        saveOrShare(blob, '25types-export-' + new Date().toISOString().slice(0, 10) + '.json');
        this.toast.success(this.locale.t('toast.dataExported'));
      } catch (e) {
        this.toast.error(this.locale.t('error.failed'));
      } finally {
        this.saving = false;
      }
    },

    async confirmDelete() {
      if (this.saving) return;
      this.saving = true;
      try {
        await api('/api/users/me', { method: 'DELETE', quiet: true });
        this.showDeleteModal = false;
        this.toast.success(this.locale.t('toast.accountDeactivated'));
        this.auth.logout();
      } catch (e) {
        this.toast.error(this.locale.t('error.failed'));
      } finally {
        this.saving = false;
      }
    },
  };
}
