import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { C, LAYOUT, STYLE } from "@/lib/design-tokens";
import type { CheckupType } from "../api/get-checkup-sync-preview";

const CHECKUP_TYPE_LABELS: Record<CheckupType, string> = {
  annual: "定期健診",
  dental: "デンタルケア",
  blood: "血液検査",
  skin: "皮膚チェック",
  cancer: "ガン検診",
  other: "その他",
};

interface CheckupSyncConfirmDialogProps {
  open: boolean;
  onOpenChange: (o: boolean) => void;
  checkupType: CheckupType;
  selectedCount: number;
  tagName: string;
  onTagNameChange: (name: string) => void;
  onConfirm: () => void;
  isPending: boolean;
}

export function CheckupSyncConfirmDialog({
  open,
  onOpenChange,
  checkupType,
  selectedCount,
  tagName,
  onTagNameChange,
  onConfirm,
  isPending,
}: CheckupSyncConfirmDialogProps) {
  const isTagNameEmpty = tagName.trim() === "";

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className={LAYOUT.modal.md}>
        <DialogHeader>
          <DialogTitle className={`text-base font-semibold ${C.text}`}>
            Lステップタグ一括付与の確認
          </DialogTitle>
          <DialogDescription>
            選択した対象者へ付与するLステップタグ名を確認します。
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-1">
          {/* タグ名入力 */}
          <div className="space-y-1.5">
            <label htmlFor="checkup-sync-tag-name" className={STYLE.formLabel}>
              付与するタグ名
              <span className={`ml-1 ${C.textRequired}`}>*</span>
            </label>
            <input
              id="checkup-sync-tag-name"
              type="text"
              value={tagName}
              onChange={(e) => onTagNameChange(e.target.value)}
              className={`w-full ${STYLE.formInput} border rounded-xs px-3 ${isTagNameEmpty ? STYLE.formInputError : ""}`}
              placeholder="例: checkup_annual_2026"
            />
            {isTagNameEmpty ? (
              <p className={`text-sm ${C.danger}`}>タグ名を入力してください</p>
            ) : null}
          </div>

          {/* 確認メッセージ */}
          <div className={`rounded-xs p-4 ${C.bgBrand10} border ${C.borderBrand}`}>
            <p className={`text-sm ${C.text}`}>
              <span className={`font-semibold ${C.textBrand}`}>
                {CHECKUP_TYPE_LABELS[checkupType]}
              </span>
              の対象者{" "}
              <span className="font-semibold">{selectedCount}名</span>
              にLステップタグ「
              <span className={`font-mono ${C.textBrand}`}>{tagName || "—"}</span>
              」を付与します。
            </p>
          </div>
        </div>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isPending}
            className={STYLE.btnOutline}
          >
            キャンセル
          </Button>
          <Button
            type="button"
            onClick={onConfirm}
            disabled={isTagNameEmpty || isPending}
            className={`${C.bgBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} ${C.textOnBrand} h-11 px-4 text-base rounded-full transition-colors shadow-none border-transparent`}
          >
            {isPending ? "付与中..." : "タグを一括付与する"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
