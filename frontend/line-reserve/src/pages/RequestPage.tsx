import { useState, useCallback } from 'react';
import { ProgressDots } from '../components/ProgressDots';
import { PrimaryButton } from '../components/PrimaryButton';
import { BackButton } from '../components/BackButton';
import { getStepProgress } from '../lib/step-progress';
import { EXPLICIT_PRIMARY_CTA_LABEL } from '../lib/advance-policy';

interface RequestPageProps {
  requestExample: string;
  initialText: string;
  isTrimming: boolean;
  onNext: (text: string) => void;
  onBack: () => void;
}

export function RequestPage({
  requestExample,
  initialText,
  isTrimming,
  onNext,
  onBack,
}: RequestPageProps) {
  const [text, setText] = useState<string>(initialText);
  // SD-16: トリミング分岐で挿入される2ステップ分、以降のフロー全体の total を一貫させる
  const { current, total } = getStepProgress('request', isTrimming);

  const handleChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setText(e.target.value);
  }, []);

  const handleNext = useCallback(() => {
    onNext(text);
  }, [onNext, text]);

  return (
    <div className="min-h-screen bg-noah-teal-light flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1">
        <ProgressDots current={current} total={total} />

        <div className="px-4">
          <BackButton onClick={onBack} />
          <h2 className="text-lg font-bold text-noah-teal-dark mb-2">ご要望・メモ</h2>
          <p className="text-sm text-noah-text-sub mb-4">任意：ご要望があればご記入ください</p>
        </div>

        <div className="flex-1 px-4">
          <label htmlFor="request-text" className="block text-sm font-medium text-noah-text-sub mb-1">
            ご要望
          </label>
          <textarea
            id="request-text"
            value={text}
            onChange={handleChange}
            placeholder={requestExample || 'ご要望・メモをご入力ください'}
            rows={6}
            className="w-full border border-gray-300 rounded-xl px-3 py-2 text-noah-text focus:outline-none focus:ring-2 focus:ring-noah-teal focus:border-transparent resize-none"
          />
        </div>

        <div className="px-4 py-6">
          <PrimaryButton onClick={handleNext}>{EXPLICIT_PRIMARY_CTA_LABEL}</PrimaryButton>
        </div>
      </div>
    </div>
  );
}
