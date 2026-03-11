// React/Framework
import { useState, useEffect } from "react";
import { useNavigate } from "react-router";

// External
import { Plus, UserRound, Save, Trash2 } from "lucide-react";
import { toast } from "sonner";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { TableCell } from "@/components/ui/table";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable, DataTableRow } from "@/components/shared/DataTable";
import { PrimaryButton } from "@/components/shared/Form";
import { StatusBadge } from "@/components/shared/StatusBadge";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { getMasterStatusColor } from "@/utils/status-helpers";
import { useMasterItems } from "@/hooks/use-master-items";

// Types
import type { MasterItem } from "@/types";

const STAFF_ROLES = [
  { value: "院長", label: "院長" },
  { value: "獣医師", label: "獣医師" },
  { value: "看護師", label: "看護師" },
  { value: "受付", label: "受付" },
  { value: "トリマー", label: "トリマー" },
  { value: "運営管理者", label: "運営管理者" },
] as const;

const COLUMNS = [
  { header: "社員番号", className: "w-[130px]" },
  { header: "氏名" },
  { header: "職種", className: "w-[130px]" },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

export function StaffSettings() {
  const navigate = useNavigate();
  const [isEditing, setIsEditing] = useState(false);
  const [selectedItem, setSelectedItem] = useState<MasterItem | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [formData, setFormData] = useState<Partial<MasterItem>>({});
  const [pendingDelete, setPendingDelete] = useState<MasterItem | null>(null);

  const { data: filteredItems, add, update, remove } = useMasterItems("staff", searchTerm);

  useEffect(() => {
    if (!isEditing) {
      setSelectedItem(null);
      setFormData({});
    }
  }, [isEditing]);

  const handleEdit = (item: MasterItem) => {
    setSelectedItem(item);
    setFormData({ ...item });
    setIsEditing(true);
  };

  const handleCreate = () => {
    setSelectedItem(null);
    setFormData({
      category: "獣医師",
      status: "active",
      price: 0,
    });
    setIsEditing(true);
  };

  const handleCloseEdit = () => {
    setIsEditing(false);
  };

  const handleSave = () => {
    if (!formData.code || !formData.name) {
      toast.error("社員番号と氏名は必須です");
      return;
    }

    if (selectedItem) {
      update(selectedItem.id, formData, {
        onSuccess: () => {
          toast.success("更新しました");
          setIsEditing(false);
        },
      });
    } else {
      add(formData as Omit<MasterItem, "id">, {
        onSuccess: () => {
          toast.success("登録しました");
          setIsEditing(false);
        },
      });
    }
  };

  const handleDeleteConfirm = () => {
    if (!pendingDelete) return;
    remove(pendingDelete.id, {
      onSuccess: () => {
        setPendingDelete(null);
        setIsEditing(false);
        toast.success("削除しました");
      },
    });
  };

  // ── Edit Form ──
  if (isEditing) {
    return (
      <>
        <PageLayout
          title={selectedItem ? "スタッフ 編集" : "スタッフ 新規登録"}
          onBack={handleCloseEdit}
          maxWidth="max-w-2xl"
          align="left"
        >
          <div className="space-y-4 py-4">
            <div className="bg-white p-6 rounded-lg border border-[rgba(55,53,47,0.16)] space-y-4 shadow-sm">
              {/* 社員番号 / 氏名 */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <Label className="text-sm text-[#37352F]/60">
                    社員番号 <span className="text-[#E03E3E]">*</span>
                  </Label>
                  <Input
                    className="h-10 text-sm bg-white text-[#37352F] border-[rgba(55,53,47,0.16)]"
                    value={formData.code ?? ""}
                    onChange={(e) => setFormData({ ...formData, code: e.target.value })}
                    placeholder="例: ST-001"
                  />
                </div>
                <div className="space-y-1.5">
                  <Label className="text-sm text-[#37352F]/60">
                    氏名 <span className="text-[#E03E3E]">*</span>
                  </Label>
                  <Input
                    className="h-10 text-sm bg-white text-[#37352F] border-[rgba(55,53,47,0.16)]"
                    value={formData.name ?? ""}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder="例: 山田 太郎"
                  />
                </div>
              </div>

              {/* 職種 */}
              <div className="space-y-1.5">
                <Label className="text-sm text-[#37352F]/60">職種</Label>
                <Select
                  value={formData.category ?? "獣医師"}
                  onValueChange={(val) => setFormData({ ...formData, category: val })}
                >
                  <SelectTrigger className="h-10 text-sm bg-white text-[#37352F] border-[rgba(55,53,47,0.16)]">
                    <SelectValue placeholder="職種を選択" />
                  </SelectTrigger>
                  <SelectContent>
                    {STAFF_ROLES.map((role) => (
                      <SelectItem key={role.value} value={role.value}>
                        {role.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* 備考 */}
              <div className="space-y-1.5">
                <Label className="text-sm text-[#37352F]/60">備考</Label>
                <Input
                  className="h-10 text-sm bg-white text-[#37352F] border-[rgba(55,53,47,0.16)]"
                  value={formData.description ?? ""}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  placeholder="補足情報など"
                />
              </div>

              {/* ステータス */}
              <div className="space-y-1.5">
                <Label className="text-sm text-[#37352F]/60">ステータス</Label>
                <div className="flex items-center gap-6 pt-2">
                  <label className="flex items-center gap-2 cursor-pointer text-sm text-[#37352F]">
                    <input
                      type="radio"
                      name="staff-status"
                      checked={formData.status !== "inactive"}
                      onChange={() => setFormData({ ...formData, status: "active" })}
                      className="w-4 h-4"
                    />
                    <span>有効</span>
                  </label>
                  <label className="flex items-center gap-2 cursor-pointer text-sm text-[#37352F]">
                    <input
                      type="radio"
                      name="staff-status"
                      checked={formData.status === "inactive"}
                      onChange={() => setFormData({ ...formData, status: "inactive" })}
                      className="w-4 h-4"
                    />
                    <span>無効</span>
                  </label>
                </div>
              </div>

              {/* Actions */}
              <div className="flex justify-between pt-4 border-t border-[rgba(55,53,47,0.09)]">
                {selectedItem ? (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setPendingDelete(selectedItem)}
                    className="h-10 text-sm text-[#E03E3E] hover:bg-[#E03E3E]/10 hover:text-[#E03E3E]"
                  >
                    <Trash2 className="mr-1.5 size-4" />
                    削除
                  </Button>
                ) : (
                  <div />
                )}
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handleCloseEdit}
                    className="h-10 text-sm bg-white text-[#37352F] border-[rgba(55,53,47,0.16)]"
                  >
                    キャンセル
                  </Button>
                  <PrimaryButton onClick={handleSave} className="gap-2">
                    <Save className="size-4" />
                    保存
                  </PrimaryButton>
                </div>
              </div>
            </div>
          </div>
        </PageLayout>

        <ConfirmDialog
          open={pendingDelete !== null}
          onClose={() => setPendingDelete(null)}
          title="スタッフを削除しますか？"
          description={`「${pendingDelete?.name}」を削除します。この操作は取り消せません。`}
          confirmLabel="削除"
          variant="destructive"
          onConfirm={handleDeleteConfirm}
        />
      </>
    );
  }

  // ── List View ──
  return (
    <>
      <PageLayout
        title="スタッフマスタ"
        icon={<UserRound className="size-5 text-[#37352F]" />}
        onBack={() => navigate("/settings")}
        headerAction={
          <PrimaryButton onClick={handleCreate}>
            <Plus className="mr-1.5 size-4" />
            新規登録
          </PrimaryButton>
        }
        maxWidth="max-w-full"
      >
        <div className="flex flex-col gap-4">
          <SearchFilterBar
            searchTerm={searchTerm}
            onSearchChange={setSearchTerm}
            placeholder="社員番号、氏名、職種で検索..."
            count={filteredItems.length}
          />

          <DataTable
            columns={COLUMNS}
            data={filteredItems}
            emptyMessage="スタッフが登録されていません"
            renderRow={(item) => (
              <DataTableRow key={item.id} onClick={() => handleEdit(item)}>
                <TableCell className="font-mono text-sm text-[#37352F]/80 py-2.5">
                  {item.code}
                </TableCell>
                <TableCell className="font-medium text-sm text-[#37352F] py-2.5">
                  {item.name}
                </TableCell>
                <TableCell className="text-sm text-[#37352F] py-2.5">
                  {item.category}
                </TableCell>
                <TableCell className="text-center py-2.5">
                  <StatusBadge colorClass={getMasterStatusColor(item.status)}>
                    {item.status === "active" ? "有効" : "無効"}
                  </StatusBadge>
                </TableCell>
                <TableCell className="text-right py-2.5">
                  <RowActionButton onClick={() => handleEdit(item)} />
                </TableCell>
              </DataTableRow>
            )}
          />
        </div>
      </PageLayout>

      <ConfirmDialog
        open={pendingDelete !== null}
        onClose={() => setPendingDelete(null)}
        title="スタッフを削除しますか？"
        description={`「${pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </>
  );
}
