export function LoadingPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-liff-brand-bg">
      <div className="text-center">
        <div
          className="w-10 h-10 border-4 border-liff-brand border-t-transparent rounded-full animate-spin mx-auto mb-4"
          aria-hidden="true"
        />
        <p className="text-gray-500 text-sm">読み込み中...</p>
      </div>
    </div>
  );
}
