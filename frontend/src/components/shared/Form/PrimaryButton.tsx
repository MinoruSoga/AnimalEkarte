import { Button, ButtonProps } from "@/components/ui/button";

export function PrimaryButton({ className, ...props }: ButtonProps) {
  return (
    <Button
      className={`bg-[#37352F] hover:bg-[#37352F]/90 text-white h-10 text-sm shadow-sm border-transparent ${className || ""}`}
      {...props}
    />
  );
}
