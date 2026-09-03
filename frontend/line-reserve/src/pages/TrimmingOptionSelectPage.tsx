import { useCallback, useState } from 'react';
import { liffApi } from '../api/liff-api';
import { ProgressDots } from '../components/ProgressDots';
import { BackButton } from '../components/BackButton';
import { useFetchState } from '@/shared-liff/use-fetch-state';
import { formatCurrency } from '@/lib/format/number';
import { getStepProgress } from '../lib/step-progress';
import { EXPLICIT_PRIMARY_CTA_LABEL } from '../lib/advance-policy';

interface TrimmingOptionSelectPageProps {
  clinicId: string;
  idToken: string;
  selectedOptionIds: number[];
  onNext: (optionIds: number[]) => void;
  onBack: () => void;
}

function formatPrice(price: number | null): string {
  if (price === null) return '';
  return `+${formatCurrency(price)}`;
}

export function TrimmingOptionSelectPage({
  clinicId,
  idToken,
  selectedOptionIds,
  onNext,
  onBack,
}: TrimmingOptionSelectPageProps) {
  const [selected, setSelected] = useState<Set<number>>(new Set(selectedOptionIds));
  const fetcher = useCallback(() => liffApi.getTrimmingOptions(clinicId, idToken), [clinicId, idToken]);
  // R-F22/R-F23: ステータス別メッセージ解決と再試行導線を共通フックに統合。
  const { data: options, loading, error, retry } = useFetchState(fetcher, 'オプションの取得');
  const { current, total } = getStepProgress('trimmingOptionSelect', true);

  const toggleOption = (id: number) => {
    setSelected(prev => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  return (
    <div className="min-h-screen bg-noah-teal-light flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1">
        <ProgressDots current={current} total={total} />

        <div className="px-4">
          <BackButton onClick={onBack} />
          <h2 className="text-lg font-bold text-noah-teal-dark mb-1">オプションを選択</h2>
          <p className="text-sm text-noah-text-sub mb-4">複数選択できます（任意）</p>
        </div>

        <div className="flex-1">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="text-noah-text-sub">読み込み中...</div>
            </div>
          ) : error ? (
            <div className="px-4 py-8 text-center text-noah-danger">
              <p role="alert">{error.message}</p>
              {error.canRetry ? (
                <button
                  type="button"
                  onClick={retry}
                  className="mt-3 text-sm font-medium text-noah-teal-dark underline"
                >
                  再試行
                </button>
              ) : null}
            </div>
          ) : (
            <div className="bg-white border-t border-noah-border">
              {(options ?? []).map(option => (
                <button
                  key={option.id}
                  type="button"
                  className="w-full flex items-center px-4 py-4 border-b border-noah-border-light text-left active:bg-noah-surface-muted"
                  onClick={() => toggleOption(option.id)}
                >
                  <div
                    className={`w-5 h-5 rounded border-2 flex-shrink-0 flex items-center justify-center mr-3 ${
                      selected.has(option.id)
                        ? 'bg-noah-teal border-noah-teal'
                        : 'border-noah-border-input'
                    }`}
                  >
                    {selected.has(option.id) ? (
                      <svg className="w-3 h-3 text-white" fill="none" viewBox="0 0 12 12" aria-hidden="true">
                        <path d="M2 6l3 3 5-5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                      </svg>
                    ) : null}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium text-noah-text-heading">{option.name}</span>
                      {option.price !== null ? (
                        <span className="text-sm text-noah-teal-dark ml-2 flex-shrink-0">
                          {formatPrice(option.price)}
                        </span>
                      ) : null}
                    </div>
                    {option.description ? (
                      <p className="text-xs text-noah-text-sub mt-0.5 truncate">{option.description}</p>
                    ) : null}
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>

        <div className="px-4 py-4 bg-white border-t border-noah-border">
          <button
            type="button"
            className="w-full py-3 bg-noah-teal text-white rounded-lg font-medium text-sm active:opacity-80"
            onClick={() => onNext(Array.from(selected))}
          >
            {EXPLICIT_PRIMARY_CTA_LABEL}
          </button>
        </div>
      </div>
    </div>
  );
}
