import { z } from 'zod';
import { API_BASE_URL } from '../lib/liff-config';
import { devError } from '@/shared-liff/dev-log';

export class LiffApiError extends Error {
  constructor(public readonly status: number, message: string) {
    super(message);
    this.name = 'LiffApiError';
  }
}

const petVaccineRecordSchema = z.object({
  vaccine_name: z.string(),
  vaccinated_at: z.string(),
  next_due_at: z.string().nullable(),
});

const petHealthCardSchema = z.object({
  pet_id: z.string(),
  pet_name: z.string(),
  species: z.string(),
  breed: z.string(),
  vaccines: z.array(petVaccineRecordSchema),
  last_visit_date: z.string().nullable(),
});

const healthCardResponseSchema = z.object({
  owner_name: z.string(),
  pets: z.array(petHealthCardSchema),
});

export type HealthCardResponse = z.infer<typeof healthCardResponseSchema>;

/** Brand-only slice of public LIFF settings (auth not required). */
const brandSettingsSchema = z.object({
  header_text: z.string().optional(),
});

export type BrandSettings = z.infer<typeof brandSettingsSchema>;

export async function linkLineAccount(
  clinicId: string,
  linkToken: string,
  lineIdToken: string,
): Promise<void> {
  const res = await fetch(`${API_BASE_URL}/api/liff/${encodeURIComponent(clinicId)}/link`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ link_token: linkToken, line_id_token: lineIdToken }),
  });

  if (!res.ok) {
    throw new LiffApiError(res.status, `Link failed: ${res.status}`);
  }
}

export async function fetchHealthCard(idToken: string, clinicId: string): Promise<HealthCardResponse> {
  const res = await fetch(`${API_BASE_URL}/api/liff/${encodeURIComponent(clinicId)}/health-card`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${idToken}`,
    },
  });

  if (!res.ok) {
    const text = await res.text().catch(() => '');
    devError('[fetchHealthCard] error:', res.status, text);
    // R-F22: status を保持した LiffApiError に統一し、呼び出し側でステータスコード別の
    // エラーメッセージ・再試行可否を判定できるようにする（linkLineAccount と同じ規約）。
    throw new LiffApiError(res.status, `HealthCard fetch failed: ${res.status}`);
  }

  const json: unknown = await res.json();
  const parsed = healthCardResponseSchema.safeParse(json);
  if (!parsed.success) {
    devError('[fetchHealthCard] invalid response shape:', parsed.error);
    throw new Error('HealthCard response validation failed');
  }

  return parsed.data;
}

/**
 * Public clinic brand settings for header chrome (BUG-026).
 * Fail closed: non-OK or invalid shape throws; callers leave header text empty.
 */
export async function fetchBrandSettings(clinicId: string): Promise<BrandSettings> {
  const res = await fetch(`${API_BASE_URL}/api/liff/${encodeURIComponent(clinicId)}/settings`, {
    method: 'GET',
  });

  if (!res.ok) {
    throw new LiffApiError(res.status, `Brand settings fetch failed: ${res.status}`);
  }

  const json: unknown = await res.json();
  const parsed = brandSettingsSchema.safeParse(json);
  if (!parsed.success) {
    devError('[fetchBrandSettings] invalid response shape:', parsed.error);
    throw new Error('Brand settings response validation failed');
  }

  return parsed.data;
}
