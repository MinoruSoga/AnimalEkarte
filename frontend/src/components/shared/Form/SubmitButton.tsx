import { useFormStatus } from "react-dom";
import { Button, type ButtonProps } from "@/components/ui/button";
import { STYLE } from "@/lib/design-tokens";

interface SubmitButtonProps extends ButtonProps {
  loadingText?: string;
}

/**
 * React 19 useFormStatus を使用して、
 * 親フォームの送信状態に応じて自動で disabled/loading 表示を切り替えるボタン。
 */
export function SubmitButton({
  children,
  loadingText = "保存中...",
  className,
  disabled,
  ...props
}: SubmitButtonProps) {
  const { pending } = useFormStatus();

  return (
    <Button
      type="submit"
      disabled={pending || disabled}
      className={`${STYLE.confirmPrimary} px-4 ${className || ""}`}
      {...props}
    >
      {pending ? loadingText : children}
    </Button>
  );
}
