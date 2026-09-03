import { useCallback, useEffect, useState, type Dispatch, type SetStateAction } from 'react';
import { resolveFetchError, type FetchErrorInfo } from './handle-fetch-error';

export interface UseFetchStateResult<T> {
  data: T | null;
  loading: boolean;
  error: FetchErrorInfo | null;
  retry: () => void;
  setData: Dispatch<SetStateAction<T | null>>;
}

/**
 * liff/line-reserve 共通の GET-on-mount フェッチ状態管理フック（R-F23）。
 * loading/error/data/retry を返し、ステータスコード別メッセージ解決は resolveFetchError（R-F22）に委譲する。
 *
 * fetcher は呼び出し側で useCallback に包んで渡すこと。依存配列を可変長の spread で
 * 受け取る設計にすると react-hooks/exhaustive-deps が静的検証できなくなるため、
 * このフック自身は [fetcher, context] のみを見る固定長の依存配列を持つ。
 * setData は取得後のローカルな楽観的更新（例: 一覧の一部を書き換える）向けに公開する。
 */
export function useFetchState<T>(fetcher: () => Promise<T>, context: string): UseFetchStateResult<T> {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<FetchErrorInfo | null>(null);
  const [retryToken, setRetryToken] = useState(0);
  // FE-RC-071: data/error のどちらも未確定な間だけ loading とする派生値。
  // 初回 mount は両方 null なので自動的に true になり、retry() 側で error を
  // 先に消してから effect を再実行するため、effect 内で同期 setState を行う
  // 必要が無くなり set-state-in-effect の抑制も不要になった。
  const loading = data === null && error === null;

  useEffect(() => {
    let cancelled = false;

    fetcher()
      .then((result) => {
        if (cancelled) return;
        setData(result);
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        setError(resolveFetchError(err, context));
      });

    return () => {
      cancelled = true;
    };
  }, [fetcher, context, retryToken]);

  const retry = useCallback(() => {
    setError(null);
    setRetryToken((t) => t + 1);
  }, []);

  return { data, loading, error, retry, setData };
}
