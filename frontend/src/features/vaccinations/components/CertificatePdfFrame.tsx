import { useEffect, useState } from "react";
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
 */
export function CertificatePdfFrame({ base64, blob, title }: CertificatePdfFrameProps) {
  const [state, setState] = useState<{ url: string | null; failed: boolean }>({ url: null, failed: false });

  useEffect(() => {
    let objectUrl: string | null = null;
    try {
      const source = blob ?? (base64 !== undefined ? base64ToPdfBlob(base64) : null);
      objectUrl = source === null ? null : URL.createObjectURL(source);
      setState({ url: objectUrl, failed: source === null });
    } catch {
      setState({ url: null, failed: true });
    }
    return () => {
      if (objectUrl !== null) URL.revokeObjectURL(objectUrl);
    };
  }, [base64, blob]);

  if (state.failed) {
    return (
      <p role="alert" className={`text-sm ${C.danger}`}>
        PDF を表示できませんでした。
      </p>
    );
  }
  if (state.url === null) return null;
  return (
    <iframe
      title={title}
      src={`${state.url}${PDF_VIEWER_FRAGMENT}`}
      className={`h-[32rem] w-full rounded-xs border ${C.borderLight} ${C.bgWhite}`}
    />
  );
}
