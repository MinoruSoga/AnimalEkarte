import { Button, ButtonProps } from "@/components/ui/button";

export function PrimaryButton({ className, ...props }: ButtonProps) {
  return (
    <Button
      className={`bg-[#2383E2] hover:bg-[#1B6EC2] text-white h-11 text-sm shadow-sm border-transparent ${className || ""}`}
      {...props}
    />
  );
}
