import { useState, useCallback, useMemo } from "react";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { Link2, Link2Off, Search } from "lucide-react";
import { C, STYLE } from "@/lib/design-tokens";
import { useAuth } from "@/features/auth";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { Owner } from "@/types/owner";
import { useGetReservationCustomers } from "../api/get-reservation-customers";
import { useUpdateOwnerLink } from "../api/update-owner-link";
import type { ReservationCustomer } from "../api/types";

// ── Types ──

interface LineReservationCustomersProps {
  owners: Owner[];
}

// ── Owner Link Dialog ──

interface LinkOwnerDialogProps {
  customer: ReservationCustomer;
  owners: Owner[];
  clinicId: string | null;
  onClose: () => void;
}

function LinkOwnerDialog({ customer, owners, clinicId, onClose }: LinkOwnerDialogProps) {
  const [search, setSearch] = useState("");
  const [selectedOwnerIdStr, setSelectedOwnerIdStr] = useState<string>(
    customer.owner_id ? String(customer.owner_id) : "",
  );
  const linkMutation = useUpdateOwnerLink(clinicId);

  const filteredOwners = useMemo(
    () =>
      owners.filter((o) =>
        o.ownerName.toLowerCase().includes(search.toLowerCase()),
      ),
    [owners, search],
  );

  const handleSave = useCallback(() => {
    const ownerID = selectedOwnerIdStr ? Number(selectedOwnerIdStr) : null;
    linkMutation.mutate(
      { customerId: customer.id, ownerID },
      { onSuccess: onClose },
    );
  }, [customer.id, selectedOwnerIdStr, linkMutation, onClose]);

  const handleUnlink = useCallback(() => {
    linkMutation.mutate(
      { customerId: customer.id, ownerID: null },
      { onSuccess: onClose },
    );
  }, [customer.id, linkMutation, onClose]);

  return (
    <Dialog open onOpenChange={(v) => { if (!v) onClose(); }}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>オーナー紐付け</DialogTitle>
        </DialogHeader>

        <div className="space-y-4 py-2">
          <div>
            <p className={`text-sm font-medium ${C.text}`}>LINE顧客</p>
            <p className={`text-sm mt-0.5 ${C.text60}`}>
              {customer.display_name}
              {customer.real_name ? ` (${customer.real_name})` : ""}
            </p>
          </div>

          <div className="space-y-2">
            <Label>オーナー検索</Label>
            <div className="relative">
              <Search className={`absolute left-2.5 top-2.5 size-4 ${C.textMuted}`} />
              <Input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="オーナー名で検索..."
                className="pl-8"
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <Label>紐付け先オーナー</Label>
            <Select value={selectedOwnerIdStr} onValueChange={setSelectedOwnerIdStr}>
              <SelectTrigger>
                <SelectValue placeholder="オーナーを選択..." />
              </SelectTrigger>
              <SelectContent>
                {filteredOwners.map((o) => (
                  <SelectItem key={o.id} value={o.id}>
                    {o.ownerName}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>

        <DialogFooter className={`${STYLE.FLEX_BETWEEN} gap-2`}>
          {customer.owner_id ? (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={handleUnlink}
              disabled={linkMutation.isPending}
              className={C.danger}
            >
              <Link2Off className="size-3.5 mr-1" />
              紐付け解除
            </Button>
          ) : (
            <span />
          )}
          <div className="flex gap-2">
            <Button type="button" variant="ghost" onClick={onClose}>
              キャンセル
            </Button>
            <Button
              type="button"
              disabled={!selectedOwnerIdStr || linkMutation.isPending}
              onClick={handleSave}
            >
              保存
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// ── Customer Row ──

interface CustomerRowProps {
  customer: ReservationCustomer;
  ownerName: string | undefined;
  onLink: (c: ReservationCustomer) => void;
}

function CustomerRow({ customer, ownerName, onLink }: CustomerRowProps) {
  const handleLink = useCallback(() => onLink(customer), [customer, onLink]);

  const linkedAt = customer.created_at
    ? format(new Date(customer.created_at), "yyyy/MM/dd HH:mm", { locale: ja })
    : "—";

  return (
    <tr className={`border-b ${C.borderLight} hover:${C.bgHover} transition-colors`}>
      <td className={`py-3 px-4 text-sm ${C.text}`}>
        {customer.display_name || "—"}
      </td>
      <td className={`py-3 px-4 text-sm ${C.text60}`}>
        {customer.real_name || "—"}
      </td>
      <td className="py-3 px-4">
        {ownerName ? (
          <span
            className={`inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-full font-medium ${C.bgBrand10} ${C.textBrand}`}
          >
            <Link2 className="size-3" />
            {ownerName}
          </span>
        ) : (
          <span className={`text-xs ${C.textMuted}`}>未紐付け</span>
        )}
      </td>
      <td className={`py-3 px-4 text-xs ${C.textMuted}`}>{linkedAt}</td>
      <td className="py-3 px-4 text-right">
        <Button type="button" variant="ghost" size="sm" onClick={handleLink}>
          <Link2 className="size-3.5 mr-1" />
          紐付け
        </Button>
      </td>
    </tr>
  );
}

// ── Main Component ──

export function LineReservationCustomers({ owners }: LineReservationCustomersProps) {
  const { user } = useAuth();
  const clinicId = user?.clinic_id ? String(user.clinic_id) : null;

  const [filterText, setFilterText] = useState("");
  const [linkTarget, setLinkTarget] = useState<ReservationCustomer | null>(null);

  const { data: customers = [], isLoading } = useGetReservationCustomers(clinicId);

  // owner_id → ownerName のマップ（O(1) lookup）
  const ownerMap = useMemo(() => {
    const m = new Map<string, string>();
    for (const o of owners) {
      m.set(o.id, o.ownerName);
    }
    return m;
  }, [owners]);

  const filteredCustomers = useMemo(
    () =>
      customers.filter((c) => {
        const q = filterText.toLowerCase();
        return (
          c.display_name.toLowerCase().includes(q) ||
          c.real_name.toLowerCase().includes(q)
        );
      }),
    [customers, filterText],
  );

  const handleLink = useCallback((c: ReservationCustomer) => setLinkTarget(c), []);
  const handleCloseLink = useCallback(() => setLinkTarget(null), []);

  return (
    <PageLayout title="LINE顧客管理" description="LINE予約を利用した顧客の一覧とオーナー紐付けを管理します。">
      {/* Filter */}
      <div className="flex items-center gap-3 mb-4">
        <div className="relative max-w-xs w-full">
          <Search className={`absolute left-2.5 top-2.5 size-4 ${C.textMuted}`} />
          <Input
            value={filterText}
            onChange={(e) => setFilterText(e.target.value)}
            placeholder="名前で絞り込み..."
            className="pl-8"
          />
        </div>
        <p className={`text-sm ml-auto ${C.textMuted}`}>
          {filteredCustomers.length} 件
        </p>
      </div>

      {/* Table */}
      <div className={`rounded-lg border ${C.borderLight} overflow-hidden`}>
        <table className="w-full text-left">
          <thead>
            <tr className={`border-b ${C.borderLight} ${C.bgPage}`}>
              <th className={`py-3 px-4 text-xs font-semibold ${C.textMuted}`}>LINE名</th>
              <th className={`py-3 px-4 text-xs font-semibold ${C.textMuted}`}>本名</th>
              <th className={`py-3 px-4 text-xs font-semibold ${C.textMuted}`}>紐付けオーナー</th>
              <th className={`py-3 px-4 text-xs font-semibold ${C.textMuted}`}>登録日時</th>
              <th className="py-3 px-4" />
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr>
                <td
                  colSpan={5}
                  className={`py-12 text-center text-sm ${C.textMuted}`}
                >
                  読み込み中...
                </td>
              </tr>
            ) : filteredCustomers.length === 0 ? (
              <tr>
                <td
                  colSpan={5}
                  className={`py-12 text-center text-sm ${C.textMuted}`}
                >
                  LINE顧客が見つかりません
                </td>
              </tr>
            ) : (
              filteredCustomers.map((c) => (
                <CustomerRow
                  key={c.id}
                  customer={c}
                  ownerName={c.owner_id ? ownerMap.get(String(c.owner_id)) : undefined}
                  onLink={handleLink}
                />
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Link Dialog */}
      {linkTarget ? (
        <LinkOwnerDialog
          customer={linkTarget}
          owners={owners}
          clinicId={clinicId}
          onClose={handleCloseLink}
        />
      ) : null}
    </PageLayout>
  );
}
