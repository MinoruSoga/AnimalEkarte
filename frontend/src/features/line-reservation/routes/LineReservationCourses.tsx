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
import { useGetCourses } from "../api/get-courses";
import { useCreateCourse } from "../api/create-course";
import { useUpdateCourse, usePatchCourseStatus, usePatchCourseSortOrder } from "../api/update-course";
import { useDeleteCourse } from "../api/delete-course";
import type { ServiceType, CreateServiceTypeRequest } from "../api/types";

// ── Constants ──

const DAY_OPTION_LABELS: Record<string, string> = {
  none: "なし",
  saturday: "土曜のみ",
  weekday: "平日のみ",
  anyday: "毎日",
};

// rendering-hoist-jsx: 静的 SelectItem JSX をモジュール定数に巻き上げ
const DAY_OPTION_SELECT_ITEMS = Object.entries(DAY_OPTION_LABELS).map(([value, label]) => (
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
      <span
        className={`size-1.5 rounded-full ${isActive ? C.bgBrand : C.bgInactive}`}
      />
      {isActive ? "有効" : "停止"}
    </span>
  );
}

// ── Course Form Dialog ──

interface CourseFormDialogProps {
  open: boolean;
  onClose: () => void;
  initial?: ServiceType | null;
  clinicId: string | null;
  onSaved: () => void;
}

function CourseFormDialog({
  open,
  onClose,
  initial,
  clinicId,
  onSaved,
}: CourseFormDialogProps) {
  const createCourse = useCreateCourse(clinicId);
  const updateCourse = useUpdateCourse(clinicId);

  const [formState, formAction] = useActionState(
    async (_prev: null, formData: FormData) => {
      const payload: CreateServiceTypeRequest = {
        name: formData.get("name") as string,
        short_name: formData.get("short_name") as string,
        duration_minutes: Number(formData.get("duration_minutes")),
        reservation_day_option: formData.get("reservation_day_option") as string,
        is_internal: formData.get("is_internal") === "true",
        reservation_visible: formData.get("reservation_visible") === "true",
        reservation_comment: formData.get("reservation_comment") as string,
        description: formData.get("description") as string,
        color: formData.get("color") as string,
      };
      try {
        if (initial) {
          await updateCourse.mutateAsync({
            id: initial.id,
            payload: { ...payload, is_active: initial.is_active, sort_order: initial.sort_order },
          });
          toast.success("コースを更新しました");
        } else {
          await createCourse.mutateAsync(payload);
          toast.success("コースを追加しました");
        }
        onSaved();
        onClose();
      } catch (err) {
        handleApiError(err, initial ? "コース更新" : "コース追加");
      }
      return null;
    },
    null
  );

  const [isInternal, setIsInternal] = useState(initial?.is_internal ?? false);
  const [reservationVisible, setReservationVisible] = useState(
    initial?.reservation_visible ?? true
  );
  const [dayOption, setDayOption] = useState(
    initial?.reservation_day_option ?? "anyday"
  );

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) onClose(); }}>
      <DialogContent className="max-w-[480px]">
        <DialogHeader>
          <DialogTitle>{initial ? "コース編集" : "コース追加"}</DialogTitle>
        </DialogHeader>
        <form action={formAction} className="space-y-4">
          <input type="hidden" name="is_internal" value={String(isInternal)} />
          <input type="hidden" name="reservation_visible" value={String(reservationVisible)} />
          <input type="hidden" name="reservation_day_option" value={dayOption} />
          <input type="hidden" name="color" value={initial?.color ?? "#038B94"} />

          <div className="space-y-1.5">
            <Label htmlFor="course-name">コース名 *</Label>
            <Input
              id="course-name"
              name="name"
              defaultValue={initial?.name ?? ""}
              required
              placeholder="例: 一般診療"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label htmlFor="course-short-name">略称</Label>
              <Input
                id="course-short-name"
                name="short_name"
                defaultValue={initial?.short_name ?? ""}
                placeholder="例: 診療"
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="course-duration">所要時間（分）</Label>
              <Input
                id="course-duration"
                name="duration_minutes"
                type="number"
                min={5}
                step={5}
                defaultValue={initial?.duration_minutes ?? 30}
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="day-option">予約可能曜日</Label>
            <Select value={dayOption} onValueChange={setDayOption}>
              <SelectTrigger id="day-option">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>{DAY_OPTION_SELECT_ITEMS}</SelectContent>
            </Select>
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="course-comment">予約ページコメント</Label>
            <Input
              id="course-comment"
              name="reservation_comment"
              defaultValue={initial?.reservation_comment ?? ""}
              placeholder="例: 初診の方もお気軽にどうぞ"
            />
          </div>

          <div className="flex items-center justify-between py-1">
            <Label htmlFor="is-internal">内部用コース（LINEに非表示）</Label>
            <Switch
              id="is-internal"
              checked={isInternal}
              onCheckedChange={setIsInternal}
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
        {/* satisfy eslint: formState is intentionally unused */}
        {formState !== undefined ? null : null}
      </DialogContent>
    </Dialog>
  );
}

// ── Main Page ──

export function LineReservationCourses() {
  const { currentClinicId } = useAuth();
  const { data: courses = [], isLoading } = useGetCourses(currentClinicId);
  const patchStatus = usePatchCourseStatus(currentClinicId);
  const patchSortOrder = usePatchCourseSortOrder(currentClinicId);
  const deleteCourse = useDeleteCourse(currentClinicId);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<ServiceType | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<ServiceType | null>(null);

  const handleEdit = useCallback((course: ServiceType) => {
    setEditTarget(course);
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
    (course: ServiceType) => {
      patchStatus.mutate(
        { id: course.id, isActive: !course.is_active },
        {
          onSuccess: () => toast.success(course.is_active ? "停止しました" : "有効にしました"),
          onError: (err) => handleApiError(err, "ステータス変更"),
        }
      );
    },
    [patchStatus]
  );

  const handleMoveUp = useCallback(
    (course: ServiceType) => {
      patchSortOrder.mutate(
        { id: course.id, sortOrder: course.sort_order - 1 },
        { onError: (err) => handleApiError(err, "並び順変更") }
      );
    },
    [patchSortOrder]
  );

  const handleMoveDown = useCallback(
    (course: ServiceType) => {
      patchSortOrder.mutate(
        { id: course.id, sortOrder: course.sort_order + 1 },
        { onError: (err) => handleApiError(err, "並び順変更") }
      );
    },
    [patchSortOrder]
  );

  const handleDeleteConfirm = useCallback(() => {
    if (!deleteTarget) return;
    deleteCourse.mutate(deleteTarget.id, {
      onSuccess: () => {
        toast.success("コースを削除しました");
        setDeleteTarget(null);
      },
      onError: (err) => {
        handleApiError(err, "コース削除");
        setDeleteTarget(null);
      },
    });
  }, [deleteTarget, deleteCourse]);

  const sortedCourses = useMemo(
    () => [...courses].sort((a, b) => a.sort_order - b.sort_order),
    [courses]
  );

  return (
    <PageLayout
      title="コース設定"
      headerAction={
        <Button size="sm" onClick={handleNew} className={STYLE.confirmPrimary}>
          <Plus className={ICON.action} />
          コース追加
        </Button>
      }
    >
      {isLoading ? (
        <div className={`text-sm ${C.textMuted} py-8 text-center`}>読み込み中...</div>
      ) : sortedCourses.length === 0 ? (
        <div className={`text-sm ${C.textMuted} py-8 text-center`}>
          コースが未登録です。「コース追加」から登録してください。
        </div>
      ) : (
        <div className={`rounded-lg border ${C.borderLight} overflow-hidden`}>
          <table className="w-full text-sm">
            <thead>
              <tr className={`${C.bgPage} border-b ${C.borderLight}`}>
                <th className={`px-4 py-3 text-left font-medium ${C.textMuted}`}>コース名</th>
                <th className={`px-4 py-3 text-left font-medium ${C.textMuted} w-[80px]`}>略称</th>
                <th className={`px-4 py-3 text-left font-medium ${C.textMuted} w-[90px]`}>時間</th>
                <th className={`px-4 py-3 text-left font-medium ${C.textMuted} w-[100px]`}>内部用</th>
                <th className={`px-4 py-3 text-left font-medium ${C.textMuted} w-[80px]`}>並び順</th>
                <th className={`px-4 py-3 text-left font-medium ${C.textMuted} w-[80px]`}>状態</th>
                <th className={`px-4 py-3 text-right font-medium ${C.textMuted} w-[180px]`}>操作</th>
              </tr>
            </thead>
            <tbody>
              {sortedCourses.map((course) => (
                <tr
                  key={course.id}
                  className={`border-b ${C.borderLight} last:border-0 ${C.hoverBgLight}`}
                >
                  <td className={`px-4 py-3 font-medium ${C.text}`}>{course.name}</td>
                  <td className={`px-4 py-3 ${C.textMuted}`}>{course.short_name || "-"}</td>
                  <td className={`px-4 py-3 ${C.textMuted}`}>{course.duration_minutes}分</td>
                  <td className="px-4 py-3">
                    {course.is_internal ? (
                      <span className={`text-xs ${C.textMuted}`}>内部用</span>
                    ) : (
                      <span className={`text-xs ${C.textBrand}`}>LINE表示</span>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-1">
                      <button
                        type="button"
                        aria-label="上に移動"
                        onClick={() => handleMoveUp(course)}
                        className={`p-0.5 rounded ${C.hoverBgMedium} ${C.textMuted} transition-colors`}
                      >
                        <ChevronUp className={ICON.action} />
                      </button>
                      <button
                        type="button"
                        aria-label="下に移動"
                        onClick={() => handleMoveDown(course)}
                        className={`p-0.5 rounded ${C.hoverBgMedium} ${C.textMuted} transition-colors`}
                      >
                        <ChevronDown className={ICON.action} />
                      </button>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <StatusBadge isActive={course.is_active} />
                  </td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex items-center justify-end gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleEdit(course)}
                        className="h-7 text-xs px-2"
                      >
                        編集
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleToggleStatus(course)}
                        className="h-7 text-xs px-2"
                      >
                        {course.is_active ? "停止" : "有効化"}
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => setDeleteTarget(course)}
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

      <CourseFormDialog
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
        title="コースを削除しますか？"
        description={`「${deleteTarget?.name ?? ""}」を削除します。この操作は元に戻せません。`}
        confirmLabel="削除"
        cancelLabel="キャンセル"
        variant="destructive"
        isPending={deleteCourse.isPending}
      />
    </PageLayout>
  );
}
