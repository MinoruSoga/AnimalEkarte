import { useFormStatus } from "react-dom";
import { Button, type ButtonProps } from "@/components/ui/button";
import { C } from "@/lib/design-tokens";

interface SubmitButtonProps extends Omit<ButtonProps, "variant"> {
  loadingText?: string;
  /**
   * "primary"（既定） — 汎用の主操作（brand と同じ primary teal + pill）。
   * "brand" — 認証など製品識別面の Brand CTA（teal + pill）。
   * "destructive" — 破壊的操作（死亡記録・削除確認など）。
   * "default" — "primary" の後方互換 alias。
   * 単一の className 文字列を選択（連結しない）ことで Tailwind の同一 specificity クラス競合を避ける。
   * shadcn Button 自身の `variant`（outline/ghost 等）と名前が衝突するため `colorVariant` という名前にしている。
   */
  colorVariant?: "default" | "primary" | "brand" | "destructive";
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
  colorVariant = "primary",
  ...props
}: SubmitButtonProps) {
  const { pending } = useFormStatus();

  const baseClassName =
    colorVariant === "brand"
      ? `${C.bgBrandIdentity} ${C.textOnBrandIdentity} ${C.hoverBgBrandIdentity} ${C.hoverTextOnBrandIdentity} ${C.activeBgBrandIdentity} ${C.activeTextOnBrandIdentity} h-11 text-xl font-bold rounded-full transition-colors shadow-none border-transparent`
      : colorVariant === "destructive"
        ? `${C.bgDanger} ${C.textWhite} ${C.hoverBgDanger90} h-11 text-base rounded-full transition-colors shadow-none border-transparent`
        : `${C.bgActionPrimary} ${C.textOnActionPrimary} ${C.hoverBgActionPrimary} ${C.hoverTextOnActionPrimary} ${C.activeBgActionPrimary} ${C.activeTextOnActionPrimary} h-11 text-base rounded-full transition-colors shadow-none border-transparent`;

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
