import { useState, useEffect } from 'react';
import liff from '@line/liff';
import { LIFF_MOCK } from '../lib/liff-config';

interface UseLiffReturn {
  idToken: string | null;
  displayName: string;
  isReady: boolean;
  initError: boolean;
}

export function useLiff(liffId: string): UseLiffReturn {
  const [idToken, setIdToken] = useState<string | null>(() => LIFF_MOCK ? 'mock-token' : null);
  const [displayName, setDisplayName] = useState<string>(() => LIFF_MOCK ? 'テストユーザー' : '');
  const [isReady, setIsReady] = useState<boolean>(() => LIFF_MOCK);
  const [initError, setInitError] = useState<boolean>(false);

  useEffect(() => {
    if (!liffId) return;
    if (LIFF_MOCK) return;

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
    }).catch(() => {
      setInitError(true);
      setIsReady(true); // ローディングを解除して ErrorPage へ遷移させる
    });
  }, [liffId]);

  return { idToken, displayName, isReady, initError };
}
