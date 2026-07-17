export interface ErrorPageTheme {
  bg: string;
  heading: string;
  body: string;
  button: string;
  buttonHover: string;
}

interface ErrorPageProps {
  message?: string;
  theme: ErrorPageTheme;
}

/** liff / line-reserve 共通のエラーページ（FE6-3: 2箇所のコピペを統合。差分は呼び出し側が渡す色クラスのみ） */
export function ErrorPage({ message = '予期しないエラーが発生しました。', theme }: ErrorPageProps) {
  return (
    <div className={`min-h-screen flex items-center justify-center ${theme.bg}`}>
      <div className="max-w-md mx-auto px-4 text-center">
        <div className="text-6xl mb-4" aria-hidden="true">⚠️</div>
        <h1 className={`text-xl font-bold ${theme.heading} mb-2`}>エラーが発生しました</h1>
        <p className={`${theme.body} mb-6`}>{message}</p>
        <button
          type="button"
          onClick={() => window.location.reload()}
          className={`py-3 px-6 ${theme.button} text-white rounded-xl font-semibold ${theme.buttonHover}`}
        >
          再読み込み
        </button>
      </div>
    </div>
  );
}
