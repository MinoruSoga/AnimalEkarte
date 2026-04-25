import { API_BASE_URL } from '../lib/liff-config';

export class LiffApiError extends Error {
  constructor(public readonly status: number, message: string) {
    super(message);
    this.name = 'LiffApiError';
  }
}

export interface PetVaccineRecord {
  vaccine_name: string;
  vaccinated_at: string;
  next_due_at: string | null;
}

export interface PetHealthCard {
  pet_id: string;
  pet_name: string;
  species: string;
  breed: string;
  next_recommended_visit_date: string | null;
  vaccines: PetVaccineRecord[];
  last_visit_date: string | null;
}

export interface HealthCardResponse {
  owner_name: string;
  pets: PetHealthCard[];
}

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
  const res = await fetch(`${API_BASE_URL}/v1/liff/health-card`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${idToken}`,
      'X-Clinic-ID': clinicId,
    },
  });

  if (!res.ok) {
    const text = await res.text().catch(() => '');
    console.error('[fetchHealthCard] error:', res.status, text);
    throw new Error(`HealthCard fetch failed: ${res.status}`);
  }

  return res.json() as Promise<HealthCardResponse>;
}
