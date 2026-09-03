import { useLiffLink } from '../hooks/use-liff-link';
import { Spinner } from '@/shared-liff/Spinner';

export function LiffLinkPage() {
  // SD-14: token/clinic_id は useLiffLink 内部で isReady（liff.init() 完了）後に
  // window.location.search から読む。LINE ログインリダイレクト（liff.state 経由）で
  // 復元される前の URL をここで固定読みしないよう、モジュール直下でのパース処理は行わない。
  const { status, errorMessage } = useLiffLink();

  if (status === 'loading' || status === 'linking') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-liff-brand-bg">
        <div className="text-center">
          <Spinner />
          <p className="text-noah-text-muted text-sm">
            {status === 'linking' ? 'LINEアカウントを連携中...' : '読み込み中...'}
          </p>
        </div>
      </div>
    );
  }

  if (status === 'success') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-liff-brand-bg">
        <div className="max-w-md mx-auto px-4 text-center">
          <div className="text-6xl mb-4" aria-hidden="true">✅</div>
          <h1 className="text-xl font-bold text-noah-text-strong mb-2">連携が完了しました</h1>
          <p className="text-noah-text-muted mb-6">
            LINEアカウントと診察券が連携されました。このページを閉じてください。
          </p>
          <button
            type="button"
            onClick={() => window.close()}
            className="py-3 px-6 bg-liff-brand text-white rounded-xl font-semibold hover:bg-liff-brand-dark"
          >
            閉じる
          </button>
        </div>
      </div>
    );
  }

  if (status === 'conflict') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-liff-brand-bg">
        <div className="max-w-md mx-auto px-4 text-center">
          <div className="text-6xl mb-4" aria-hidden="true">ℹ️</div>
          <h1 className="text-xl font-bold text-noah-text-strong mb-2">連携済みです</h1>
          <p className="text-noah-text-muted mb-6">{errorMessage}</p>
          <button
            type="button"
            onClick={() => window.close()}
            className="py-3 px-6 bg-liff-brand text-white rounded-xl font-semibold hover:bg-liff-brand-dark"
          >
            閉じる
          </button>
        </div>
      </div>
    );
  }

  // expired or error
  const isExpired = status === 'expired';
  return (
    <div className="min-h-screen flex items-center justify-center bg-liff-brand-bg">
      <div className="max-w-md mx-auto px-4 text-center">
        <div className="text-6xl mb-4" aria-hidden="true">⚠️</div>
        <h1 className="text-xl font-bold text-noah-text-strong mb-2">
          {isExpired ? 'リンクが無効です' : 'エラーが発生しました'}
        </h1>
        <p className="text-noah-text-muted mb-6">{errorMessage}</p>
        {!isExpired ? (
          <button
            type="button"
            onClick={() => window.location.reload()}
            className="py-3 px-6 bg-liff-brand text-white rounded-xl font-semibold hover:bg-liff-brand-dark"
          >
            再試行
          </button>
        ) : null}
      </div>
    </div>
  );
}
