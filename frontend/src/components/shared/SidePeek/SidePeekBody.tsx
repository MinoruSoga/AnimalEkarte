import type { ReactNode } from "react";
import { STYLE } from "@/lib/design-tokens";

export function SidePeekBody({ children }: { children: ReactNode }) {
  return (
    <div className={STYLE.sidePeekBody}>
      <div className="px-16 pb-8">{children}</div>
    </div>
  );
}
