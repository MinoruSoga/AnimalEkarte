import { useMemo } from 'react';
import { PrimaryButton } from '../components/PrimaryButton';

interface CompletePageProps {
  reservationId: number;
  notes: string;
  onMyReservations: () => void;
  onNewReservation: () => void;
}

function extractConfirmationNumber(notes: string): string | null {
  const match = notes.match(/R-\d{8}-\d{4}/);
  return match ? match[0] : null;
}

export function CompletePage({
  reservationId,
  notes,
  onMyReservations,
  onNewReservation,
}: CompletePageProps) {
  const confirmationNumber = useMemo(() => extractConfirmationNumber(notes), [notes]);

  return (
    <div className="min-h-screen bg-noah-teal-light flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1 items-center justify-center px-4 py-8">
        <div className="text-center mb-8">
          <div className="text-6xl mb-4" aria-hidden="true">✅</div>
          <h1 className="text-xl font-bold text-noah-teal-dark mb-2">予約が完了しました</h1>
          <p className="text-noah-text-sub text-sm">
            ご予約いただきありがとうございます
          </p>
        </div>

        {confirmationNumber ? (
          <div className="w-full bg-white rounded-xl border border-gray-200 shadow-sm p-4 mb-6 text-center">
            <p className="text-sm text-noah-text-sub mb-1">確認番号</p>
            <p className="text-xl font-bold text-noah-text font-mono">{confirmationNumber}</p>
          </div>
        ) : reservationId > 0 ? (
          <div className="w-full bg-white rounded-xl border border-gray-200 shadow-sm p-4 mb-6 text-center">
            <p className="text-sm text-noah-text-sub mb-1">予約ID</p>
            <p className="text-xl font-bold text-noah-text">{reservationId}</p>
          </div>
        ) : null}

        <div className="w-full space-y-3">
          <PrimaryButton onClick={onMyReservations}>
            マイ予約を確認
          </PrimaryButton>
          <button
            type="button"
            onClick={onNewReservation}
            className="w-full py-3 px-4 border border-gray-300 rounded-xl text-noah-text-sub font-semibold hover:bg-white"
          >
            新しい予約をする
          </button>
        </div>
      </div>
    </div>
  );
}
