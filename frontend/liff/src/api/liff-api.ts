import { z } from 'zod';
import { API_BASE_URL } from '../lib/liff-config';

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
    console.error('[fetchHealthCard] error:', res.status, text);
    // R-F22: status を保持した LiffApiError に統一し、呼び出し側でステータスコード別の
    // エラーメッセージ・再試行可否を判定できるようにする（linkLineAccount と同じ規約）。
    throw new LiffApiError(res.status, `HealthCard fetch failed: ${res.status}`);
  }

  const json: unknown = await res.json();
  const parsed = healthCardResponseSchema.safeParse(json);
  if (!parsed.success) {
    console.error('[fetchHealthCard] invalid response shape:', parsed.error);
    throw new Error('HealthCard response validation failed');
  }

  return parsed.data;
}
