import { Button, ButtonProps } from "@/components/ui/button";
import { C } from "@/lib/design-tokens";

export function PrimaryButton({ className, ...props }: ButtonProps) {
  return (
    <Button
      className={`${C.bgAccent} ${C.bgAccentHover} ${C.textWhite} h-11 text-sm shadow-sm border-transparent ${className || ""}`}
      {...props}
    />
  );
}
