import { useState, useEffect } from 'react';
import { fetchHealthCard, type HealthCardResponse } from '../api/liff-api';

interface PetHealthPageProps {
  idToken: string;
  displayName: string;
  pictureUrl: string | null;
}

type LoadState = 'loading' | 'error' | 'done';

export function PetHealthPage({ idToken, displayName, pictureUrl }: PetHealthPageProps) {
  const [loadState, setLoadState] = useState<LoadState>('loading');
  const [data, setData] = useState<HealthCardResponse | null>(null);

  const clinicId = new URLSearchParams(window.location.search).get('clinic_id') ?? '';

  // 同期目的のため useEffect 内 setState は許容。
  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    if (!clinicId) {
      setLoadState('error');
      return;
    }

    fetchHealthCard(idToken, clinicId)
      .then((res) => {
        setData(res);
        setLoadState('done');
      })
      .catch((err: unknown) => {
        console.error('[PetHealthPage] fetchHealthCard failed:', err);
        setLoadState('error');
      });
  }, [idToken, clinicId]);
  /* eslint-enable react-hooks/set-state-in-effect */

  if (loadState === 'loading') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-green-50">
        <div className="text-center">
          <div className="w-10 h-10 border-4 border-green-500 border-t-transparent rounded-full animate-spin mx-auto mb-4" aria-hidden="true" />
          <p className="text-gray-500 text-sm">読み込み中...</p>
        </div>
      </div>
    );
  }

  if (loadState === 'error' || !data) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-green-50">
        <div className="max-w-md mx-auto px-4 text-center">
          <div className="text-6xl mb-4" aria-hidden="true">⚠️</div>
          <h1 className="text-xl font-bold text-gray-800 mb-2">データ取得に失敗しました</h1>
          <button
            type="button"
            onClick={() => window.location.reload()}
            className="py-3 px-6 bg-green-500 text-white rounded-xl font-semibold hover:bg-green-600"
          >
            再読み込み
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-green-50 pb-10">
      {/* ヘッダー */}
      <div className="bg-white shadow-sm px-4 py-4 flex items-center gap-3">
        {pictureUrl ? (
          <img
            src={pictureUrl}
            alt={displayName}
            className="w-10 h-10 rounded-full object-cover"
          />
        ) : null}
        <div>
          <p className="text-xs text-gray-400">飼い主</p>
          <p className="text-base font-semibold text-gray-800">{data.owner_name || displayName}</p>
        </div>
      </div>

      <div className="max-w-lg mx-auto px-4 mt-6 space-y-6">
        {data.pets.length === 0 ? (
          <p className="text-center text-gray-500 mt-10">ペット情報はありません</p>
        ) : null}

        {data.pets.map((pet) => (
          <div key={pet.pet_id} className="bg-white rounded-2xl shadow-sm overflow-hidden">
            {/* ペット基本情報 */}
            <div className="bg-green-500 px-5 py-4">
              <h2 className="text-lg font-bold text-white">{pet.pet_name}</h2>
              <p className="text-green-100 text-sm">{pet.species} / {pet.breed}</p>
            </div>

            <div className="px-5 py-4 space-y-3">
              {/* 最終来院日 */}
              <div className="flex justify-between text-sm">
                <span className="text-gray-500">最終来院日</span>
                <span className="font-medium text-gray-800">
                  {pet.last_visit_date ? pet.last_visit_date : '記録なし'}
                </span>
              </div>

              {/* 次回来院推奨日 */}
              <div className="flex justify-between text-sm">
                <span className="text-gray-500">次回来院推奨日</span>
                <span className="font-medium text-green-600">
                  {pet.next_recommended_visit_date ? pet.next_recommended_visit_date : 'なし'}
                </span>
              </div>

              {/* ワクチン記録 */}
              {pet.vaccines.length > 0 ? (
                <div className="mt-3">
                  <p className="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-2">ワクチン記録</p>
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b border-gray-100">
                          <th className="text-left py-2 text-gray-500 font-medium">ワクチン名</th>
                          <th className="text-left py-2 text-gray-500 font-medium">接種日</th>
                          <th className="text-left py-2 text-gray-500 font-medium">次回予定日</th>
                        </tr>
                      </thead>
                      <tbody>
                        {pet.vaccines.map((v, i) => (
                          <tr key={i} className="border-b border-gray-50 last:border-0">
                            <td className="py-2 text-gray-800">{v.vaccine_name}</td>
                            <td className="py-2 text-gray-600">{v.vaccinated_at}</td>
                            <td className="py-2 text-green-600">{v.next_due_at ? v.next_due_at : '—'}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              ) : null}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
