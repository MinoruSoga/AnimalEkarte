import { useState, useEffect } from 'react';
import liff from '@line/liff';
import { LIFF_MOCK } from '../lib/liff-config';

interface UseLiffReturn {
  idToken: string | null;
  displayName: string;
  isReady: boolean;
}

export function useLiff(liffId: string): UseLiffReturn {
  const [idToken, setIdToken] = useState<string | null>(null);
  const [displayName, setDisplayName] = useState<string>('');
  const [isReady, setIsReady] = useState<boolean>(false);

  useEffect(() => {
    if (!liffId) return;

    if (LIFF_MOCK) {
      setIdToken('mock-token');
      setDisplayName('テストユーザー');
      setIsReady(true);
      return;
    }

    liff.init({ liffId }).then(async () => {
      if (!liff.isLoggedIn()) {
        liff.login();
        return;
      }
      const token = liff.getIDToken();
      setIdToken(token);

      try {
        const profile = await liff.getProfile();
        setDisplayName(profile.displayName);
      } catch {
        // プロフィール取得失敗は無視
      }

      setIsReady(true);
    });
  }, [liffId]);

  return { idToken, displayName, isReady };
}
