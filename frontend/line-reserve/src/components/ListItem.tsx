interface ListItemProps {
  onClick: () => void;
  children: React.ReactNode;
  subtitle?: string;
  description?: string;
  imageUrl?: string;
  imageFallback?: string;
}

export function ListItem({
  onClick,
  children,
  subtitle,
  description,
  imageUrl,
  imageFallback,
}: ListItemProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="w-full flex items-center justify-between px-4 py-4 border-b border-noah-border hover:bg-noah-teal-light active:bg-noah-teal-light text-left"
    >
      <div className="flex items-center gap-3 flex-1 min-w-0">
        {imageUrl ? (
          <img
            src={imageUrl}
            alt=""
            className="w-12 h-12 rounded-lg object-cover flex-shrink-0 bg-noah-surface-subtle"
            onError={(e) => {
              if (imageFallback) {
                (e.currentTarget as HTMLImageElement).src = imageFallback;
              } else {
                e.currentTarget.style.display = "none";
              }
            }}
          />
        ) : null}
        <div className="min-w-0">
          <div className="text-noah-text-strong font-medium">{children}</div>
          {subtitle ? <div className="text-sm text-noah-text-muted mt-0.5">{subtitle}</div> : null}
          {description ? (
            <div className="text-xs text-noah-text-faint mt-1 line-clamp-2">{description}</div>
          ) : null}
        </div>
      </div>
      <svg
        xmlns="http://www.w3.org/2000/svg"
        className="h-5 w-5 text-noah-text-faint flex-shrink-0 ml-2"
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
