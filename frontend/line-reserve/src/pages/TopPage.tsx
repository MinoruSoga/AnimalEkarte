import type { LiffSettings } from '../types/models';
import { PrimaryButton } from '../components/PrimaryButton';

interface TopPageProps {
  settings: LiffSettings;
  onNewReservation: () => void;
  onMyReservations: () => void;
}

export function TopPage({ settings, onNewReservation, onMyReservations }: TopPageProps) {
  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      {/* ヘッダー */}
      <header className="bg-line-green text-white py-4 px-4 text-center">
        <h1 className="text-lg font-bold">{settings.header_text}</h1>
      </header>

      {/* メインコンテンツ */}
      <main className="flex-1 max-w-md mx-auto w-full px-4 py-8 flex flex-col gap-4">
        <div className="bg-white rounded-lg border border-gray-200 p-6 text-center">
          <p className="text-gray-600 mb-6 text-sm">
            ご予約はこちらから手続きください
          </p>
          <PrimaryButton onClick={onNewReservation}>
            新規予約
          </PrimaryButton>
        </div>

        <button
          type="button"
          onClick={onMyReservations}
          className="w-full bg-white border border-gray-300 text-gray-700 rounded-lg py-3 px-4 font-semibold hover:bg-gray-50 active:bg-gray-100 transition-colors"
        >
          予約確認・キャンセル
        </button>
      </main>

      {/* フッター */}
      {settings.phone_number ? (
        <footer className="py-4 text-center text-sm text-gray-500 border-t border-gray-200 bg-white">
          <a href={`tel:${settings.phone_number}`} className="text-line-green">
            📞 {settings.phone_number}
          </a>
        </footer>
      ) : null}
    </div>
  );
}
