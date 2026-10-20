import { useEffect, useMemo, useRef } from "react";
import { C } from "@/lib/design-tokens";
import { PDF_VIEWER_FRAGMENT, base64ToPdfBlob } from "../lib/pdf";

interface CertificatePdfFrameProps {
  /** どちらか一方。base64 は preview 応答、blob は /content の PDF。 */
  base64?: string;
  blob?: Blob;
  title: string;
}

/**
 * server renderer が作った PDF を表示するだけの枠。FE は様式を再描画しない。
 * この枠自体に印刷ボタンは無い（確認 preview 用）。
 * Blob の決定は render 中の純粋な計算、object URL の作成と破棄は effect（iframe の src へ直接設定）で行う。
 */
export function CertificatePdfFrame({ base64, blob, title }: CertificatePdfFrameProps) {
  const source = useMemo<Blob | null>(() => {
    if (blob !== undefined) return blob;
    if (base64 === undefined) return null;
    try {
      return base64ToPdfBlob(base64);
    } catch {
      return null;
    }
  }, [base64, blob]);
  const frameRef = useRef<HTMLIFrameElement>(null);

  useEffect(() => {
    const frame = frameRef.current;
    if (source === null || frame === null) return;
    const objectUrl = URL.createObjectURL(source);
    frame.src = `${objectUrl}${PDF_VIEWER_FRAGMENT}`;
    return () => {
      URL.revokeObjectURL(objectUrl);
    };
  }, [source]);

  if (source === null) {
    return (
      <p role="alert" className={`text-sm ${C.danger}`}>
        PDF を表示できませんでした。
      </p>
    );
  }
  return <iframe ref={frameRef} title={title} className={`h-[32rem] w-full rounded-xs border ${C.borderLight} ${C.bgWhite}`} />;
}
