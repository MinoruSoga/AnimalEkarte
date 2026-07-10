import { useState, useEffect } from 'react';
import { useLiff } from '@/shared-liff/use-liff';
import { LIFF_ID, LIFF_MOCK } from '../lib/liff-config';
import { linkLineAccount, LiffApiError } from '../api/liff-api';

type LinkStatus = 'loading' | 'linking' | 'success' | 'conflict' | 'expired' | 'error';

interface UseLiffLinkResult {
  status: LinkStatus;
  errorMessage: string | null;
}

export function useLiffLink(clinicId: string, linkToken: string): UseLiffLinkResult {
  const { idToken, isReady, initError } = useLiff(LIFF_ID);
  const [status, setStatus] = useState<LinkStatus>('loading');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // 認証完了 → API 連携 → ステータス更新。useEffect 内 setState は同期目的のため許容。
  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    if (!isReady) return;

    if (initError) {
      setStatus('error');
      setErrorMessage('LINE認証に失敗しました');
      return;
    }

    if (!idToken) {
      setStatus('error');
      setErrorMessage('ログイン情報が取得できませんでした');
      return;
    }

    if (!clinicId || !linkToken) {
      setStatus('error');
      setErrorMessage('無効なURLです。QRコードを再度読み取ってください');
      return;
    }

    setStatus('linking');

    if (LIFF_MOCK) {
      const timer = setTimeout(() => setStatus('success'), 800);
      return () => clearTimeout(timer);
    }

    linkLineAccount(clinicId, linkToken, idToken)
      .then(() => setStatus('success'))
      .catch((err: unknown) => {
        if (err instanceof LiffApiError) {
          switch (err.status) {
            case 409:
              setStatus('conflict');
              setErrorMessage('このLINEアカウントはすでに連携済みです');
              break;
            case 400:
              setStatus('expired');
              setErrorMessage('リンクトークンが無効または期限切れです。スタッフにお声がけください');
              break;
            case 401:
              setStatus('error');
              setErrorMessage('LINE認証に失敗しました。もう一度お試しください');
              break;
            default:
              setStatus('error');
              setErrorMessage('連携中にエラーが発生しました。しばらくしてからお試しください');
          }
        } else {
          setStatus('error');
          setErrorMessage('連携中にエラーが発生しました。しばらくしてからお試しください');
        }
      });
  }, [isReady, idToken, initError, clinicId, linkToken]);
  /* eslint-enable react-hooks/set-state-in-effect */

  return { status, errorMessage };
}
