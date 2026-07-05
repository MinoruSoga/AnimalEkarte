import { useFormStatus } from "react-dom";
import { Button, type ButtonProps } from "@/components/ui/button";
import { C, STYLE } from "@/lib/design-tokens";

interface SubmitButtonProps extends Omit<ButtonProps, "variant"> {
  loadingText?: string;
  /**
   * "brand" — DESIGN.md `button-primary`（brand blue `#0075DE` + pill `{rounded.full}`）。
   * 既定の "default"（STYLE.confirmPrimary、旧 accent ブルー + 4px 角丸）はアプリ全体で共有されるため変更しない。
   * 単一の className 文字列を選択（連結しない）ことで Tailwind の同一 specificity クラス競合を避ける。
   * shadcn Button 自身の `variant`（outline/ghost 等）と名前が衝突するため `colorVariant` という名前にしている。
   */
  colorVariant?: "default" | "brand";
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
  colorVariant = "default",
  ...props
}: SubmitButtonProps) {
  const { pending } = useFormStatus();

  const baseClassName =
    colorVariant === "brand"
      ? `${C.bgBrand} ${C.hoverBgBrand} text-white h-11 text-base rounded-full transition-colors shadow-none border-transparent`
      : STYLE.confirmPrimary;

  return (
    <Button
      type="submit"
      disabled={pending || disabled}
      className={`${baseClassName} px-4 ${className || ""}`}
      {...props}
    >
      {pending ? loadingText : children}
    </Button>
  );
}
