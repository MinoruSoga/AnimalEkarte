import { useLiffLink } from '../hooks/use-liff-link';

const params = new URLSearchParams(window.location.search);
const LINK_TOKEN = params.get('token') ?? '';
const CLINIC_ID = params.get('clinic_id') ?? '';

function Spinner() {
  return (
    <div
      className="w-10 h-10 border-4 border-liff-brand border-t-transparent rounded-full animate-spin mx-auto mb-4"
      aria-hidden="true"
    />
  );
}

export function LiffLinkPage() {
  const { status, errorMessage } = useLiffLink(CLINIC_ID, LINK_TOKEN);

  if (status === 'loading' || status === 'linking') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-liff-brand-bg">
        <div className="text-center">
          <Spinner />
          <p className="text-gray-500 text-sm">
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
          <h1 className="text-xl font-bold text-gray-800 mb-2">連携が完了しました</h1>
          <p className="text-gray-500 mb-6">
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
          <h1 className="text-xl font-bold text-gray-800 mb-2">連携済みです</h1>
          <p className="text-gray-500 mb-6">{errorMessage}</p>
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
        <h1 className="text-xl font-bold text-gray-800 mb-2">
          {isExpired ? 'リンクが無効です' : 'エラーが発生しました'}
        </h1>
        <p className="text-gray-500 mb-6">{errorMessage}</p>
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
