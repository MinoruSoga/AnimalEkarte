import {
  memo,
  useActionState,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { UserRound } from "lucide-react";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { MasterSidePanel } from "@/components/shared/SidePeek";
import { LAYOUT } from "@/lib/design-tokens";
import {
  useGetStaffCapableReservationTypes,
  useGetStaffClinics,
  useGetStaffPermissionGroups,
  type ClinicSummary,
  type Staff,
} from "../api/staffs";
import type { Occupation } from "../api/occupations";
import type { PermissionGroup } from "../api/permission-groups";
import type { ReservationType } from "../api/reservation-types";
import { StaffBasicInfoSection } from "./StaffBasicInfoSection";
import { StaffLineReservationSection } from "./StaffLineReservationSection";
import {
  StaffClinicsSection,
  StaffExcludedReservationTypesSection,
  StaffPermissionGroupsSection,
} from "./StaffSidePanelSections";
import { staffToFormData, type StaffFormData } from "../lib/staff-side-panel-model";
import { useEditableIdSelection } from "../hooks/use-editable-id-selection";

interface StaffSidePanelProps {
  item: Staff | null;
  onClose: () => void;
  onSave: (data: StaffFormData) => Promise<boolean> | boolean;
  onSaveWithResult?: (data: StaffFormData, targetId?: string) => Promise<Staff | null>;
  canEditPersisted?: boolean;
  onDeleteRequest?: (item: Staff) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
  allOccupations: Occupation[];
  allGroups: PermissionGroup[];
  onSaveGroups: (staffId: string, groupIds: string[]) => Promise<boolean>;
  allClinics: ClinicSummary[];
  onSaveClinics: (staffId: string, clinicIds: string[]) => Promise<boolean>;
  allReservationTypes: ReservationType[];
  onSaveCapableReservationTypes: (
    staffId: string,
    reservationTypeIds: string[],
  ) => Promise<boolean>;
}

/**
 * 作成成功の ID をこのマウント内で採用し、ドラフト全体を再初期化したり React key を切り替えたりしない。
 * 送信時と一致するパスワードだけは採用時にクリアする。
 * 採用後の更新可否は、新規モードの作成権限ではなく、その時点の最新の編集権限で判定する。
 */
export const StaffSidePanel = memo(function StaffSidePanel({
  item,
  onClose,
  onSave,
  onSaveWithResult,
  canEditPersisted = true,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
  allOccupations,
  allGroups,
  onSaveGroups,
  allClinics,
  onSaveClinics,
  allReservationTypes,
  onSaveCapableReservationTypes,
}: StaffSidePanelProps) {
  const [committedStaff, setCommittedStaff] = useState<Staff | null>(null);
  const effectiveItem = item ?? committedStaff;
  const isNew = effectiveItem === null;
  const staffId = effectiveItem?.id ?? null;
  const effectiveReadOnly = committedStaff !== null ? canEditPersisted !== true : readOnly === true;

  const [formData, setFormData] = useState<StaffFormData>(() => staffToFormData(item));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");
  const mountedRef = useRef(true);
  const savingRef = useRef(false);
  const editVersionRef = useRef(0);

  useLayoutEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  useEffect(() => {
    onDirtyChange?.(isDirty);
  }, [isDirty, onDirtyChange]);

  /**
   * 編集バージョンを進め、親の未保存破棄ガードを次の描画を待たずに同期で有効化する。
   */
  const markDirty = useCallback(() => {
    editVersionRef.current += 1;
    setIsDirty(true);
    onDirtyChange?.(true);
  }, [onDirtyChange]);
  const setFormDataDirty = useCallback<typeof setFormData>(
    (updater) => {
      setFormData(updater);
      markDirty();
    },
    [markDirty],
  );

  const activeReservationTypes = useMemo(
    () => allReservationTypes.filter((reservationType) => reservationType.isActive),
    [allReservationTypes],
  );

  const { data: serverGroupIds } = useGetStaffPermissionGroups(staffId);
  const {
    ids: groupIds,
    idSet: groupIdSet,
    handleToggle: handleGroupToggle,
  } = useEditableIdSelection({ serverIds: serverGroupIds, markDirty });

  const { data: serverClinicIds } = useGetStaffClinics(staffId);
  const {
    ids: clinicIds,
    idSet: clinicIdSet,
    handleToggle: handleClinicToggle,
  } = useEditableIdSelection({ serverIds: serverClinicIds, markDirty });

  const { data: serverCapableIds } = useGetStaffCapableReservationTypes(staffId);
  const {
    ids: capableIds,
    idSet: capableIdSet,
    handleToggle: handleCapabilityToggle,
  } = useEditableIdSelection({ serverIds: serverCapableIds, markDirty });

  /**
   * ID の順序と重複を無視して集合を比較し、関連先が同じなら不要な更新を避ける。
   */
  const setsEqual = useCallback((a: string[], b: string[]) => {
    const setA = new Set(a);
    const setB = new Set(b);
    if (setA.size !== setB.size) return false;
    for (const id of setA) {
      if (!setB.has(id)) return false;
    }
    return true;
  }, []);

  /**
   * 基本保存を await してから、変更された関連付けのみを順次書き込む。
   * 部分失敗時は dirty 状態を保持し、完了済みの書き込みはロールバックしない。
   * 観測された新規作成エンティティは新しいドラフトがあっても採用するが、
   * dirty 解除とクローズは同じ編集バージョンかつマウント済みの場合のみ行う。
   * 未観測のネットワーク結果に対する冪等性は保証しない。
   */
  const runSaveSequence = useCallback(async () => {
    if (!formData.name.trim()) {
      setNameError("氏名を入力してください");
      return;
    }
    setNameError("");
    const startVersion = editVersionRef.current;
    const snapshot = {
      isNew,
      staffId,
      groupIds,
      clinicIds,
      capableIds,
      formPassword: formData.password,
      groupsChanged: serverGroupIds !== undefined && !setsEqual(groupIds, serverGroupIds),
      clinicsChanged: serverClinicIds !== undefined && !setsEqual(clinicIds, serverClinicIds),
      capableChanged: serverCapableIds !== undefined && !setsEqual(capableIds, serverCapableIds),
    };
    try {
      const savedEntity = onSaveWithResult
        ? await onSaveWithResult(formData, committedStaff?.id)
        : await onSave(formData);
      if (!savedEntity) return;
      if (onSaveWithResult && snapshot.isNew && savedEntity !== true && mountedRef.current) {
        setCommittedStaff(savedEntity);
        setFormData((prev) =>
          prev.password === snapshot.formPassword ? { ...prev, password: "" } : prev,
        );
      }
      if (!snapshot.isNew && snapshot.staffId) {
        const steps: Array<[boolean, (id: string, ids: string[]) => Promise<boolean>, string[]]> = [
          [snapshot.groupsChanged, onSaveGroups, snapshot.groupIds],
          [snapshot.clinicsChanged, onSaveClinics, snapshot.clinicIds],
          [snapshot.capableChanged, onSaveCapableReservationTypes, snapshot.capableIds],
        ];
        for (const [changed, saveAssoc, ids] of steps) {
          if (!changed) continue;
          if (!mountedRef.current || editVersionRef.current !== startVersion) return;
          const ok = await saveAssoc(snapshot.staffId, ids);
          if (!ok) return;
        }
      }
      if (!mountedRef.current) return;
      if (editVersionRef.current !== startVersion) return;
      onDirtyChange?.(false);
      setIsDirty(false);
      toast.success(snapshot.isNew ? "登録しました" : "更新しました");
      onClose();
    } catch (error) {
      if (mountedRef.current) {
        handleApiError(error, "保存");
      }
    }
  }, [
    formData,
    isNew,
    staffId,
    groupIds,
    clinicIds,
    capableIds,
    serverGroupIds,
    serverClinicIds,
    serverCapableIds,
    setsEqual,
    onSave,
    onSaveWithResult,
    committedStaff,
    onSaveGroups,
    onSaveClinics,
    onSaveCapableReservationTypes,
    onDirtyChange,
    onClose,
  ]);

  /**
   * pending 表示は React 19 のネイティブ form action に委ねる。
   * 同時実行の再入はローカル inflight ガードで拒否する。
   */
  const [, formAction, isPending] = useActionState(async (_previous: null, _formData: FormData) => {
    if (savingRef.current) return null;
    savingRef.current = true;
    try {
      await runSaveSequence();
    } finally {
      if (mountedRef.current) {
        savingRef.current = false;
      }
    }
    return null;
  }, null);

  const handleTitleChange = useCallback(
    (value: string) => {
      setFormDataDirty((prev) => ({ ...prev, name: value }));
      if (value.trim()) setNameError("");
    },
    [setFormDataDirty],
  );

  /**
   * ここで dirty をクリアしない。閉じる/破棄確認は親の既存ガードに委譲する。
   */
  const handleClose = useCallback(() => {
    onClose();
  }, [onClose]);

  return (
    <MasterSidePanel
      tabletOverlay
      isNew={isNew}
      title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose}
      action={formAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<UserRound className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      isPending={isPending}
      titleError={nameError}
      titleMaxLength={100}
      titleDescription="スタッフの氏名です。カルテ・予約・シフトなど各画面で担当者名として表示されます。"
      // EMR-386: 新規作成時は共有既定の「無題」ではなく用途が分かる仮タイトルを渡す。
      // 既存スタッフは未指定（undefined）のまま共有既定を維持する。
      titlePlaceholder={isNew ? "新規スタッフ登録" : undefined}
      readOnly={effectiveReadOnly}
    >
      <StaffBasicInfoSection
        item={effectiveItem}
        isNew={isNew}
        disableCreateCredentials={isNew === true && isPending === true ? true : undefined}
        formData={formData}
        setFormDataDirty={setFormDataDirty}
        allOccupations={allOccupations}
      />

      <StaffLineReservationSection formData={formData} setFormDataDirty={setFormDataDirty} />

      <fieldset
        className="min-w-0 border-0 p-0 m-0"
        disabled={effectiveReadOnly || (!isNew && serverCapableIds === undefined)}
      >
        <StaffExcludedReservationTypesSection
          activeReservationTypes={activeReservationTypes}
          allReservationTypes={allReservationTypes}
          capableIdSet={capableIdSet}
          isNew={isNew}
          onToggle={handleCapabilityToggle}
        />
      </fieldset>

      <fieldset
        className="min-w-0 border-0 p-0 m-0"
        disabled={effectiveReadOnly || (!isNew && serverClinicIds === undefined)}
      >
        <StaffClinicsSection
          allClinics={allClinics}
          clinicIdSet={clinicIdSet}
          isNew={isNew}
          onToggle={handleClinicToggle}
        />
      </fieldset>

      <fieldset
        className="min-w-0 border-0 p-0 m-0"
        disabled={effectiveReadOnly || (!isNew && serverGroupIds === undefined)}
      >
        <StaffPermissionGroupsSection
          allGroups={allGroups}
          groupIdSet={groupIdSet}
          isNew={isNew}
          onToggle={handleGroupToggle}
        />
      </fieldset>
    </MasterSidePanel>
  );
});
