import { Button } from "@/components/ui/button";
import { C } from "@/lib/design-tokens";
import { useDeleteOwnerTag } from "../api/delete-owner-tag";

interface LstepTagRemoveInlineProps {
  tagName: string;
  ownerId: string;
  onCancel: () => void;
}

export function LstepTagRemoveInline({ tagName, ownerId, onCancel }: LstepTagRemoveInlineProps) {
  const { mutate, isPending } = useDeleteOwnerTag(ownerId);

  const handleRemove = () => {
    mutate(tagName, { onSuccess: onCancel });
  };

  return (
    <div
      className={`flex items-center gap-2 rounded-md border ${C.borderNotice} ${C.bgNotice40} px-3 py-2 text-sm`}
    >
      <span className={C.text}>
        <span className="font-medium">{tagName}</span> を解除しますか？
      </span>
      <Button
        type="button"
        size="sm"
        variant="destructive"
        className="h-7 px-3 text-xs"
        disabled={isPending}
        onClick={handleRemove}
      >
        {isPending ? "解除中..." : "解除"}
      </Button>
      <Button
        type="button"
        size="sm"
        variant="ghost"
        className={`h-7 px-3 text-xs ${C.text60} ${C.hoverText}`}
        onClick={onCancel}
      >
        キャンセル
      </Button>
    </div>
  );
}
