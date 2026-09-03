import { Spinner } from '@/shared-liff/Spinner';

export function LoadingPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-liff-brand-bg">
      <div className="text-center">
        <Spinner />
        <p className="text-noah-text-muted text-sm">読み込み中...</p>
      </div>
    </div>
  );
}
