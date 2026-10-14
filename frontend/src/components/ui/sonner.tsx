import { Toaster as Sonner, ToasterProps } from "sonner";

const Toaster = ({ ...props }: ToasterProps) => {
  return (
    <Sonner
      theme="light"
      className="toaster group"
      closeButton
      // EMR-381: top-left だとサイドナビ/ページヘッダと重なるため右上（offset 72 でヘッダ下）へ。
      position="top-right"
      offset={72}
      // F003: モバイルの Sonner 既定 mobileOffset=16 ではヘッダと重なるため、
      // desktop と同じヘッダクリアランスを適用する。
      mobileOffset={72}
      duration={4000}
      style={
        {
          "--normal-bg": "var(--popover)",
          "--normal-text": "var(--popover-foreground)",
          "--normal-border": "var(--border)",
        } as React.CSSProperties
      }
      {...props}
    />
  );
};

export { Toaster };
