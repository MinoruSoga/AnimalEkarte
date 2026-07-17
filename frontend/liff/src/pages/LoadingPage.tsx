import { Spinner } from '@/shared-liff/Spinner';

export function LoadingPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-liff-brand-bg">
      <div className="text-center">
        <Spinner />
        <p className="text-gray-500 text-sm">読み込み中...</p>
      </div>
    </div>
  );
}
