import { useState, useCallback, useActionState, useMemo } from "react";
import { toast } from "sonner";
import { Plus, ChevronUp, ChevronDown } from "lucide-react";
import { handleApiError } from "@/lib/handle-api-error";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { useAuth } from "@/features/auth";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
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
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useGetReservationStaffs } from "../api/get-staffs";
import { useCreateReservationStaff } from "../api/create-staff";
import { useUpdateReservationStaff, usePatchStaffStatus, usePatchStaffSortOrder } from "../api/update-staff";
import { useDeleteReservationStaff } from "../api/delete-staff";
import type { Staff, CreateStaffRequest } from "../api/types";

// ── Constants ──

const STAFF_TYPE_LABELS: Record<string, string> = {
  doctor: "医師",
  nurse: "看護師",
  resource: "リソース",
};

// rendering-hoist-jsx: 静的 SelectItem JSX をモジュール定数に巻き上げ
const STAFF_TYPE_SELECT_ITEMS = Object.entries(STAFF_TYPE_LABELS).map(([value, label]) => (
  <SelectItem key={value} value={value}>
    {label}
  </SelectItem>
));

// ── Status Badge ──

function StatusBadge({ isActive }: { isActive: boolean }) {
  return (
    <span
      className={`inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-full font-medium ${
        isActive
          ? `${C.bgBrand10} ${C.textBrand}`
          : `${C.bgInactive} ${C.textMuted}`
      }`}
    >
      <span className={`size-1.5 rounded-full ${isActive ? C.bgBrand : C.bgInactive}`} />
      {isActive ? "有効" : "停止"}
    </span>
  );
}

// ── Staff Form Dialog ──

interface StaffFormDialogProps {
  open: boolean;
  onClose: () => void;
  initial?: Staff | null;
  clinicId: string | null;
  onSaved: () => void;
}

function StaffFormDialog({ open, onClose, initial, clinicId, onSaved }: StaffFormDialogProps) {
  const createStaff = useCreateReservationStaff(clinicId);
  const updateStaff = useUpdateReservationStaff(clinicId);

  const [staffType, setStaffType] = useState(initial?.staff_type ?? "doctor");
  const [reservationVisible, setReservationVisible] = useState(
    initial?.reservation_visible ?? true
  );

  const [formState, formAction] = useActionState(
    async (_prev: null, formData: FormData) => {
      const payload: CreateStaffRequest = {
        name: formData.get("name") as string,
        staff_type: staffType,
        reservation_visible: reservationVisible,
        reservation_comment: formData.get("reservation_comment") as string,
      };
      try {
        if (initial) {
          await updateStaff.mutateAsync({
            id: initial.id,
            payload: { ...payload, is_active: initial.is_active, sort_order: initial.sort_order },
          });
          toast.success("スタッフを更新しました");
        } else {
          await createStaff.mutateAsync(payload);
          toast.success("スタッフを追加しました");
        }
        onSaved();
        onClose();
      } catch (err) {
        handleApiError(err, initial ? "スタッフ更新" : "スタッフ追加");
      }
      return null;
    },
    null
  );

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) onClose(); }}>
      <DialogContent className="max-w-[440px]">
        <DialogHeader>
          <DialogTitle>{initial ? "スタッフ編集" : "スタッフ追加"}</DialogTitle>
        </DialogHeader>
        <form action={formAction} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="staff-name">名前 *</Label>
            <Input
              id="staff-name"
              name="name"
              defaultValue={initial?.name ?? ""}
              required
              placeholder="例: 田中 太郎"
            />
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="staff-type">種別</Label>
            <Select value={staffType} onValueChange={setStaffType}>
              <SelectTrigger id="staff-type">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>{STAFF_TYPE_SELECT_ITEMS}</SelectContent>
            </Select>
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="staff-comment">コメント</Label>
            <Input
              id="staff-comment"
              name="reservation_comment"
              defaultValue={initial?.reservation_comment ?? ""}
              placeholder="例: 小動物専門"
            />
          </div>

          <div className="flex items-center justify-between py-1">
            <Label htmlFor="res-visible">予約ページに表示</Label>
            <Switch
              id="res-visible"
              checked={reservationVisible}
              onCheckedChange={setReservationVisible}
            />
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              キャンセル
            </Button>
            <SubmitButton>{initial ? "更新" : "追加"}</SubmitButton>
          </DialogFooter>
        </form>
        {formState !== undefined ? null : null}
      </DialogContent>
    </Dialog>
  );
}

// ── Main Page ──

export function LineReservationStaffs() {
  const { currentClinicId } = useAuth();
  const { data: staffs = [], isLoading } = useGetReservationStaffs(currentClinicId);
  const patchStatus = usePatchStaffStatus(currentClinicId);
  const patchSortOrder = usePatchStaffSortOrder(currentClinicId);
  const deleteStaff = useDeleteReservationStaff(currentClinicId);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<Staff | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Staff | null>(null);

  const handleEdit = useCallback((staff: Staff) => {
    setEditTarget(staff);
    setDialogOpen(true);
  }, []);

  const handleNew = useCallback(() => {
    setEditTarget(null);
    setDialogOpen(true);
  }, []);

  const handleSaved = useCallback(() => {
    setEditTarget(null);
  }, []);

  const handleToggleStatus = useCallback(
    (staff: Staff) => {
      patchStatus.mutate(
        { id: staff.id, isActive: !staff.is_active },
        {
          onSuccess: () => toast.success(staff.is_active ? "停止しました" : "有効にしました"),
          onError: (err) => handleApiError(err, "ステータス変更"),
        }
      );
    },
    [patchStatus]
  );

  const handleMoveUp = useCallback(
    (staff: Staff) => {
      patchSortOrder.mutate(
        { id: staff.id, sortOrder: staff.sort_order - 1 },
        { onError: (err) => handleApiError(err, "並び順変更") }
      );
    },
    [patchSortOrder]
  );

  const handleMoveDown = useCallback(
    (staff: Staff) => {
      patchSortOrder.mutate(
        { id: staff.id, sortOrder: staff.sort_order + 1 },
        { onError: (err) => handleApiError(err, "並び順変更") }
      );
    },
    [patchSortOrder]
  );

  const handleDeleteConfirm = useCallback(() => {
    if (!deleteTarget) return;
    deleteStaff.mutate(deleteTarget.id, {
      onSuccess: () => {
        toast.success("スタッフを削除しました");
        setDeleteTarget(null);
      },
      onError: (err) => {
        handleApiError(err, "スタッフ削除");
        setDeleteTarget(null);
      },
    });
  }, [deleteTarget, deleteStaff]);

  const sortedStaffs = useMemo(
    () => [...staffs].sort((a, b) => a.sort_order - b.sort_order),
    [staffs]
  );

  return (
    <PageLayout
      title="スタッフ設定"
      headerAction={
        <Button size="sm" onClick={handleNew} className={STYLE.confirmPrimary}>
          <Plus className={ICON.action} />
          スタッフ追加
        </Button>
      }
    >
      {isLoading ? (
        <div className={`text-sm ${C.textMuted} py-8 text-center`}>読み込み中...</div>
      ) : sortedStaffs.length === 0 ? (
        <div className={`text-sm ${C.textMuted} py-8 text-center`}>
          スタッフが未登録です。「スタッフ追加」から登録してください。
        </div>
      ) : (
        <div className={`rounded-lg border ${C.borderLight} overflow-hidden`}>
          <table className="w-full text-sm">
            <thead>
              <tr className={`${C.bgPage} border-b ${C.borderLight}`}>
                <th className={`px-4 py-3 text-left font-medium ${C.textMuted}`}>名前</th>
                <th className={`px-4 py-3 text-left font-medium ${C.textMuted} w-[100px]`}>種別</th>
                <th className={`px-4 py-3 text-left font-medium ${C.textMuted} w-[80px]`}>並び順</th>
                <th className={`px-4 py-3 text-left font-medium ${C.textMuted} w-[80px]`}>状態</th>
                <th className={`px-4 py-3 text-right font-medium ${C.textMuted} w-[180px]`}>操作</th>
              </tr>
            </thead>
            <tbody>
              {sortedStaffs.map((staff) => (
                <tr
                  key={staff.id}
                  className={`border-b ${C.borderLight} last:border-0 ${C.hoverBgLight}`}
                >
                  <td className={`px-4 py-3 font-medium ${C.text}`}>{staff.name}</td>
                  <td className={`px-4 py-3 ${C.textMuted}`}>
                    {STAFF_TYPE_LABELS[staff.staff_type] ?? staff.staff_type}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-1">
                      <button
                        type="button"
                        aria-label="上に移動"
                        onClick={() => handleMoveUp(staff)}
                        className={`p-0.5 rounded ${C.hoverBgMedium} ${C.textMuted} transition-colors`}
                      >
                        <ChevronUp className={ICON.action} />
                      </button>
                      <button
                        type="button"
                        aria-label="下に移動"
                        onClick={() => handleMoveDown(staff)}
                        className={`p-0.5 rounded ${C.hoverBgMedium} ${C.textMuted} transition-colors`}
                      >
                        <ChevronDown className={ICON.action} />
                      </button>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <StatusBadge isActive={staff.is_active} />
                  </td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex items-center justify-end gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleEdit(staff)}
                        className="h-7 text-xs px-2"
                      >
                        編集
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleToggleStatus(staff)}
                        className="h-7 text-xs px-2"
                      >
                        {staff.is_active ? "停止" : "有効化"}
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => setDeleteTarget(staff)}
                        className={`h-7 text-xs px-2 ${C.danger}`}
                      >
                        削除
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <StaffFormDialog
        open={dialogOpen}
        onClose={() => { setDialogOpen(false); setEditTarget(null); }}
        initial={editTarget}
        clinicId={currentClinicId}
        onSaved={handleSaved}
      />

      <ConfirmDialog
        open={deleteTarget !== null}
        onClose={() => setDeleteTarget(null)}
        onConfirm={handleDeleteConfirm}
        title="スタッフを削除しますか？"
        description={`「${deleteTarget?.name ?? ""}」を削除します。この操作は元に戻せません。`}
        confirmLabel="削除"
        cancelLabel="キャンセル"
        variant="destructive"
        isPending={deleteStaff.isPending}
      />
    </PageLayout>
  );
}
