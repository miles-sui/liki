/**
 * API helpers for E2E test data setup.
 * Calls the real API (via Caddy) so tests don't have to click through setup flows.
 */
import type { Page } from '@playwright/test';
import { DatabaseSync } from 'node:sqlite';

import { resolve } from 'node:path';

const BASE = process.env.BASE_URL || 'http://localhost:8080';
const DB_PATH = process.env.TEST_DB_PATH || resolve(__dirname, '../../../data/25types.db');

export async function req(
  method: string,
  path: string,
  body?: Record<string, unknown>,
  token?: string,
): Promise<{ status: number; data: Record<string, any> }> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  const res = await fetch(`${BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });
  const json = await res.json();
  if (json.error) throw new Error(`${json.error.code}: ${json.error.message}`);
  return { status: res.status, data: json.data as Record<string, any> };
}

/** Inject token into localStorage and navigate to a page as logged-in user. */
export async function loginViaToken(page: Page, token: string, destination: string): Promise<void> {
  await page.goto('/en/login');
  await page.evaluate((t) => { localStorage.setItem('token', t); }, token);
  await page.goto(destination);
}

/**
 * Accept the next browser confirm() dialog.
 * Must be called BEFORE the action that triggers the dialog.
 * Playwright auto-dismisses confirm() returning false when no handler is set up.
 */
export function acceptDialog(page: Page): void {
  page.once('dialog', (dialog) => dialog.accept());
}

export async function register(
  name: string,
  password: string,
  anonToken?: string,
  email?: string,
): Promise<{ token: string; userId: number }> {
  const payload: Record<string, unknown> = { name, password };
  payload.email = email || `suiqiang+e2e-${name}@foxmail.com`;
  if (anonToken) payload.anonymous_token = anonToken;
  const { data } = await req('POST', '/api/auth/register', payload);
  return { token: data.token, userId: data.user.id };
}

export async function login(name: string, password: string): Promise<{ token: string; userId: number }> {
  const { data } = await req('POST', '/api/auth/login', { name, password });
  return { token: data.token, userId: data.user.id };
}

export const FULL_ANSWERS = [
  { qid: 'Q01', selections: ['W', 'F'] }, { qid: 'Q02', selections: ['W', 'M'] },
  { qid: 'Q03', selections: ['F', 'E'] }, { qid: 'Q04', selections: ['W', 'F'] },
  { qid: 'Q05', selections: ['E', 'M'] }, { qid: 'Q06', selections: ['W', 'F'] },
  { qid: 'Q07', selections: ['W', 'E'] }, { qid: 'Q08', selections: ['F', 'E'] },
  { qid: 'Q09', selections: ['W', 'E'] }, { qid: 'Q10', selections: ['F', 'M'] },
  { qid: 'Q11', selections: ['W', 'E'] }, { qid: 'Q12', selections: ['F', 'E'] },
  { qid: 'Q13', selections: ['W', 'R'] }, { qid: 'Q14', selections: ['W', 'F'] },
  { qid: 'Q15', selections: ['E', 'M'] }, { qid: 'Q16', selections: ['W', 'M'] },
  { qid: 'Q17', selections: ['F', 'E'] }, { qid: 'Q18', selections: ['W', 'F'] },
  { qid: 'Q19', selections: ['F', 'R'] }, { qid: 'Q20', selections: ['W', 'E'] },
  { qid: 'Q21', selections: ['W', 'E'] }, { qid: 'Q22', selections: ['F', 'R'] },
  { qid: 'Q23', selections: ['W', 'F'] }, { qid: 'Q24', selections: ['E', 'M'] },
  { qid: 'Q25', selections: ['W', 'R'] }, { qid: 'Q26', selections: ['W', 'F'] },
  { qid: 'Q27', selections: ['W', 'E'] }, { qid: 'Q28', selections: ['W', 'R'] },
  { qid: 'Q29', selections: ['F', 'E'] }, { qid: 'Q30', selections: ['F', 'M'] },
];

export async function submitAssessment(token?: string, anonToken?: string): Promise<{ complete: boolean }> {
  const body: Record<string, unknown> = { answers: FULL_ANSWERS, anonymous_token: anonToken || 'e2e-anon' };
  const { data } = await req('POST', '/api/assessments', body, token);
  return { complete: data.complete };
}

export async function setPublic(token: string): Promise<void> {
  await req('PATCH', '/api/users/me', { is_public: true }, token);
}

export async function createReviewLink(token: string): Promise<{ id: number; token: string }> {
  const { data } = await req('POST', '/api/reviews', {}, token);
  return { id: data.id, token: data.token };
}

export async function renewReviewLink(token: string, id: number): Promise<{ token: string }> {
  const { data } = await req('POST', `/api/reviews/${id}/renew`, {}, token);
  return { token: data.token };
}

export async function createMatchLink(token: string): Promise<{ id: number; token: string }> {
  const { data } = await req('POST', '/api/match-links', {}, token);
  return { id: data.id, token: data.token };
}

export async function createBond(token: string, withUserId: number): Promise<Record<string, any>> {
  const { data } = await req('POST', '/api/bond', { with_user_id: withUserId }, token);
  return data;
}

/** Submit answers via match link (anonymous or authenticated). */
export async function submitMatchLink(
  token: string,
  answers: Array<{ qid: string; selections: string[] }>,
  opts?: { reviewerName?: string; userToken?: string; anonymousToken?: string },
): Promise<Record<string, any>> {
  const body: Record<string, unknown> = { answers };
  if (opts?.reviewerName) body.reviewer_name = opts.reviewerName;
  if (opts?.anonymousToken) body.anonymous_token = opts.anonymousToken;
  const { data } = await req('POST', '/api/m/' + token, body, opts?.userToken);
  return data;
}

/** Retrieve the email verification token from the SQLite database for a given user ID. */
export function getEmailVerToken(userId: number): string | null {
  let db: DatabaseSync;
  try {
    db = new DatabaseSync(DB_PATH, { open: true, readOnly: true });
  } catch {
    return null;
  }
  try {
    const row = db
      .prepare(
        `SELECT token FROM user_tokens
         WHERE user_id = ? AND token_type = 'email_verify'
         AND expires_at > datetime('now')
         ORDER BY created_at DESC LIMIT 1`,
      )
      .get(userId) as { token: string } | undefined;
    return row?.token || null;
  } finally {
    db.close();
  }
}

/** Clear a verified email from any user so retries don't hit UNIQUE constraint. */
export function clearVerifiedEmail(email: string): void {
  let db: DatabaseSync;
  try {
    db = new DatabaseSync(DB_PATH, { open: true });
  } catch {
    return;
  }
  try {
    // email has NOT NULL default '' — clear to empty string instead of NULL.
    db.prepare(`UPDATE users SET email = '', email_verified_at = NULL, pending_email = '' WHERE email = ? OR pending_email = ?`)
      .run(email, email);
  } finally {
    db.close();
  }
}


