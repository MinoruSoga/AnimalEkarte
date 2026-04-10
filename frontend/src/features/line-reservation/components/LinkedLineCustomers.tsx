/**
 * LinkedLineCustomers - 飼主詳細ページに表示するLINE連携セクション
 *
 * この飼主に紐付いたLINE顧客を表示し、紐付け/解除を行う。
 * app/pages/OwnerFormPage で line-reservation feature から注入する（依存逆転）。
 */
import { useState, useCallback, useMemo, memo } from "react";
import { Link2, Link2Off, Search, MessageCircle } from "lucide-react";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { C, PALETTE } from "@/lib/design-tokens";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { useGetReservationCustomers } from "../api/get-reservation-customers";
import { useUpdateOwnerLink } from "../api/update-owner-link";
import type { ReservationCustomer } from "../api/types";

interface LinkedLineCustomersProps {
  clinicId: string | null;
  ownerId: number;
}

export const LinkedLineCustomers = memo(function LinkedLineCustomers({ clinicId, ownerId }: LinkedLineCustomersProps) {
  const { data: allCustomers = [] } = useGetReservationCustomers(clinicId);
  const linkMutation = useUpdateOwnerLink(clinicId);
  const [showLinkDialog, setShowLinkDialog] = useState(false);

  // この飼主に紐付いたLINE顧客
  const linked = useMemo(
    () => allCustomers.filter((c) => c.owner_id === ownerId),
    [allCustomers, ownerId],
  );

  // 未紐付けのLINE顧客（リンクダイアログ用）
  const unlinked = useMemo(
    () => allCustomers.filter((c) => !c.owner_id),
    [allCustomers],
  );

  const handleUnlink = useCallback(
    (customerId: number) => {
      linkMutation.mutate({ customerId, ownerID: null });
    },
    [linkMutation],
  );

  const handleLink = useCallback(
    (customerId: number) => {
      linkMutation.mutate(
        { customerId, ownerID: ownerId },
        { onSuccess: () => setShowLinkDialog(false) },
      );
    },
    [linkMutation, ownerId],
  );

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h3 className={`text-sm font-semibold ${C.text} flex items-center gap-1.5`}>
          <MessageCircle className="size-4" style={{ color: PALETTE.lineGreen }} />
          LINE連携
        </h3>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => setShowLinkDialog(true)}
          disabled={unlinked.length === 0}
        >
          <Link2 className="size-3.5 mr-1" />
          LINEアカウントを紐付け
        </Button>
      </div>

      {linked.length === 0 ? (
        <p className={`text-sm ${C.textMuted} py-2`}>
          紐付けされたLINEアカウントはありません
        </p>
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
                  <MessageCircle className="size-4 text-white" />
                </div>
                <div>
                  <p className={`text-sm font-medium ${C.text}`}>
                    {c.display_name || "名前なし"}
                  </p>
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
                <Link2Off className="size-3.5 mr-1" />
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
  unlinked: ReservationCustomer[];
  onLink: (customerId: number) => void;
  onClose: () => void;
  isPending: boolean;
}

function LinkSearchDialog({ unlinked, onLink, onClose, isPending }: LinkSearchDialogProps) {
  const [search, setSearch] = useState("");

  const filtered = useMemo(
    () => {
      const q = search.toLowerCase();
      if (!q) return unlinked;
      return unlinked.filter(
        (c) =>
          c.display_name.toLowerCase().includes(q) ||
          c.real_name.toLowerCase().includes(q),
      );
    },
    [unlinked, search],
  );

  return (
    <Dialog open onOpenChange={(v) => { if (!v) onClose(); }}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>LINEアカウントを紐付け</DialogTitle>
        </DialogHeader>

        <div className="space-y-3 py-2">
          <div className="relative">
            <Search className={`absolute left-2.5 top-2.5 size-4 ${C.textMuted}`} />
            <Input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="LINE名・本名で検索..."
              className="pl-8"
            />
          </div>

          <div className="max-h-60 overflow-y-auto space-y-1">
            {filtered.length === 0 ? (
              <p className={`text-sm ${C.textMuted} text-center py-4`}>
                未紐付けのLINE顧客が見つかりません
              </p>
            ) : (
              filtered.map((c) => (
                <button
                  key={c.id}
                  type="button"
                  onClick={() => onLink(c.id)}
                  disabled={isPending}
                  className={`w-full flex items-center gap-3 p-3 rounded-lg text-left hover:${C.bgHover} transition-colors disabled:opacity-50`}
                >
                  <div
                    className="size-8 rounded-full flex items-center justify-center flex-shrink-0"
                    style={{ backgroundColor: PALETTE.lineGreen }}
                  >
                    <MessageCircle className="size-4 text-white" />
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
