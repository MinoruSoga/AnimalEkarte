import { C } from "@/lib/design-tokens";
import type { MedicalRecordAddendum } from "@/types/generated/models";

interface AddendumItemProps {
  addendum: MedicalRecordAddendum;
}

function formatDateTime(iso: string): string {
  if (!iso) return "-";
  return iso.slice(0, 19).replace("T", " ").replace(/-/g, "/");
}

export function AddendumItem({ addendum }: AddendumItemProps) {
  const formattedDate = formatDateTime(addendum.created_at);

  return (
    <div
      className={`border ${C.borderLight} rounded-xs p-4 space-y-3`}
      data-testid="addendum-item"
    >
      <div className={`flex items-center gap-2 text-sm ${C.text60}`}>
        <time dateTime={addendum.created_at}>{formattedDate}</time>
        <span aria-hidden="true">·</span>
        <span>スタッフID: {addendum.author_user_id}</span>
      </div>
      <div>
        <p className={`text-sm font-medium ${C.text60} mb-1`}>修正理由</p>
        <p className={`text-base ${C.text}`}>{addendum.reason}</p>
      </div>
      <div>
        <p className={`text-sm font-medium ${C.text60} mb-1`}>修正内容</p>
        <p className={`text-base ${C.text} whitespace-pre-wrap`}>{addendum.after_text}</p>
      </div>
    </div>
  );
}
