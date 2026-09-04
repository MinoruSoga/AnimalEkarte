/**
 * LinkedLineCustomers - 飼主詳細ページに表示するLINE連携セクション
 *
 * この飼主に紐付いたLINE顧客を表示し、紐付け/解除を行う。
 * app/pages/OwnerFormPage で line-reservation feature から注入する（依存逆転）。
 */
import { useState, useCallback, useLayoutEffect, useMemo, useRef, memo } from "react";
import { Link2, Link2Off, Search, MessageCircle } from "lucide-react";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { toast } from "sonner";
import { C, ICON, PALETTE } from "@/lib/design-tokens";
import { usePermission } from "@/hooks/use-permission";
import { EmptyState } from "@/components/shared/DataStates";
import { normalizeKana } from "@/lib/normalize-kana";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { useGetLineCustomers } from "../api/get-line-customers";
import { useUpdateOwnerLink } from "../api/update-owner-link";
import type { LineCustomer } from "../api/types";

interface LinkedLineCustomersProps {
  clinicId: string | null;
  ownerId: number;
}

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

export const LinkedLineCustomers = memo(function LinkedLineCustomers({
  clinicId,
  ownerId,
}: LinkedLineCustomersProps) {
  const { canEdit } = usePermission("owners");
  const canEditRef = useRef(canEdit);
  useLayoutEffect(() => {
    canEditRef.current = canEdit;
  }, [canEdit]);
  const { data: allCustomers = [] } = useGetLineCustomers(clinicId);
  const linkMutation = useUpdateOwnerLink(clinicId);
  const { mutate } = linkMutation;
  const [showLinkDialog, setShowLinkDialog] = useState(false);

  // この飼主に紐付いたLINE顧客
  const linked = useMemo(
    () => allCustomers.filter((c) => c.owner_id === ownerId),
    [allCustomers, ownerId],
  );

  // 未紐付けのLINE顧客（リンクダイアログ用）
  const unlinked = useMemo(() => allCustomers.filter((c) => !c.owner_id), [allCustomers]);

  const handleUnlink = useCallback(
    (customerId: number) => {
      if (canEditRef.current !== true) {
        toast.error(PERMISSION_DENIED_MESSAGE);
        return;
      }
      mutate({ customerId, ownerID: null });
    },
    [mutate],
  );

  const handleLink = useCallback(
    (customerId: number) => {
      if (canEditRef.current !== true) {
        toast.error(PERMISSION_DENIED_MESSAGE);
        return;
      }
      mutate({ customerId, ownerID: ownerId }, { onSuccess: () => setShowLinkDialog(false) });
    },
    [mutate, ownerId],
  );

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h3 className={`text-sm font-semibold ${C.text} flex items-center gap-1.5`}>
          <MessageCircle className={ICON.sm} style={{ color: PALETTE.lineGreen }} />
          LINE連携
        </h3>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => setShowLinkDialog(true)}
          disabled={unlinked.length === 0}
        >
          <Link2 className={`${ICON.smXs} mr-1`} />
          LINEアカウントを紐付け
        </Button>
      </div>

      {linked.length === 0 ? (
        <EmptyState message="紐付けされたLINEアカウントはありません" />
      ) : (
        <div className="space-y-2">
          {linked.map((c) => (
            <div
              key={c.id}
              className={`flex items-center justify-between p-3 rounded-lg border ${C.borderLight} ${C.bgPage}`}
            >
              <div className="flex items-center gap-3">
                <div
                  className="size-8 rounded-full flex items-center justify-center"
                  style={{ backgroundColor: PALETTE.lineGreen }}
                >
                  <MessageCircle className={`${ICON.sm} text-white`} />
                </div>
                <div>
                  <p className={`text-sm font-medium ${C.text}`}>{c.display_name || "名前なし"}</p>
                  <p className={`text-xs ${C.textMuted}`}>
                    {c.real_name ? `本名: ${c.real_name}` : ""}
                    {c.created_at
                      ? ` / 登録: ${format(new Date(c.created_at), "yyyy/MM/dd", { locale: ja })}`
                      : ""}
                  </p>
                </div>
              </div>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => handleUnlink(c.id)}
                disabled={linkMutation.isPending}
                className={C.danger}
              >
                <Link2Off className={`${ICON.smXs} mr-1`} />
                解除
              </Button>
            </div>
          ))}
        </div>
      )}

      {/* 紐付けダイアログ */}
      {showLinkDialog ? (
        <LinkSearchDialog
          unlinked={unlinked}
          onLink={handleLink}
          onClose={() => setShowLinkDialog(false)}
          isPending={linkMutation.isPending}
        />
      ) : null}
    </div>
  );
});

// ---- 未紐付けLINE顧客を検索して選択するダイアログ ----

interface LinkSearchDialogProps {
  unlinked: LineCustomer[];
  onLink: (customerId: number) => void;
  onClose: () => void;
  isPending: boolean;
}

function LinkSearchDialog({ unlinked, onLink, onClose, isPending }: LinkSearchDialogProps) {
  const [search, setSearch] = useState("");

  const filtered = useMemo(() => {
    const q = normalizeKana(search).toLowerCase();
    if (!q) return unlinked;
    return unlinked.filter(
      (c) =>
        normalizeKana(c.display_name).toLowerCase().includes(q) ||
        normalizeKana(c.real_name).toLowerCase().includes(q),
    );
  }, [unlinked, search]);

  return (
    <Dialog
      open
      onOpenChange={(v) => {
        if (!v) onClose();
      }}
    >
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>LINEアカウントを紐付け</DialogTitle>
          <DialogDescription className="sr-only">
            未紐付けのLINE顧客を検索して飼主に紐付けます。
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-3 py-2">
          <div className="relative">
            <Search className={`absolute left-2.5 top-2.5 ${ICON.sm} ${C.textMuted}`} />
            <Input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="LINE名・本名で検索..."
              className="pl-8"
            />
          </div>

          <div className="max-h-60 overflow-y-auto space-y-1">
            {filtered.length === 0 ? (
              <EmptyState message="未紐付けのLINE顧客が見つかりません" />
            ) : (
              filtered.map((c) => (
                <button
                  key={c.id}
                  type="button"
                  onClick={() => onLink(c.id)}
                  disabled={isPending}
                  className={`w-full flex items-center gap-3 p-3 rounded-lg text-left ${C.hoverBgLight} transition-colors disabled:opacity-50`}
                >
                  <div
                    className="size-8 rounded-full flex items-center justify-center flex-shrink-0"
                    style={{ backgroundColor: PALETTE.lineGreen }}
                  >
                    <MessageCircle className={`${ICON.sm} text-white`} />
                  </div>
                  <div className="min-w-0">
                    <p className={`text-sm font-medium ${C.text} truncate`}>
                      {c.display_name || "名前なし"}
                    </p>
                    {c.real_name ? (
                      <p className={`text-xs ${C.textMuted} truncate`}>{c.real_name}</p>
                    ) : null}
                  </div>
                </button>
              ))
            )}
          </div>
        </div>

        <DialogFooter>
          <Button type="button" variant="ghost" onClick={onClose}>
            閉じる
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
