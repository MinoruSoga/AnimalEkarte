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
    <div className="min-h-screen bg-gray-50 flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1 items-center justify-center px-4 py-8">
        <div className="text-center mb-8">
          <div className="text-6xl mb-4" aria-hidden="true">✅</div>
          <h1 className="text-xl font-bold text-gray-800 mb-2">予約が完了しました</h1>
          <p className="text-gray-600 text-sm">
            ご予約いただきありがとうございます
          </p>
        </div>

        {confirmationNumber ? (
          <div className="w-full bg-white rounded-lg border border-gray-200 p-4 mb-6 text-center">
            <p className="text-sm text-gray-500 mb-1">確認番号</p>
            <p className="text-xl font-bold text-gray-800 font-mono">{confirmationNumber}</p>
          </div>
        ) : reservationId > 0 ? (
          <div className="w-full bg-white rounded-lg border border-gray-200 p-4 mb-6 text-center">
            <p className="text-sm text-gray-500 mb-1">予約ID</p>
            <p className="text-xl font-bold text-gray-800">{reservationId}</p>
          </div>
        ) : null}

        <div className="w-full space-y-3">
          <PrimaryButton onClick={onMyReservations}>
            マイ予約を確認
          </PrimaryButton>
          <button
            type="button"
            onClick={onNewReservation}
            className="w-full py-3 px-4 border border-gray-300 rounded-lg text-gray-700 font-semibold hover:bg-gray-50"
          >
            新しい予約をする
          </button>
        </div>
      </div>
    </div>
  );
}
