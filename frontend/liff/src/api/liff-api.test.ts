import { afterEach, describe, expect, it, vi } from 'vitest';
import { linkLineAccount } from './liff-api';

describe('linkLineAccount', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('204 No Content を本文解析せず連携成功として扱う', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 }));
    vi.stubGlobal('fetch', fetchMock);

    await expect(linkLineAccount('clinic/1', 'link-token', 'line-id-token')).resolves.toBeUndefined();
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining('/api/liff/clinic%2F1/link'),
      expect.objectContaining({
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          link_token: 'link-token',
          line_id_token: 'line-id-token',
        }),
      }),
    );
  });
});
