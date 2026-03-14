// React/Framework
import { Link } from "react-router";

// External
import { useQuery } from "@tanstack/react-query";
import { ExternalLink } from "lucide-react";

// Internal
import { axios } from "@/lib/axios";
import { C } from "@/lib/design-tokens";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import type { Owner as BackendOwner } from "@/types/generated/models";

interface OwnerQuickViewModalProps {
  open: boolean;
  onClose: () => void;
  ownerId: string;
}

interface OwnerQuickViewData {
  id: string;
  ownerName: string;
  ownerNameKana: string;
  phone: string;
  email: string;
  postalCode: string;
  address1: string;
  address2: string;
  membershipType: string;
  remarks: string;
}

const MEMBERSHIP_TYPE_LABELS: Record<string, string> = {
  non_member: "非会員",
  member: "会員",
  deceased: "退亡者",
  transferred: "他診/準",
};

async function fetchOwnerQuickView(id: string): Promise<OwnerQuickViewData> {
  const { data } = await axios.get<BackendOwner>(`/v1/owners/${id}`);
  return {
    id: String(data.id ?? 0),
    ownerName: data.owner_name ?? "",
    ownerNameKana: data.owner_name_kana ?? "",
    phone: data.phone ?? "",
    email: data.email ?? "",
    postalCode: data.postal_code ?? "",
    address1: data.address1 ?? "",
    address2: data.address2 ?? "",
    membershipType:
      MEMBERSHIP_TYPE_LABELS[data.membership_type ?? ""] ??
      data.membership_type ??
      "",
    remarks: data.remarks ?? "",
  };
}

interface FieldRowProps {
  label: string;
  value: string;
}

function FieldRow({ label, value }: FieldRowProps) {
  return (
    <div className={`flex gap-2 py-2 px-2 -mx-2 rounded-[3px] min-h-[40px] items-start`}>
      <span className={`w-[140px] shrink-0 text-sm ${C.text65} select-none`}>
        {label}
      </span>
      <span className={`flex-1 text-sm ${C.text} break-all`}>
        {value || <span className={C.text40}>-</span>}
      </span>
    </div>
  );
}

export function OwnerQuickViewModal({
  open,
  onClose,
  ownerId,
}: OwnerQuickViewModalProps) {
  const { data: owner, isLoading } = useQuery({
    queryKey: ["owner-quick-view", ownerId],
    queryFn: () => fetchOwnerQuickView(ownerId),
    enabled: open && ownerId !== "",
    staleTime: 5 * 60 * 1000,
  });

  const address = [owner?.address1, owner?.address2].filter(Boolean).join(" ");

  return (
    <Sheet open={open} onOpenChange={(isOpen) => { if (!isOpen) onClose(); }}>
      <SheetContent side="right" className="w-[400px] sm:max-w-[400px] flex flex-col p-0">
        <SheetHeader className="px-5 pt-5 pb-3 border-b border-[rgba(55,53,47,0.09)] shrink-0">
          <SheetTitle className={`text-base font-semibold ${C.text}`}>
            飼い主情報
          </SheetTitle>
        </SheetHeader>

        <div className="flex-1 overflow-y-auto px-5 py-3">
          {isLoading && (
            <div className={`flex items-center justify-center h-32 text-sm ${C.text40}`}>
              読み込み中...
            </div>
          )}

          {!isLoading && owner && (
            <div className={`divide-y ${C.divideDivider}`}>
              <FieldRow label="飼主No" value={owner.id} />
              <FieldRow label="飼主名" value={owner.ownerName} />
              <FieldRow label="飼主名(カナ)" value={owner.ownerNameKana} />
              <FieldRow label="電話番号" value={owner.phone} />
              <FieldRow label="メールアドレス" value={owner.email} />
              <FieldRow label="郵便番号" value={owner.postalCode} />
              <FieldRow label="住所" value={address} />
              <FieldRow label="会員区分" value={owner.membershipType} />
              <FieldRow label="備考" value={owner.remarks} />
            </div>
          )}

          {!isLoading && !owner && (
            <div className={`flex items-center justify-center h-32 text-sm ${C.text40}`}>
              飼い主情報が見つかりません
            </div>
          )}
        </div>

        {owner && (
          <div className={`shrink-0 px-5 py-4 border-t ${C.borderLight}`}>
            <Link
              to={`/owners/${ownerId}`}
              onClick={onClose}
              className={`inline-flex items-center gap-1.5 text-sm ${C.accent} hover:underline underline-offset-2 decoration-[#2383E2]/50`}
            >
              詳細を見る
              <ExternalLink className="size-3.5" />
            </Link>
          </div>
        )}
      </SheetContent>
    </Sheet>
  );
}
