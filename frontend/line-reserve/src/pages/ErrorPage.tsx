interface ErrorPageProps {
  message?: string;
}

export function ErrorPage({ message = '予期しないエラーが発生しました。' }: ErrorPageProps) {
  return (
    <div className="min-h-screen flex items-center justify-center bg-noah-teal-light">
      <div className="max-w-md mx-auto px-4 text-center">
        <div className="text-6xl mb-4" aria-hidden="true">⚠️</div>
        <h1 className="text-xl font-bold text-noah-teal-dark mb-2">エラーが発生しました</h1>
        <p className="text-noah-text-sub mb-6">{message}</p>
        <button
          type="button"
          onClick={() => window.location.reload()}
          className="py-3 px-6 bg-noah-teal text-white rounded-xl font-semibold hover:bg-noah-teal-dark"
        >
          再読み込み
        </button>
      </div>
    </div>
  );
}
