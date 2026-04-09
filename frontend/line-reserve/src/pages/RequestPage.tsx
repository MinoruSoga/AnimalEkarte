import { useState, useCallback } from 'react';
import { ProgressDots } from '../components/ProgressDots';
import { PrimaryButton } from '../components/PrimaryButton';
import { BackButton } from '../components/BackButton';

interface RequestPageProps {
  requestExample: string;
  initialText: string;
  onNext: (text: string) => void;
  onBack: () => void;
}

export function RequestPage({
  requestExample,
  initialText,
  onNext,
  onBack,
}: RequestPageProps) {
  const [text, setText] = useState<string>(initialText);

  const handleChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setText(e.target.value);
  }, []);

  const handleNext = useCallback(() => {
    onNext(text);
  }, [onNext, text]);

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1">
        <ProgressDots current={6} total={8} />

        <div className="px-4">
          <BackButton onClick={onBack} />
          <h2 className="text-lg font-bold text-gray-800 mb-2">ご要望・メモ</h2>
          <p className="text-sm text-gray-500 mb-4">任意：ご要望があればご記入ください</p>
        </div>

        <div className="flex-1 px-4">
          <label htmlFor="request-text" className="block text-sm font-medium text-gray-700 mb-1">
            ご要望
          </label>
          <textarea
            id="request-text"
            value={text}
            onChange={handleChange}
            placeholder={requestExample || 'ご要望・メモをご入力ください'}
            rows={6}
            className="w-full border border-gray-300 rounded-lg px-3 py-2 text-gray-800 focus:outline-none focus:ring-2 focus:ring-line-green focus:border-transparent resize-none"
          />
        </div>

        <div className="px-4 py-6">
          <PrimaryButton onClick={handleNext}>次へ</PrimaryButton>
        </div>
      </div>
    </div>
  );
}
