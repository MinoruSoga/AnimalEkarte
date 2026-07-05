import { Button, ButtonProps } from "@/components/ui/button";
import { C } from "@/lib/design-tokens";

interface PrimaryButtonProps extends ButtonProps {
  /**
   * "brand" — DESIGN.md `button-primary`（brand blue `#0075DE` + pill `{rounded.full}`）。
   * 既定の "default"（旧 accent ブルー + 4px 角丸）はアプリ全体で共有されるため変更しない。
   * SubmitButton と対称のプロパティ名・実装（単一 className 文字列を選択し連結しない）。
   */
  colorVariant?: "default" | "brand";
}

export function PrimaryButton({ className, colorVariant = "default", ...props }: PrimaryButtonProps) {
  const baseClassName =
    colorVariant === "brand"
      ? `${C.bgBrand} ${C.hoverBgBrand} ${C.textWhite} h-11 text-sm shadow-none rounded-full border-transparent`
      : `${C.bgAccent} ${C.bgAccentHover} ${C.textWhite} h-11 text-sm shadow-sm border-transparent`;

  return (
    <Button
      className={`${baseClassName} ${className || ""}`}
      {...props}
    />
  );
}
