import {
  useState,
  lazy,
  Suspense,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
} from "react";
import { useNavigate, useParams, useLoaderData } from "react-router";
import { useQueryClient } from "@tanstack/react-query";
import { User, Receipt } from "lucide-react";
import { toast } from "sonner";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { useTitle } from "@/hooks/use-title";
import { usePostalCodeLookup } from "../hooks/use-postal-code-lookup";
import { useAuth } from "@/hooks/use-auth";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { handleApiError } from "@/lib/handle-api-error";
import { setStoredClinicId } from "@/lib/current-clinic";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";
import { OwnerInfoSection } from "../components/OwnerInfoSection";
import { OwnerPetsSection } from "../components/OwnerPetsSection";
import { useOwnerForm } from "../hooks/use-owner-form";
import { resolvePostCreateOwnerNavigation } from "../lib/post-create-owner-navigation";
import type { PetMutations } from "@/types/pet";
import type { OwnerData, MembershipTypeLabel } from "../types";
import type { OwnerLoaderData } from "../loaders";
import { ResourceOwners } from "@/types/generated/models";

// Lazy-loaded modal — only loaded when first opened (bundle-dynamic-imports)
const PetEditModal = lazy(() =>
  import("../components/PetEditModal").then(m => ({ default: m.PetEditModal }))
);

// rendering-hoist-jsx: アクセシビリティ用定数をモジュールレベルに巻き上げ（毎レンダー再生成を回避）
const OWNER_FIELD_ID_MAP: Record<string, string> = {
  ownerName: "ownerName",
  ownerNameKana: "ownerNameKana",
  phone: "phone",
  email: "email",
  discountRate: "discountRate",
  postalCode: "postalCode",
  homePostalCode: "homePostalCode",
};
// BUG-023: 形式/範囲エラーも優先フォーカス対象に含める
const OWNER_PRIORITY_FIELDS = [
  "ownerName",
  "ownerNameKana",
  "phone",
  "email",
  "postalCode",
  "homePostalCode",
  "discountRate",
] as const;

interface OwnerFormProps {
  petMutations?: PetMutations;
  lineSection?: React.ReactNode;
  accountingSection?: React.ReactNode;
}

export function OwnerForm({ petMutations, lineSection, accountingSection }: OwnerFormProps = {}) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { id: ownerId } = useParams();
  const { canEdit, canCreate, canDelete } = usePermission("owners");
  const canEditRef = useRef(canEdit);
  useLayoutEffect(() => {
    canEditRef.current = canEdit;
  }, [canEdit]);
  // #84: 登録先医院の選択肢（所属医院のみ）と現在の医院。複数所属時のみセレクト表示
  const { user, currentClinicId } = useAuth();
  // BUG-372: 割引権限（値引率フィールド制御用）
  const { canEdit: canEditDiscount } = usePermission("discount");
  // 会計履歴セクションは閲覧専用なので accounting:view で出し分ける。
  // 権限がないユーザーにはセクション全体（見出し含む）を非表示にする。
  const { canView: canViewAccounting } = usePermission("accounting");
  const { isDirty, markDirty, markClean } = useUnsavedChanges();
  const [deletePetTarget, setDeletePetTarget] = useState<{
    id: string;
    name: string;
  } | null>(null);
  const [pendingOwnerChange, setPendingOwnerChange] = useState<{
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
    handlePetLifecycleChange,
    formAction,
    formState,
    fieldErrors,
    clearFieldError,
  } = useOwnerForm(ownerId, initialOwner, petMutations, {
    canCreate,
    canEdit,
    canDelete,
  });
  const editingPetRef = useRef(editingPet);
  useLayoutEffect(() => {
    editingPetRef.current = editingPet;
  }, [editingPet]);

  const canSubmit = isEdit ? canEdit : canCreate;

  useTitle(isEdit ? `飼主編集 (${ownerData.ownerName})` : "飼主登録");

  // React 19 Action の成功を検知して遷移
  // BUG-065: 新規登録後は詳細ページへリダイレクト
  // BUG-010: 登録先医院 ≠ グローバル選択なら clinic を切替えて hard navigate（X-Clinic-ID 整合）
  useEffect(() => {
    if (formState.success) {
      markClean();
      if (!isEdit && formState.data) {
        const payload = formState.data as { id: string; clinicId?: string } | string;
        const createdOwnerId = typeof payload === "string" ? payload : payload.id;
        const targetClinicId =
          typeof payload === "string"
            ? ownerData.clinicId
            : (payload.clinicId ?? ownerData.clinicId);
        const plan = resolvePostCreateOwnerNavigation({
          ownerId: createdOwnerId,
          targetClinicId,
          currentClinicId,
        });
        if (plan.mode === "hard") {
          if (!setStoredClinicId(plan.clinicId)) {
            toast.error("クリニックの切替に失敗しました。登録は完了しています。医院を切り替えてから詳細を開いてください。");
            navigate(paths.owners.getHref());
            return;
          }
          // switchClinic と同様: 旧 clinic キャッシュを捨ててから新 X-Clinic-ID で詳細をロード
          queryClient.clear();
          window.location.assign(plan.href);
          return;
        }
        navigate(plan.href);
      } else if (isEdit) {
        navigate(paths.owners.getHref());
      }
    }
  }, [
    formState.success,
    formState.data,
    formState.timestamp,
    navigate,
    markClean,
    isEdit,
    ownerData.clinicId,
    currentClinicId,
    queryClient,
  ]);

  // fieldErrors の「キーの集合」が変わったときだけ発火させたい（値の再代入では発火不要）。
  // オブジェクト参照そのものを deps に使うと毎レンダーで新規オブジェクトになり無限発火するため、
  // キー集合の署名文字列を useMemo 化して deps にする（FE-RC-066）。
  const errorFieldsSignature = useMemo(
    () => Object.keys(fieldErrors).sort().join(","),
    [fieldErrors],
  );

  // BUG-084: バリデーションエラー後に最初のエラーフィールドへフォーカスを移動する
  // フォームのアクセシビリティ改善（WCAG 2.4.3 Focus Order / 3.3.1 Error Identification）
  useEffect(() => {
    const errorFields = errorFieldsSignature ? errorFieldsSignature.split(",") : [];
    if (errorFields.length === 0) return;
    // 優先度順にフォーカスする最初のフィールドを探す
    const firstErrorField = OWNER_PRIORITY_FIELDS.find((f) => errorFields.includes(f)) ?? errorFields[0];
    const domId = OWNER_FIELD_ID_MAP[firstErrorField] ?? firstErrorField;
    const el = document.getElementById(domId) as HTMLElement | null;
    el?.focus();
  }, [errorFieldsSignature]);

  const handleBack = () => {
    navigate(paths.owners.getHref());
  };

  // BUG-373: 飼主変更 — discount_rate/membership_type が異なる時のみ確認モーダル
  const handlePetChangeOwner = useCallback(
    (newOwner: { id: string; name: string; discountRate: number; membershipType: string }) => {
      const currentEditingPet = editingPetRef.current;
      if (
        canEditRef.current !== true ||
        !currentEditingPet?.id ||
        currentEditingPet.status === "死亡" ||
        !petMutations
      ) {
        return;
      }
      const needsConfirm =
        (ownerData.discountRate ?? 0) !== newOwner.discountRate ||
        ownerData.membershipType !== newOwner.membershipType;
      if (needsConfirm) {
        setPendingOwnerChange({ id: newOwner.id, name: newOwner.name });
      } else {
        if (canEditRef.current !== true) return;
        petMutations.updatePetMutate(
          { id: currentEditingPet.id, req: { owner_id: Number(newOwner.id) } },
          {
            onSuccess: () => {
              toast.success(`飼主を ${newOwner.name} に変更しました`);
              setPetModalOpen(false);
            },
            onError: (error) => {
              handleApiError(error, "飼主変更");
            },
          },
        );
      }
    },
    [petMutations, ownerData.discountRate, ownerData.membershipType, setPetModalOpen],
  );

  const handleConfirmOwnerChange = useCallback(() => {
    const currentEditingPet = editingPetRef.current;
    if (
      canEditRef.current !== true ||
      !pendingOwnerChange ||
      !currentEditingPet?.id ||
      currentEditingPet.status === "死亡" ||
      !petMutations
    ) {
      return;
    }
    const newOwner = pendingOwnerChange;
    if (canEditRef.current !== true) return;
    petMutations.updatePetMutate(
      { id: currentEditingPet.id, req: { owner_id: Number(newOwner.id) } },
      {
        onSuccess: () => {
          toast.success(`飼主を ${newOwner.name} に変更しました`);
          setPendingOwnerChange(null);
          setPetModalOpen(false);
        },
        onError: (error) => {
          handleApiError(error, "飼主変更");
          setPendingOwnerChange(null);
        },
      },
    );
  }, [pendingOwnerChange, petMutations, setPetModalOpen]);

  // rerender-functional-setstate: setOwnerData・markDirty は両方安定した参照なので
  // useCallback で handleInputChange を安定化できる → MembershipTypeButtons memo の前提条件
  const handleInputChange = useCallback((field: string, value: string | boolean | number | null | undefined) => {
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
  const handleMembershipChange = useCallback((type: MembershipTypeLabel) => {
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

  // BUG-023: HTML5 の type=email / type=number max が formAction 前に静かにブロックするため noValidate。
  // 表示は useOwnerForm の fieldErrors に一本化。
  return (
    <form action={formAction} noValidate className="h-full">
      <PageLayout
        title={isEdit ? "飼主・ペット　編集" : "飼主・ペット　登録"}
        onBack={handleBack}
        resource={ResourceOwners}
        maxWidth={LAYOUT.pageContentMaxWidth.full}
        headerAction={
          canSubmit ? (
            <SubmitButton size="sm" colorVariant="primary">
              {isEdit ? "更新" : "登録"}
            </SubmitButton>
          ) : null
        }
      >
        <NavigationBlocker when={isDirty ? !isLoading : false} />
        <fieldset disabled={!canSubmit} className="border-0 p-0 m-0 min-w-0">
          <div className={`mb-4 rounded-lg ${C.bgWhite} p-4 border ${C.borderMedium}`}>
            <h2 className={`mb-3 text-sm font-bold ${C.text} flex items-center gap-2`}>
              <User className={`${ICON.action} ${C.text60}`} />
              飼主情報
            </h2>
            <OwnerInfoSection
              ownerData={ownerData}
              fieldErrors={fieldErrors}
              isEdit={isEdit}
              canEditDiscount={canEditDiscount}
              clinicOptions={user?.clinics}
              currentClinicId={currentClinicId}
              onChange={handleInputChange}
              onClearError={clearFieldError}
              onMembershipChange={handleMembershipChange}
              onPostalCodeLookup={handlePostalCodeLookup}
            />
          </div>

          <OwnerPetsSection
            pets={pets}
            ownerId={ownerId}
            canEdit={canEdit}
            canCreate={canCreate}
            canDelete={canDelete}
            onAddPet={handleAddPet}
            onEditPet={handleEditPet}
            onDeleteRequest={handleDeletePetRequest}
          />
        </fieldset>

        {/* 会計履歴セクション（編集モード + accounting:view 権限保有 + app層からの注入時のみ表示） */}
        {isEdit && ownerId && canViewAccounting && accountingSection ? (
          <div className="mb-4 space-y-3">
            <div className="flex items-center justify-between">
              <h2 className={`text-sm font-bold ${C.text} flex items-center gap-2`}>
                <Receipt className={`${ICON.action} ${C.text60}`} />
                会計履歴
              </h2>
            </div>
            <Suspense fallback={null}>
              {accountingSection}
            </Suspense>
          </div>
        ) : null}

        {/* LINE連携セクション（編集モードのみ・app層から注入） */}
        {isEdit && lineSection ? (
          <div className={`mt-6 p-4 rounded-lg border ${C.borderLight}`}>
            {lineSection}
          </div>
        ) : null}

        <Suspense fallback={null}>
          <PetEditModal
            key={editingPet?.id ?? "new"}
            open={petModalOpen}
            onOpenChange={setPetModalOpen}
            ownerName={ownerData.ownerName || "新規飼主"}
            petData={editingPet ?? undefined}
            onSave={handleSavePet}
            onChangeOwner={handlePetChangeOwner}
            onPetLifecycleChange={handlePetLifecycleChange}
          />
        </Suspense>

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

        <ConfirmDialog
          open={!!pendingOwnerChange}
          onClose={() => setPendingOwnerChange(null)}
          onConfirm={handleConfirmOwnerChange}
          title="飼主変更の確認"
          description="飼主によって値引率や会員区分が異なるため、今後の会計金額が変動する可能性があります。変更を続行してよろしいですか?"
          confirmLabel="続行"
          cancelLabel="キャンセル"
        />
      </PageLayout>
    </form>
  );
}
