import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import { C } from "@/lib/design-tokens";

const DEFAULT_MESSAGE = "ログイン状態を確認しています";

interface SessionPendingProps {
  message?: string;
}

/** Non-sensitive branded shell while auth/session or login chunk is pending. Hook-free; no auth feature imports. */
export function SessionPending({ message = DEFAULT_MESSAGE }: SessionPendingProps) {
  return (
    <div
      className={`min-h-screen flex flex-col items-center justify-center px-4 ${C.bgPage}`}
      role="status"
      aria-live="polite"
      aria-atomic="true"
    >
      <div
        className={`inline-flex items-center justify-center size-[48px] rounded-xl mb-4 ${C.bgBrandIdentity}`}
      >
        <Stethoscope className={`size-[26px] ${C.textWhite}`} aria-hidden="true" />
      </div>
      <p className={`text-heading-3 font-bold leading-tight ${C.text} mb-1`}>ノア動物病院</p>
      <p className={`text-base ${C.text50}`}>{message}</p>
    </div>
  );
}
