// React/Framework
import { useState, lazy, Suspense, memo, useCallback, useEffect } from "react";
import { useNavigate, useParams, useLoaderData } from "react-router";

// External
import { Plus, Edit, User, PawPrint, MoreHorizontal, Calendar, FileText, Scissors, Bed, CreditCard, Trash2 } from "lucide-react";
import { toast } from "sonner";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionDatePicker } from "@/components/shared/NotionDatePicker";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { useTitle } from "@/hooks/use-title";
import { usePostalCodeLookup } from "@/hooks/use-postal-code-lookup";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { usePermission } from "@/features/auth";

// Relative
import { useOwnerForm } from "../hooks/use-owner-form";
import type { PetMutations } from "@/types/pet";
import type { PetFormData, OwnerData } from "../types";
// Lazy-loaded modal — only loaded when first opened (bundle-dynamic-imports)
const PetEditModal = lazy(() =>
  import("../components/PetEditModal").then(m => ({ default: m.PetEditModal }))
);
import { MEMBERSHIP_TYPE_VALUES } from "../types";
import type { MembershipType } from "../types";
import type { OwnerLoaderData } from "../loaders";
import { ResourceOwners } from "@/types/generated/models";

// rerender-memo: 会員区分ボタンを memo 化して ownerData の他フィールド変更による
// 不要な再レンダリングと inline onClick 生成を排除する
interface MembershipTypeButtonsProps {
  value: MembershipType;
  onChange: (type: MembershipType) => void;
}

const MembershipTypeButtons = memo(function MembershipTypeButtons({
  value,
  onChange,
}: MembershipTypeButtonsProps) {
  return (
    <div className="flex gap-1.5 flex-wrap">
      {MEMBERSHIP_TYPE_VALUES.map((type) => (
        <Button
          key={type}
          type="button"
          variant={value === type ? "default" : "outline"}
          size="sm"
          className={
            value === type
              ? `${STYLE.confirmPrimary} text-sm px-3`
              : `h-11 text-sm ${C.text} ${C.hoverBgMedium} ${C.borderMedium} px-3`
          }
          onClick={() => onChange(type)}
        >
          {type}
        </Button>
      ))}
    </div>
  );
});

// rerender-memo: ペット行を memo 化して pets.map 内の N×8 インライン関数生成を排除
interface PetTableRowProps {
  pet: PetFormData;
  ownerId: string | undefined;
  canEdit: boolean;
  canCreate: boolean;
  canDelete: boolean;
  onEdit: (pet: PetFormData) => void;
  onDeleteRequest: (id: string, name: string) => void;
}

const PetTableRow = memo(function PetTableRow({
  pet,
  ownerId,
  canEdit,
  canCreate,
  canDelete,
  onEdit,
  onDeleteRequest,
}: PetTableRowProps) {
  const navigate = useNavigate();
  const backFrom = ownerId ? `/owners/${ownerId}` : "/owners";

  return (
    <TableRow
      className={`transition-colors ${C.borderDivider} ${C.hoverBgPage} h-12 ${canEdit ? "cursor-pointer" : "cursor-default"}`}
      onClick={canEdit ? () => onEdit(pet) : undefined}
    >
      <TableCell className={STYLE.tableCell}>{pet.petNumber}</TableCell>
      <TableCell className={STYLE.tableCell}>{pet.petName}</TableCell>
      <TableCell className={STYLE.tableCell}>{pet.status}</TableCell>
      <TableCell className={STYLE.tableCell}>{pet.species}</TableCell>
      <TableCell className={STYLE.tableCell}>{pet.gender}</TableCell>
      <TableCell className={STYLE.tableCell}>
        {pet.birthDate ? pet.birthDate.slice(0, 10) : ""}
      </TableCell>
      <TableCell className={STYLE.tableCell}>{pet.color}</TableCell>
      <TableCell className={STYLE.tableCell}>
        {pet.weight ? `${pet.weight} kg` : ""}
      </TableCell>
      <TableCell className={STYLE.tableCell}>{pet.environment}</TableCell>
      <TableCell className={`${STYLE.tableCell} truncate max-w-[200px]`}>
        {pet.remarks}
      </TableCell>
      <TableCell className="py-2">
        <div className="flex gap-1 justify-end" onClick={(e) => e.stopPropagation()}>
          <DropdownMenu>
            <DropdownMenuTrigger
              className={`inline-flex items-center justify-center rounded-[4px] cursor-pointer ${STYLE.tableActionBtn} ${C.hoverBgLight}`}
              aria-label="操作メニューを開く"
            >
              <MoreHorizontal className={ICON.page} />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>操作</DropdownMenuLabel>
              {canEdit ? (
                <DropdownMenuItem onClick={() => onEdit(pet)}>
                  <Edit className={`mr-2 ${ICON.action}`} />
                  詳細・編集
                </DropdownMenuItem>
              ) : null}
              {canCreate ? (
                <>
                  <DropdownMenuItem onClick={() => navigate(`/reservations?petId=${pet.id}`)}>
                    <Calendar className={`mr-2 ${ICON.action}`} />
                    予約作成
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => navigate(`/medical-records/new?petId=${pet.id}`, { state: { from: backFrom } })}
                  >
                    <FileText className={`mr-2 ${ICON.action}`} />
                    カルテ作成
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => navigate(`/trimming/new?petId=${pet.id}`, { state: { from: backFrom } })}
                  >
                    <Scissors className={`mr-2 ${ICON.action}`} />
                    トリミング
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => navigate(`/hospitalization/new?petId=${pet.id}`, { state: { from: backFrom } })}
                  >
                    <Bed className={`mr-2 ${ICON.action}`} />
                    入院登録
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => navigate(`/accounting/new?petId=${pet.id}`, { state: { from: backFrom } })}
                  >
                    <CreditCard className={`mr-2 ${ICON.action}`} />
                    会計登録
                  </DropdownMenuItem>
                </>
              ) : null}
              {canDelete ? (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    onClick={() => onDeleteRequest(pet.id, pet.petName)}
                    className={`${C.danger} focus:${C.danger} ${C.focusBgLight}`}
                  >
                    <Trash2 className={`mr-2 ${ICON.action}`} />
                    削除
                  </DropdownMenuItem>
                </>
              ) : null}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </TableCell>
    </TableRow>
  );
});

// rendering-hoist-jsx: ペットテーブルの列ヘッダーは静的なので毎レンダーの再生成を排除
const PET_TABLE_HEADER = (
  <TableHeader>
    <TableRow className={`hover:bg-transparent ${C.bgPage} border-b ${C.borderMedium} h-12`}>
      <TableHead className={STYLE.tableCellMuted}>ペット番号</TableHead>
      <TableHead className={STYLE.tableCellMuted}>ペット名</TableHead>
      <TableHead className={STYLE.tableCellMuted}>生死</TableHead>
      <TableHead className={STYLE.tableCellMuted}>種別</TableHead>
      <TableHead className={STYLE.tableCellMuted}>性別</TableHead>
      <TableHead className={STYLE.tableCellMuted}>生年月日</TableHead>
      <TableHead className={STYLE.tableCellMuted}>毛色</TableHead>
      <TableHead className={STYLE.tableCellMuted}>体重</TableHead>
      <TableHead className={STYLE.tableCellMuted}>環境</TableHead>
      <TableHead className={STYLE.tableCellMuted}>備考</TableHead>
      <TableHead className={STYLE.tableCellMuted}>操作</TableHead>
    </TableRow>
  </TableHeader>
);

// rerender-memo: 飼主情報フィールド群を memo 化する
// ペット追加/削除・モーダル開閉・deletePetTarget 変更では ownerData/fieldErrors が変わらないため
// 17 フィールド全体の再レンダーを防ぐ（clearFieldError は useOwnerForm 内で useCallback 済み）
interface OwnerInfoSectionProps {
  ownerData: OwnerData;
  fieldErrors: Record<string, string>;
  isEdit: boolean;
  onChange: (field: string, value: string | boolean | number) => void;
  onClearError: (field: string) => void;
  onMembershipChange: (type: MembershipType) => void;
  onPostalCodeLookup: (postalCodeField: string, addressField: string) => void;
}

const OwnerInfoSection = memo(function OwnerInfoSection({
  ownerData,
  fieldErrors,
  isEdit,
  onChange,
  onClearError,
  onMembershipChange,
  onPostalCodeLookup,
}: OwnerInfoSectionProps) {
  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
      {/* Row 1 */}
      <div className="space-y-1.5">
        <Label htmlFor="ownerId" className={`text-sm ${C.text60}`}>飼主No</Label>
        {isEdit ? (
          <Input
            id="ownerId"
            type="text"
            value={ownerData.ownerId}
            disabled
            className={`${STYLE.formInput} disabled:opacity-50`}
          />
        ) : (
          <p className={`flex h-9 items-center px-3 text-sm ${C.text40} italic`}>
            登録時に自動採番されます
          </p>
        )}
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="postalCode" className={`text-sm ${C.text60}`}>郵便番号</Label>
        <div className="flex gap-1.5">
          <Input
            id="postalCode"
            placeholder="123-4567"
            value={ownerData.postalCode}
            aria-invalid={!!fieldErrors.postalCode}
            aria-describedby={fieldErrors.postalCode ? "postalCode-error" : undefined}
            onChange={(e) => { onChange("postalCode", e.target.value); onClearError("postalCode"); }}
            className={`${STYLE.formInput} ${fieldErrors.postalCode ? STYLE.formInputError : ""}`}
          />
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="shrink-0 h-9 text-xs px-2"
            onClick={() => onPostalCodeLookup("postalCode", "address1")}
          >
            検索
          </Button>
        </div>
        <FormFieldError id="postalCode-error" message={fieldErrors.postalCode} />
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="company" className={`text-sm ${C.text60}`}>会社名</Label>
        <Input
          id="company"
          value={ownerData.company}
          onChange={(e) => onChange("company", e.target.value)}
          className={STYLE.formInput}
        />
      </div>
      <div className="space-y-1.5">
        <Label className={`text-sm ${C.text60}`}>会員区分</Label>
        <MembershipTypeButtons value={ownerData.membershipType} onChange={onMembershipChange} />
      </div>

      {/* Row 2 */}
      <div className="space-y-1.5">
        <Label htmlFor="ownerName" className={`text-sm ${C.text60}`}>
          飼主名 <span className={C.textRequired}>*</span>
        </Label>
        <Input
          id="ownerName"
          value={ownerData.ownerName}
          maxLength={100}
          aria-invalid={!!fieldErrors.ownerName}
          aria-describedby={fieldErrors.ownerName ? "ownerName-error" : undefined}
          onChange={(e) => { onChange("ownerName", e.target.value); onClearError("ownerName"); }}
          className={`${STYLE.formInput} ${fieldErrors.ownerName ? STYLE.formInputError : ""}`}
        />
        <FormFieldError id="ownerName-error" message={fieldErrors.ownerName} />
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="address1" className={`text-sm ${C.text60}`}>住所1（会社）</Label>
        <Input id="address1" value={ownerData.address1} onChange={(e) => onChange("address1", e.target.value)} maxLength={200} className={STYLE.formInput} />
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="homePostalCode" className={`text-sm ${C.text60}`}>郵便番号(自宅)</Label>
        <div className="flex gap-1.5">
          <Input
            id="homePostalCode"
            placeholder="123-4567"
            value={ownerData.homePostalCode || ""}
            aria-invalid={!!fieldErrors.homePostalCode}
            aria-describedby={fieldErrors.homePostalCode ? "homePostalCode-error" : undefined}
            onChange={(e) => { onChange("homePostalCode", e.target.value); onClearError("homePostalCode"); }}
            className={`${STYLE.formInput} ${fieldErrors.homePostalCode ? STYLE.formInputError : ""}`}
          />
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="shrink-0 h-9 text-xs px-2"
            onClick={() => onPostalCodeLookup("homePostalCode", "homeAddress1")}
          >
            検索
          </Button>
        </div>
        <FormFieldError id="homePostalCode-error" message={fieldErrors.homePostalCode} />
      </div>
      <div className="space-y-1.5 col-span-2 lg:col-span-1 lg:row-span-3">
        <Label className={`text-sm ${C.text60}`}>危険人物</Label>
        <div className="flex items-center space-x-2 mb-2 h-10">
          <Switch
            id="dangerous"
            checked={ownerData.isDangerous}
            onCheckedChange={(checked) => onChange("isDangerous", checked)}
            className="origin-left mr-2"
          />
          <label htmlFor="dangerous" className={`text-sm cursor-pointer ${C.text}`}>該当する</label>
        </div>
        <Label htmlFor="remarks" className={`text-sm ${C.text60}`}>備考・特記事項</Label>
        <Textarea
          id="remarks"
          rows={6}
          value={ownerData.remarks}
          maxLength={1000}
          onChange={(e) => onChange("remarks", e.target.value)}
          className={`text-sm ${C.text} min-h-[140px] resize-none ${C.borderMedium} p-3`}
        />
      </div>

      {/* Row 3 */}
      <div className="space-y-1.5">
        <Label htmlFor="ownerNameKana" className={`text-sm ${C.text60}`}>
          飼主名(カナ) <span className={C.textRequired}>*</span>
        </Label>
        <Input
          id="ownerNameKana"
          placeholder="ハヤシ フミアキ"
          value={ownerData.ownerNameKana}
          aria-invalid={!!fieldErrors.ownerNameKana}
          aria-describedby={fieldErrors.ownerNameKana ? "ownerNameKana-error" : undefined}
          onChange={(e) => { onChange("ownerNameKana", e.target.value); onClearError("ownerNameKana"); }}
          className={`${STYLE.formInput} ${fieldErrors.ownerNameKana ? STYLE.formInputError : ""}`}
        />
        <FormFieldError id="ownerNameKana-error" message={fieldErrors.ownerNameKana} />
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="address2" className={`text-sm ${C.text60}`}>住所2（会社）</Label>
        <Input id="address2" value={ownerData.address2} onChange={(e) => onChange("address2", e.target.value)} maxLength={200} className={STYLE.formInput} />
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="homeAddress1" className={`text-sm ${C.text60}`}>住所1(自宅)</Label>
        <Input id="homeAddress1" value={ownerData.homeAddress1} onChange={(e) => onChange("homeAddress1", e.target.value)} maxLength={200} className={STYLE.formInput} />
      </div>

      {/* Row 4 */}
      <div className="space-y-1.5">
        <Label htmlFor="birthDate" className={`text-sm ${C.text60}`}>飼主生年月日</Label>
        <NotionDatePicker id="birthDate" value={ownerData.birthDate} onChange={(val) => onChange("birthDate", val)} placeholder="生年月日を選択…" disabledDays={{ after: new Date() }} />
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="email" className={`text-sm ${C.text60}`}>メールアドレス</Label>
        <Input
          id="email"
          type="email"
          value={ownerData.email}
          aria-invalid={!!fieldErrors.email}
          aria-describedby={fieldErrors.email ? "email-error" : undefined}
          onChange={(e) => { onChange("email", e.target.value); onClearError("email"); }}
          className={`${STYLE.formInput} ${fieldErrors.email ? STYLE.formInputError : ""}`}
        />
        <FormFieldError id="email-error" message={fieldErrors.email} />
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="homeAddress2" className={`text-sm ${C.text60}`}>住所2(自宅)</Label>
        <Input id="homeAddress2" value={ownerData.homeAddress2} onChange={(e) => onChange("homeAddress2", e.target.value)} maxLength={200} className={STYLE.formInput} />
      </div>

      {/* Row 5 */}
      <div className="space-y-1.5">
        <Label htmlFor="phone" className={`text-sm ${C.text60}`}>
          電話番号 <span className={C.textRequired}>*</span>
        </Label>
        <Input
          id="phone"
          placeholder="090-1234-5678"
          value={ownerData.phone}
          aria-invalid={!!fieldErrors.phone}
          aria-describedby={fieldErrors.phone ? "phone-error" : undefined}
          onChange={(e) => { onChange("phone", e.target.value); onClearError("phone"); }}
          className={`${STYLE.formInput} ${fieldErrors.phone ? STYLE.formInputError : ""}`}
        />
        <FormFieldError id="phone-error" message={fieldErrors.phone} />
      </div>
      <div className="space-y-1.5 col-span-1 lg:col-span-2">
        <Label htmlFor="companyPhone" className={`text-sm ${C.text60}`}>会社 電話番号</Label>
        <Input id="companyPhone" value={ownerData.companyPhone} onChange={(e) => onChange("companyPhone", e.target.value)} className={STYLE.formInput} />
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="discountRate" className={`text-sm ${C.text60}`}>値引率 (%)</Label>
        <NumberInput
          id="discountRate"
          min={0}
          max={100}
          step={1}
          value={ownerData.discountRate || ""}
          aria-invalid={!!fieldErrors.discountRate}
          aria-describedby={fieldErrors.discountRate ? "discountRate-error" : undefined}
          onChange={(v) => { onChange("discountRate", Number(v)); onClearError("discountRate"); }}
          suffix="%"
          className={`${STYLE.formInput} ${fieldErrors.discountRate ? STYLE.formInputError : ""}`}
        />
        <FormFieldError id="discountRate-error" message={fieldErrors.discountRate} />
      </div>
    </div>
  );
});

export function OwnerForm({ petMutations }: { petMutations?: PetMutations } = {}) {
  const navigate = useNavigate();
  const { id: ownerId } = useParams();
  const { canEdit, canCreate, canDelete } = usePermission("owners");
  const { isDirty, markDirty, markClean } = useUnsavedChanges();
  const [deletePetTarget, setDeletePetTarget] = useState<{
    id: string;
    name: string;
  } | null>(null);

  const loaderData = useLoaderData() as OwnerLoaderData | undefined;
  const initialOwner = loaderData?.owner;

  const {
    isEdit,
    isLoading,
    ownerData,
    setOwnerData,
    pets,
    petModalOpen,
    setPetModalOpen,
    editingPet,
    handleAddPet,
    handleEditPet,
    handleDeletePet,
    handleSavePet,
    formAction,
    formState,
    fieldErrors,
    clearFieldError,
  } = useOwnerForm(ownerId, initialOwner, petMutations);

  const canSubmit = isEdit ? canEdit : canCreate;

  useTitle(isEdit ? `飼主編集 (${ownerData.ownerName})` : "飼主登録");

  // React 19 Action の成功を検知して遷移
  // BUG-065: 新規登録後は詳細ページへリダイレクト
  useEffect(() => {
    if (formState.success) {
      markClean();
      if (!isEdit && formState.data) {
        navigate(paths.owners.detail.getHref(formState.data as string));
      } else if (isEdit) {
        navigate(paths.owners.getHref());
      }
    }
  }, [formState.success, formState.data, formState.timestamp, navigate, markClean, isEdit]);

  // BUG-084: バリデーションエラー後に最初のエラーフィールドへフォーカスを移動する
  // フォームのアクセシビリティ改善（WCAG 2.4.3 Focus Order / 3.3.1 Error Identification）
  useEffect(() => {
    const errorFields = Object.keys(fieldErrors);
    if (errorFields.length === 0) return;
    // フィールド名とDOM IDのマッピング（OwnerInfoSection内の htmlFor と対応）
    const FIELD_ID_MAP: Record<string, string> = {
      ownerName: "ownerName",
      ownerNameKana: "ownerNameKana",
      phone: "phone",
      email: "email",
      discountRate: "discountRate",
    };
    // 優先度順にフォーカスする最初のフィールドを探す
    const PRIORITY_FIELDS = ["ownerName", "ownerNameKana", "phone", "email", "discountRate"];
    const firstErrorField = PRIORITY_FIELDS.find((f) => errorFields.includes(f)) ?? errorFields[0];
    const domId = FIELD_ID_MAP[firstErrorField] ?? firstErrorField;
    const el = document.getElementById(domId) as HTMLElement | null;
    el?.focus();
  // fieldErrors オブジェクトのキー変化で発火させるため JSON.stringify を使用
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(Object.keys(fieldErrors).sort())]);

  const handleBack = () => {
    navigate(paths.owners.getHref());
  };

  // FE-085: ペットの飼主変更ハンドラ
  const handlePetChangeOwner = useCallback(
    (newOwner: { id: string; name: string }) => {
      if (!editingPet?.id || !petMutations) return;
      petMutations.updatePetMutate(
        { id: editingPet.id, req: { owner_id: Number(newOwner.id) } },
        {
          onSuccess: () => {
            toast.success(`飼主を ${newOwner.name} に変更しました`);
            setPetModalOpen(false);
          },
          onError: () => toast.error("飼主変更に失敗しました"),
        },
      );
    },
    [editingPet, petMutations, setPetModalOpen],
  );

  // rerender-functional-setstate: setOwnerData・markDirty は両方安定した参照なので
  // useCallback で handleInputChange を安定化できる → MembershipTypeButtons memo の前提条件
  const handleInputChange = useCallback((field: string, value: string | boolean | number) => {
    setOwnerData(prev => ({ ...prev, [field]: value }));
    markDirty();
  }, [setOwnerData, markDirty]);

  const handleConfirmDeletePet = () => {
    if (deletePetTarget) {
      handleDeletePet(deletePetTarget.id);
      setDeletePetTarget(null);
    }
  };

  // rerender-memo: PetTableRow に渡すハンドラを安定化
  const handleDeletePetRequest = useCallback((id: string, name: string) => {
    setDeletePetTarget({ id, name });
  }, []);

  // MembershipTypeButtons の onChange prop を安定化（handleInputChange が stable なので依存なし）
  const handleMembershipChange = useCallback((type: MembershipType) => {
    handleInputChange("membershipType", type);
  }, [handleInputChange]);

  // 郵便番号検索
  const { lookup } = usePostalCodeLookup();
  const handlePostalCodeLookup = useCallback(
    async (postalCodeField: string, addressField: string) => {
      const postalCode = String(ownerData[postalCodeField as keyof OwnerData] ?? "");
      const result = await lookup(postalCode);
      if (result) {
        handleInputChange(addressField, `${result.prefecture}${result.city}${result.town}`);
      }
    },
    [ownerData, lookup, handleInputChange],
  );

  return (
    <form action={formAction}>
      <PageLayout
        title={isEdit ? "飼主・ペット　編集" : "飼主・ペット　登録"}
        onBack={handleBack}
        resource={ResourceOwners}
        maxWidth="max-w-[1400px]"
        headerAction={
          canSubmit ? (
            <SubmitButton size="sm">
              {isEdit ? "更新" : "登録"}
            </SubmitButton>
          ) : null
        }
      >
        <NavigationBlocker when={isDirty && !isLoading} />
        <fieldset disabled={!canSubmit} className="contents">

      {/* Owner Information Form */}
      {/* rerender-memo: OwnerInfoSection はペット操作・モーダル開閉では再レンダーしない */}
      <div className={`mb-4 rounded-lg bg-white p-4 border ${C.borderMedium}`}>
        <h2 className={`mb-3 text-sm font-bold ${C.text} flex items-center gap-2`}>
          <User className={`${ICON.action} ${C.text60}`} />
          飼主情報
        </h2>
        <OwnerInfoSection
          ownerData={ownerData}
          fieldErrors={fieldErrors}
          isEdit={isEdit}
          onChange={handleInputChange}
          onClearError={clearFieldError}
          onMembershipChange={handleMembershipChange}
          onPostalCodeLookup={handlePostalCodeLookup}
        />
      </div>

      {/* Pet Information */}
      <div className="mb-4 space-y-3">
        <div className="flex items-center justify-between">
          <h2 className={`text-sm font-bold ${C.text} flex items-center gap-2`}>
            <PawPrint className={`${ICON.action} ${C.text60}`} />
            ペット情報
          </h2>
          {canEdit ? (
            <Button
              size="sm"
              onClick={handleAddPet}
              className={`${STYLE.confirmPrimary} gap-1.5 text-sm px-4`}
            >
              <Plus className={ICON.action} />
              ペット追加
            </Button>
          ) : null}
        </div>

        <div className={`rounded-lg bg-white overflow-hidden border ${C.borderMedium}`}>
          <Table>
            {PET_TABLE_HEADER}
            <TableBody>
              {pets.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={11}
                    className={`text-center py-8 text-sm ${C.text60}`}
                  >
                    ペット情報がありません。「ペット追加」ボタンから追加してください。
                  </TableCell>
                </TableRow>
              ) : (
                // rerender-memo: PetTableRow で各行を memo 化 — N×8 インライン関数生成を排除
                pets.map((pet) => (
                  <PetTableRow
                    key={pet.id}
                    pet={pet}
                    ownerId={ownerId}
                    canEdit={canEdit}
                    canCreate={canCreate}
                    canDelete={canDelete}
                    onEdit={handleEditPet}
                    onDeleteRequest={handleDeletePetRequest}
                  />
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </div>

        </fieldset>
      <Suspense fallback={null}>
        <PetEditModal
          key={editingPet?.id ?? "new"}
          open={petModalOpen}
          onOpenChange={setPetModalOpen}
          ownerName={ownerData.ownerName || "新規飼主"}
          petData={editingPet ?? undefined}
          onSave={handleSavePet}
          onChangeOwner={handlePetChangeOwner}
        />
      </Suspense>

      {/* Delete Pet Confirm Dialog */}
      <ConfirmDialog
        open={!!deletePetTarget}
        onClose={() => setDeletePetTarget(null)}
        onConfirm={handleConfirmDeletePet}
        title="ペットを削除しますか？"
        description={`ペット「${deletePetTarget?.name}」を削除します。この操作は取り消すことができません。`}
        confirmLabel="削除"
        cancelLabel="キャンセル"
        variant="destructive"
      />
    </PageLayout>
    </form>
  );
}
