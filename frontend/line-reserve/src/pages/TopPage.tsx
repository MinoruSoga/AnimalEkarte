import type { LiffSettings } from "../types/models";
import { PrimaryButton } from "../components/PrimaryButton";

interface TopPageProps {
  settings: LiffSettings;
  onNewReservation: () => void;
  onMyReservations: () => void;
}

export function TopPage({ settings, onNewReservation, onMyReservations }: TopPageProps) {
  return (
    <div className="min-h-screen bg-noah-teal-light flex flex-col">
      {/* ヘッダー */}
      <header className="bg-noah-teal text-white py-5 px-4 text-center">
        <h1 className="text-lg font-bold tracking-tight">
          {settings.header_text || "ノア動物病院"}
        </h1>
        <p className="text-sm text-white/80 mt-1">オンライン予約</p>
      </header>

      {/* メインコンテンツ */}
      <main className="flex-1 max-w-md mx-auto w-full px-4 py-8 flex flex-col gap-4">
        <div className="bg-white rounded-xl border border-noah-border-light p-6 text-center">
          <p className="text-noah-text-sub mb-6 text-sm">ご予約はこちらから手続きください</p>
          <PrimaryButton onClick={onNewReservation}>新規予約</PrimaryButton>
        </div>

        <button
          type="button"
          onClick={onMyReservations}
          className="w-full bg-white border border-noah-border text-noah-text rounded-xl py-3 px-4 font-semibold hover:bg-noah-surface-muted active:bg-noah-surface-subtle transition-colors"
        >
          予約確認・キャンセル
        </button>
      </main>

      {/* フッター */}
      {settings.phone_number ? (
        <footer className="py-4 text-center text-sm text-noah-text-sub border-t border-noah-border bg-white">
          <a href={`tel:${settings.phone_number}`} className="text-noah-teal font-medium">
            📞 {settings.phone_number}
          </a>
        </footer>
      ) : null}
    </div>
  );
}
