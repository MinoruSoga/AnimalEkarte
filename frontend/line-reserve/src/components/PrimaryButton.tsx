interface PrimaryButtonProps {
  onClick?: () => void;
  disabled?: boolean;
  type?: 'button' | 'submit';
  children: React.ReactNode;
  className?: string;
}

export function PrimaryButton({
  onClick,
  disabled = false,
  type = 'button',
  children,
  className = '',
}: PrimaryButtonProps) {
  return (
    <button
      type={type}
      onClick={onClick}
      disabled={disabled}
      className={`w-full py-3 px-4 rounded-lg font-semibold text-white transition-colors ${
        disabled
          ? 'bg-gray-300 cursor-not-allowed'
          : 'bg-line-green hover:bg-line-green-dark active:bg-line-green-dark'
      } ${className}`}
    >
      {children}
    </button>
  );
}
