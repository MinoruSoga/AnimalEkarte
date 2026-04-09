interface ListItemProps {
  onClick: () => void;
  children: React.ReactNode;
  subtitle?: string;
}

export function ListItem({ onClick, children, subtitle }: ListItemProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="w-full flex items-center justify-between px-4 py-4 border-b border-gray-200 hover:bg-noah-teal-light active:bg-noah-teal-light text-left"
    >
      <div>
        <div className="text-gray-800 font-medium">{children}</div>
        {subtitle ? <div className="text-sm text-gray-500 mt-0.5">{subtitle}</div> : null}
      </div>
      <svg
        xmlns="http://www.w3.org/2000/svg"
        className="h-5 w-5 text-gray-400 flex-shrink-0"
        viewBox="0 0 20 20"
        fill="currentColor"
        aria-hidden="true"
      >
        <path
          fillRule="evenodd"
          d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z"
          clipRule="evenodd"
        />
      </svg>
    </button>
  );
}
