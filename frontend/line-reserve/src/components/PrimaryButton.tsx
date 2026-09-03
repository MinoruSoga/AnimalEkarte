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
          ? 'bg-noah-disabled cursor-not-allowed'
          : 'bg-noah-teal hover:bg-noah-teal-dark active:bg-noah-teal-dark'
      } ${className}`}
    >
      {children}
    </button>
  );
}
