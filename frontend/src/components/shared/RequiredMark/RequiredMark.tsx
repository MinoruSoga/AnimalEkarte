import { C } from "@/lib/design-tokens";

export function RequiredMark() {
  return (
    <span aria-hidden="true" className={C.textRequired}>
      *
    </span>
  );
}
