import { useState, useEffect, useRef } from 'react';
import { useLiff } from '@/shared-liff/use-liff';
import { LIFF_ID, LIFF_MOCK, LINK_SUCCESS_DISPLAY_MS } from '../lib/liff-config';
import { linkLineAccount, LiffApiError } from '../api/liff-api';

type LinkStatus = 'loading' | 'linking' | 'success' | 'conflict' | 'expired' | 'error';

interface UseLiffLinkResult {
  status: LinkStatus;
  errorMessage: string | null;
}

export function useLiffLink(): UseLiffLinkResult {
  const { idToken, isReady, initError } = useLiff(LIFF_ID);
  const [status, setStatus] = useState<LinkStatus>('loading');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const linkPromiseRef = useRef<Promise<void> | null>(null);

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

    // SD-14: token/clinic_id は isReady（liff.init() 完了）後に初めて読む。
    // LINE ログインリダイレクト（未ログイン状態での初回アクセス等）を経由する場合、
    // LIFF SDK は元のクエリを liff.state に包んで戻し、liff.init() の完了までに
    // history.replaceState で元の URL（?token=...&clinic_id=...）へ復元する。
    // isReady より前に window.location.search を読むと、この復元前の
    // liff.state 付き URL を掴んでしまい token/clinic_id が欠落する。
    const params = new URLSearchParams(window.location.search);
    const clinicId = params.get('clinic_id') ?? '';
    const linkToken = params.get('token') ?? '';

    if (!clinicId || !linkToken) {
      setStatus('error');
      setErrorMessage('無効なURLです。QRコードを再度読み取ってください');
      return;
    }

    setStatus('linking');

    if (LIFF_MOCK) {
      const timer = setTimeout(() => setStatus('success'), LINK_SUCCESS_DISPLAY_MS);
      return () => clearTimeout(timer);
    }

    linkPromiseRef.current ??= linkLineAccount(clinicId, linkToken, idToken)
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
  }, [isReady, idToken, initError]);
  /* eslint-enable react-hooks/set-state-in-effect */

  return { status, errorMessage };
}
