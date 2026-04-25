import { API_BASE_URL } from '../lib/liff-config';

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
