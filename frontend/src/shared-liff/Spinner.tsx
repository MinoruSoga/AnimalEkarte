type SpinnerSize = "sm" | "md";

interface SpinnerProps {
  size?: SpinnerSize;
  borderColorClassName?: string;
}

const SIZE_CLASS_NAME: Record<SpinnerSize, string> = {
  sm: "w-8 h-8 mb-3",
  md: "w-10 h-10 mb-4",
};

const DEFAULT_BORDER_COLOR_CLASS_NAME = "border-liff-brand";

/** liff / line-reserve 共通のローディングスピナー（FE5-22: 3箇所のコピペを統合） */
export function Spinner({
  size = "md",
  borderColorClassName = DEFAULT_BORDER_COLOR_CLASS_NAME,
}: SpinnerProps) {
  return (
    <div
      className={`${SIZE_CLASS_NAME[size]} border-4 ${borderColorClassName} border-t-transparent rounded-full animate-spin mx-auto`}
      aria-hidden="true"
    />
  );
}
