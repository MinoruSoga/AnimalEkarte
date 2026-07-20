import { Button, ButtonProps } from "@/components/ui/button";
import { C } from "@/lib/design-tokens";

interface PrimaryButtonProps extends ButtonProps {
  /**
   * "brand"（既定） — DESIGN.md `button-primary`（brand blue #0075DE + pill `{rounded.full}`・FE10 字義準拠）。
   * "default" — 旧 accent 系トークン + 角丸（`C.bgAccent` — FE10 で値は brand に統合済み）。新規実装では使用しないこと。
   * SubmitButton と対称のプロパティ名・実装（単一 className 文字列を選択し連結しない）。
   */
  colorVariant?: "default" | "brand";
}

export function PrimaryButton({ className, colorVariant = "brand", ...props }: PrimaryButtonProps) {
  const baseClassName =
    colorVariant === "brand"
      ? `${C.bgBrand} ${C.hoverBgBrand} ${C.textWhite} h-11 text-sm shadow-none rounded-full border-transparent`
      : `${C.bgAccent} ${C.bgAccentHover} ${C.textWhite} h-11 text-sm shadow-none border-transparent`;

  return (
    <Button
      className={`${baseClassName} ${className || ""}`}
      {...props}
    />
  );
}
