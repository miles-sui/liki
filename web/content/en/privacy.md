# Privacy Policy

This policy describes how 25 Types collects, uses, stores, and protects your personal data. We take a minimal-data approach: we collect only what is necessary to provide the service, and we do not sell or monetize your data.

## Data We Collect

**Account data**: username (required), email (optional, for password recovery and receipts), and a password hash (your password is never stored in plaintext).

**Assessment data**: your responses to 30 forced-choice questions across 6 rounds, the computed deviation vector d, and your identity label (one of 25 types).

**Peer review data**: when someone reviews you, their responses are aggregated into a peer profile. Individual reviewer answers are never exposed to you or to other reviewers.

**Session data**: a JWT (JSON Web Token) stored in a browser cookie for authenticated sessions. No other client-side data is stored.

**Payment data**: subscription status and passport expiry timestamp. We do not store credit card numbers or payment instrument details — these are handled entirely by our payment provider.

## How We Use Your Data

- **Personality computation**: mapping your responses to a deviation vector and identity label.
- **Bond analysis**: computing interaction dynamics between two consenting profiles.
- **Flow projection**: projecting your effective shape across the twelve solar months.
- **Peer aggregation**: pooling peer responses into an aggregate profile for comparison with your self-assessment.
- **Account management**: authentication, password recovery, and email verification.
- **Aggregate statistics**: anonymized data may be used for model validation and improvement.

We do not use your data for automated profiling with legal effect, for sale to third parties, or for advertising.

## Data Storage

Your data is stored in a SQLite database running on our own server. We do not use shared cloud database services for personal data. The database operates in WAL (Write-Ahead Logging) mode for reliability.

## Cookies and Sessions

We set a single essential cookie: a JWT for authentication. This cookie is required for logged-in functionality (saving results, managing peer reviews, accessing Bond and Flow). No tracking cookies, no analytics cookies, no advertising cookies. Full details in our Cookie Policy.

## Third-Party Services

**Email delivery**: transactional emails (verification, password reset, welcome) are sent through Resend. Your email address and the template parameters are transmitted for sending only; no behavioral data is shared.

**Payment processing**: payments are handled by a third-party provider. We receive only your subscription status and expiry timestamp — never your payment instrument details. The provider may set its own session cookies on its domain during checkout.

We do not embed third-party trackers, social media widgets, or external analytics scripts.

## Your Rights

We extend the following rights to all users regardless of location:

- **Access**: export your assessment data as structured JSON at any time, free of charge.
- **Rectification**: update your username and email in your account settings.
- **Deletion**: deactivate your account. A 7-day grace period allows recovery; after that, your data undergoes full anonymization (username replaced, email and password hash cleared, assessments detached).
- **Portability**: your assessment records are available in structured JSON.
- **Objection**: you may object to processing by deleting your account.

If your jurisdiction provides additional rights (for example, under GDPR in the EU/EEA, UK GDPR, PIPL in China, CCPA/CPRA in California, PIPEDA in Canada, PDPA in Singapore, or the Australian Privacy Act), those rights apply to you as well. For requests specific to your jurisdiction, contact us.

## Data Retention

- **Active accounts**: data is retained while your account is active.
- **Deactivated accounts**: 7-day grace period with full data intact, followed by anonymization. Deactivated accounts older than 7 days are anonymized: username becomes a generic label, email and password hash are cleared, and assessments are detached from your identity.
- **Anonymous peer assessments**: assessments with no associated user are deleted after 90 days.
- **Orphaned assessments**: anonymized assessments from deleted accounts are cleaned up after 2 years.
- **Payment records**: retained as required for financial compliance.

## Security Measures

- **Password hashing**: argon2id with memory 47104 KiB, 1 iteration, 4 parallelism — using Go's `golang.org/x/crypto` implementation.
- **Authentication**: JWT (HMAC-SHA256, 30-day validity) with server-side `token_version` — changing your password or logging out instantly invalidates all sessions across all devices.
- **Token generation**: `crypto/rand` (Go standard library CSPRNG) for all random tokens.
- **Transport**: HTTPS enforced via Caddy reverse proxy.
- **Rate limiting**: authentication endpoints are rate-limited at the proxy layer.
- **Database**: SQLite in single-writer WAL mode, preventing write contention and race conditions.
- **Code**: no plaintext secrets in the codebase. All credentials are supplied via environment variables.

## Age Restrictions

25 Types is not directed at children under 13, or under the relevant digital age of consent in your jurisdiction (e.g., 16 in parts of the EU). We do not knowingly collect data from children. If you believe a child has provided us with personal data, contact us for immediate removal.

## International Data Transfers

Our server may be located in a jurisdiction different from yours. By using the service, you acknowledge that your data may be transferred to and processed in the server's location. We apply the same privacy protections regardless of where data is processed, and we rely on applicable legal mechanisms (such as standard contractual clauses where required) for cross-border transfers.

## Changes to This Policy

Material changes will be announced via a notice on the website. Continued use after changes constitutes acceptance. The date of the last update is shown below.

## Contact

For privacy-related inquiries, data access requests, or deletion requests, please contact us. We aim to respond within 30 days, as required by GDPR and similar regulations.

*Last updated: 2026-05-09*
