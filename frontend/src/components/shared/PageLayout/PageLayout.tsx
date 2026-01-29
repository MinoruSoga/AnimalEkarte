import { ReactNode } from "react";
import { FormHeader } from "@/components/shared/Form";

interface PageLayoutProps {
  children: ReactNode;
  title: string;
  description?: string;
  onBack?: () => void;
  icon?: ReactNode;
  headerAction?: ReactNode;
  maxWidth?: string;
  className?: string;
  align?: "center" | "left";
}

export const PageLayout = ({
  children,
  title,
  description: _description,
  onBack,
  icon,
  headerAction,
  maxWidth = "max-w-[1440px]",
  className,
  align = "center",
}: PageLayoutProps) => {
  return (
    <div
      className={`flex flex-col h-full bg-[#fafafa] overflow-hidden ${className || ""}`}
    >
      <FormHeader
        title={title}
        onBack={onBack}
        icon={icon}
        action={headerAction}
      />
      <div className="flex-1 w-full">
        <div
          className={`${maxWidth} ${align === "center" ? "mx-auto" : ""} w-full p-2`}
        >
          {children}
        </div>
      </div>
    </div>
  );
};
