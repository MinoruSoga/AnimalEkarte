interface ProgressDotsProps {
  current: number; // 1-based
  total: number;
}

export function ProgressDots({ current, total }: ProgressDotsProps) {
  return (
    <div className="flex items-center justify-center gap-2 py-3">
      {Array.from({ length: total }, (_, i) => (
        <div
          key={i}
          className={`rounded-full transition-all ${
            i + 1 === current
              ? 'w-3 h-3 bg-noah-teal'
              : i + 1 < current
                ? 'w-2 h-2 bg-noah-teal opacity-50'
                : 'w-2 h-2 bg-noah-disabled'
          }`}
          aria-label={`ステップ ${i + 1}${i + 1 === current ? '（現在）' : ''}`}
        />
      ))}
    </div>
  );
}
