import axios from "axios";

export interface FetchErrorInfo {
  status: number | undefined;
  message: string;
  canRetry: boolean;
}

interface StatusCarryingError {
  status: number;
}

function hasStatus(err: unknown): err is StatusCarryingError {
  return (
    typeof err === "object" &&
    err !== null &&
    "status" in err &&
    typeof (err as { status: unknown }).status === "number"
  );
}

function extractStatus(err: unknown): number | undefined {
  if (axios.isAxiosError(err)) {
    return err.response?.status;
  }
  if (hasStatus(err)) {
    return err.status;
  }
  return undefined;
}

/**
 * liff/line-reserve 共通のステータスコード別エラーメッセージ解決（R-F22）。
 * axios の AxiosError（line-reserve）と status を持つエラー（liff の LiffApiError 等）の
 * 両方をダックタイピングで判別する。クラス定義そのものは各アプリ側に閉じたまま共有できる。
 *
 * - 401: LINEアプリ再起動・再ログインを促す。トークン失効は再試行だけでは解消しないため canRetry=false。
 * - 5xx / ステータス不明（ネットワーク断等）: 一時的な障害として再試行を許可する。
 * - それ以外の4xx: コンテキスト付きの汎用メッセージ。再試行は許可するが解消しない可能性がある。
 */
export function resolveFetchError(err: unknown, context: string): FetchErrorInfo {
  const status = extractStatus(err);

  if (status === 401) {
    return {
      status,
      message: "ログイン情報の有効期限が切れました。LINEアプリを再起動して開き直してください。",
      canRetry: false,
    };
  }

  if (status !== undefined && status >= 500) {
    return {
      status,
      message: "サーバーエラーが発生しました。しばらく経ってから再度お試しください。",
      canRetry: true,
    };
  }

  if (status !== undefined) {
    return {
      status,
      message: `${context}に失敗しました。`,
      canRetry: true,
    };
  }

  return {
    status: undefined,
    message: `${context}に失敗しました。しばらく経ってから再度お試しください。`,
    canRetry: true,
  };
}
