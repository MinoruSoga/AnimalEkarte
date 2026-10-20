import { AlertTriangle, Info, OctagonAlert } from "lucide-react";
import type { ReactNode } from "react";
import { BADGE, C, ICON } from "@/lib/design-tokens";
import type { IssueNoticeCode } from "../lib/certificate-issue-reducer";
import { ISSUE_NOTICE } from "../lib/certificate-messages";

const TONE_STYLE = { error: BADGE.red, warning: BADGE.yellow, info: BADGE.blue } as const;
const TONE_ICON = { error: OctagonAlert, warning: AlertTriangle, info: Info } as const;

/** 色だけに頼らず、アイコンと見出し文で区別する通知。 */
export function CertificateNotice({
  code,
  children,
}: {
  code: IssueNoticeCode;
  children?: ReactNode;
}) {
  const { tone, title, detail } = ISSUE_NOTICE[code];
  const Icon = TONE_ICON[tone];
  return (
    <div role={tone === "info" ? "status" : "alert"} className={`rounded-xs border p-3 ${TONE_STYLE[tone]}`}>
      <p className="flex items-center gap-1.5 text-sm font-semibold">
        <Icon className={ICON.xs} aria-hidden="true" />
        {title}
      </p>
      <p className={`mt-1 text-sm ${C.text}`}>{detail}</p>
      {children}
    </div>
  );
}
