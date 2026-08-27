import { getAutoAdvanceHelperText, type FlowAdvanceStep } from '../lib/advance-policy';

interface AutoAdvanceHintProps {
  step: FlowAdvanceStep;
}

/**
 * BUG-030: auto-on-select ステップ専用の進む旨ヒント。
 * explicit-cta ステップでは何も描画しない。
 */
export function AutoAdvanceHint({ step }: AutoAdvanceHintProps) {
  const text = getAutoAdvanceHelperText(step);
  if (!text) {
    return null;
  }

  return (
    <p className="text-sm text-noah-text-sub mb-4" data-testid="auto-advance-hint">
      {text}
    </p>
  );
}
