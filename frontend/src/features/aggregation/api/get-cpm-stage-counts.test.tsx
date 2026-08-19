import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { server } from '@/testing/mocks/node';
import { createTestWrapper } from '@/testing/utils';
import {
  useGetCPMStageCounts,
  toCPMCountBaseParams,
} from './get-cpm-stage-counts';

const createWrapper = createTestWrapper;

// per_page=1 でも total は母集団全体を返す前提の固定 total。
const SEGMENT_TOTALS: Record<string, number> = {
  cpm_encounter: 5,
  cpm_growing: 12,
  cpm_core: 42,
  cpm_spot: 3,
  cpm_noah: 1,
  cpm_dormant: 20,
  cpm_unclassified: 1,
};

describe('useGetCPMStageCounts (ISSUE-180)', () => {
  beforeEach(() => {
    localStorage.setItem('auth_current_clinic:v1', '1');
    server.use(
      http.get(
        '/api/v1/clinics/:clinic_id/owners/aggregations',
        ({ request }) => {
          const url = new URL(request.url);
          const stage = url.searchParams.get('cpm_stage') ?? '';
          const total = SEGMENT_TOTALS[stage] ?? 0;
          // per_page=1 を想定: owners は最大1件だが total は母集団全体。
          const owners =
            total > 0
              ? [
                  {
                    owner_id: `${stage}-1`,
                    owner_name: 'sample',
                    annual_visit_count: 0,
                    total_visit_count: 0,
                    last_visit_date: null,
                    first_visit_date: null,
                  },
                ]
              : [];
          return HttpResponse.json({ owners, total, page: 1, per_page: 1 });
        }
      )
    );
  });

  afterEach(() => {
    localStorage.removeItem('auth_current_clinic:v1');
  });

  it('reads each segment headcount from total, not the returned page length', async () => {
    const { result } = renderHook(() => useGetCPMStageCounts({ year: 2026 }), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    // 旧失敗モード: owners.length（=1）を数えると人数がすべて 1 になる。
    // total を読むことで母集団全体の人数を反映する。
    expect(result.current.counts.cpm_core).toBe(42);
    expect(result.current.counts.cpm_noah).toBe(1);
    expect(result.current.counts.cpm_encounter).toBe(5);
    expect(result.current.counts.cpm_dormant).toBe(20);
    expect(result.current.counts.cpm_unclassified).toBe(1);
    expect(result.current.total).toBe(84);
  });

  it('returns 0 for an empty segment without breaking', async () => {
    server.use(
      http.get('/api/v1/clinics/:clinic_id/owners/aggregations', () =>
        HttpResponse.json({ owners: [], total: 0, page: 1, per_page: 1 })
      )
    );

    const { result } = renderHook(() => useGetCPMStageCounts({}), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.total).toBe(0);
    expect(result.current.counts.cpm_core).toBe(0);
  });

  it('surfaces isError when a segment request fails', async () => {
    server.use(
      http.get('/api/v1/clinics/:clinic_id/owners/aggregations', () =>
        HttpResponse.json({ error: 'boom' }, { status: 500 })
      )
    );

    const { result } = renderHook(() => useGetCPMStageCounts({}), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
  });
});

describe('toCPMCountBaseParams', () => {
  it('strips pagination/sort/cpm_stage but keeps population filters', () => {
    const result = toCPMCountBaseParams({
      page: 3,
      per_page: 50,
      sort: 'annual_amount',
      order: 'desc',
      cpm_stage: 'cpm_core',
      year: 2026,
      amount_basis: 'gross_total_amount',
      min_amount: 1000,
    });

    expect(result).toEqual({
      year: 2026,
      amount_basis: 'gross_total_amount',
      min_amount: 1000,
    });
  });
});
