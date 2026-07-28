import { Button, ButtonProps } from "@/components/ui/button";
import { C } from "@/lib/design-tokens";

interface PrimaryButtonProps extends ButtonProps {
  /**
   * "primary"（既定） — 汎用の主操作（brand と同じ primary teal + pill）。
   * "brand" — 認証など製品識別面の Brand CTA（teal + pill）。
   * "default" — "primary" の後方互換 alias。
   * SubmitButton と対称のプロパティ名・実装（単一 className 文字列を選択し連結しない）。
   */
  colorVariant?: "default" | "primary" | "brand";
}

export function PrimaryButton({ className, colorVariant = "primary", ...props }: PrimaryButtonProps) {
  const baseClassName =
    colorVariant === "brand"
      ? `${C.bgBrandIdentity} ${C.textOnBrandIdentity} ${C.hoverBgBrandIdentity} ${C.hoverTextOnBrandIdentity} ${C.activeBgBrandIdentity} ${C.activeTextOnBrandIdentity} h-11 text-xl font-bold shadow-none rounded-full border-transparent`
      : `${C.bgActionPrimary} ${C.textOnActionPrimary} ${C.hoverBgActionPrimary} ${C.hoverTextOnActionPrimary} ${C.activeBgActionPrimary} ${C.activeTextOnActionPrimary} h-11 text-base shadow-none rounded-full border-transparent`;

  return (
    <Button
      className={`${baseClassName} ${className || ""}`}
      {...props}
    />
  );
}
