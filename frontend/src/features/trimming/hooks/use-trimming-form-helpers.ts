import { useState, useEffect, useCallback, useRef } from "react";
import { useNavigate } from "react-router";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { paths } from "@/config/paths";
import type { TrimmingFormData } from "@/types/trimming";
import type { Pet } from "@/types";
import { uploadTrimmingImage } from "../api/upload-trimming-image";
import { deleteTrimmingImage } from "../api/delete-trimming-image";

type TrimmingImageField = "styleImage" | "completedImage";

type TrimmingImageOverrides = (
  updater: (prev: Partial<TrimmingFormData>) => Partial<TrimmingFormData>,
) => void;

/** 未永続化のまま残ったアップロード済みキーを捨てる best-effort 掃除。失敗は握り潰す。 */
function discardUploadedKey(key: string): void {
  deleteTrimmingImage(key).catch(() => {
    // best-effort GC: 削除失敗時は orphan が残るだけなので通知しない
  });
}

function selectedImageFile(event: React.ChangeEvent<HTMLInputElement>): File | null {
  const file = event.target.files?.[0] ?? null;
  event.target.value = "";
  return file;
}

function useTrimmingImageField(
  field: TrimmingImageField,
  setLocalOverrides: TrimmingImageOverrides,
  markDirty: () => void,
) {
  const requestIdRef = useRef(0);
  const [preview, setPreview] = useState<string | null>(null);
  const [pendingCount, setPendingCount] = useState(0);
  // 即時プレビュー用 object URL と未保存アップロードキーを追跡する。
  const blobUrlsRef = useRef<Set<string>>(new Set());
  const sessionKeyRef = useRef<string | null>(null);
  // フォーム値（最後に確定した key / 既存レコード値）に対応するプレビュー。
  // アップロード失敗時は常にここへ戻す — 置き換え途中の blob に依存しない。
  const committedPreviewRef = useRef<string | null>(null);

  const releaseBlob = useCallback((blobUrl: string) => {
    if (blobUrlsRef.current.delete(blobUrl)) {
      URL.revokeObjectURL(blobUrl);
    }
  }, []);

  const releaseAllBlobs = useCallback(() => {
    for (const url of blobUrlsRef.current) {
      URL.revokeObjectURL(url);
    }
    blobUrlsRef.current.clear();
  }, []);

  // アンマウント時: in-flight を無効化し、blob URL と未保存アップロードを解放する。
  // 送信済みキー（markSubmitted 済み）はレコード参照の可能性があるため捨てない。
  useEffect(
    () => () => {
      requestIdRef.current += 1;
      releaseAllBlobs();
      if (sessionKeyRef.current) {
        discardUploadedKey(sessionKeyRef.current);
        sessionKeyRef.current = null;
      }
    },
    [releaseAllBlobs],
  );

  const handleChange = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const file = selectedImageFile(event);
      if (!file) return undefined;
      const requestId = requestIdRef.current + 1;
      requestIdRef.current = requestId;
      const blobUrl = URL.createObjectURL(file);
      blobUrlsRef.current.add(blobUrl);
      setPreview(blobUrl); // アップロード完了を待たず即時プレビュー
      setPendingCount((count) => count + 1);
      markDirty(); // 画像選択自体が未保存変更（転送中の離脱も警告対象）
      const isCurrent = () => requestIdRef.current === requestId;
      return uploadTrimmingImage(file)
        .then((uploaded) => {
          if (!isCurrent()) {
            // stale: このキーは sessionKeyRef に入ったことがない（=未送信確定）
            // ので安全に捨てられる。blob も表示中でないため即解放。
            releaseBlob(blobUrl);
            discardUploadedKey(uploaded.key);
            return;
          }
          const replacedSessionKey = sessionKeyRef.current;
          sessionKeyRef.current = uploaded.key;
          setLocalOverrides((prev) => ({ ...prev, [field]: uploaded.key }));
          const previousCommittedPreview = committedPreviewRef.current;
          if (uploaded.signed_url !== "") {
            committedPreviewRef.current = uploaded.signed_url;
            setPreview(uploaded.signed_url);
            releaseBlob(blobUrl); // 表示が署名URLへ切替わった blob は即解放
          } else {
            committedPreviewRef.current = blobUrl; // blob が確定プレビューとして残る
          }
          // 置き換えで不要になった旧確定 blob（signed_url 空の前回値）を解放
          if (
            previousCommittedPreview !== null &&
            previousCommittedPreview !== committedPreviewRef.current &&
            blobUrlsRef.current.has(previousCommittedPreview)
          ) {
            releaseBlob(previousCommittedPreview);
          }
          // 未保存で置き換わった先行アップロードは到達不能 → 掃除
          if (replacedSessionKey && replacedSessionKey !== uploaded.key) {
            discardUploadedKey(replacedSessionKey);
          }
        })
        .catch((error) => {
          releaseBlob(blobUrl);
          if (!isCurrent()) return;
          setPreview(committedPreviewRef.current); // 直前の確定 key/プレビューを残す
          handleApiError(error, "画像アップロード");
        })
        .finally(() => {
          setPendingCount((count) => count - 1);
        });
    },
    [field, setLocalOverrides, markDirty, releaseBlob],
  );

  const remove = useCallback(() => {
    requestIdRef.current += 1;
    setLocalOverrides((prev) => ({ ...prev, [field]: "" }));
    markDirty();
    setPreview(null);
    committedPreviewRef.current = null;
    const sessionKey = sessionKeyRef.current;
    sessionKeyRef.current = null;
    if (sessionKey) {
      discardUploadedKey(sessionKey); // 未送信セッションキーのみ
    }
    releaseAllBlobs();
  }, [field, setLocalOverrides, markDirty, releaseAllBlobs]);

  // 保存リクエスト送信時に呼ぶ。ペイロードへ含めたキーは送信以後サーバー側の
  // 所有物として扱い、手放したキー（session 管理だったもの。無ければ null）を返す。
  // 応答が確定 pre-commit 失敗（4xx = サーバーが commit していない）なら
  // caller は restoreSubmitted で session 所有へ戻す。応答無し/5xx は commit
  // 状態不明のため戻さない — 参照中オブジェクトを消すより orphan の方が安全。
  const markSubmitted = useCallback((): string | null => {
    const key = sessionKeyRef.current;
    sessionKeyRef.current = null;
    return key;
  }, []);

  const restoreSubmitted = useCallback((key: string | null) => {
    // 別キーへ既に所有権が動いていたら上書きしない（起こり得ないが安全側）。
    if (key === null || sessionKeyRef.current !== null) return;
    sessionKeyRef.current = key;
  }, []);

  // hydration 等で確定済みプレビューを外部から設定する（record の署名URL）。
  const setCommittedPreview = useCallback((value: string | null) => {
    committedPreviewRef.current = value;
    setPreview(value);
  }, []);

  return {
    preview,
    setPreview: setCommittedPreview,
    pendingCount,
    handleChange,
    remove,
    markSubmitted,
    restoreSubmitted,
  };
}

type TrimmingHydrateSource = {
  styleRequest?: string;
  courseId?: string | null;
  optionIds?: string[] | null;
  bw?: string | null;
  bwUnit?: string | null;
  bt?: string | null;
  usedShampoo?: string | null;
  usedRibbon?: string | null;
  remarks?: string | null;
  staffId?: string | null;
  staff?: string | null;
  styleImage?: string | null;
  completedImage?: string | null;
  reservationTypeId?: string;
};

function toBwUnit(value: string | null | undefined): TrimmingFormData["bwUnit"] {
  return value === "g" ? "g" : "Kg";
}

function existingTrimmingOverrides(existing: TrimmingHydrateSource): Partial<TrimmingFormData> {
  return {
    styleRequest: existing.styleRequest,
    courseId: existing.courseId ?? "",
    optionIds: existing.optionIds ?? [],
    bw: existing.bw ?? "",
    bwUnit: toBwUnit(existing.bwUnit),
    bt: existing.bt ?? "",
    usedShampoo: existing.usedShampoo ?? "",
    usedRibbon: existing.usedRibbon ?? "",
    remarks: existing.remarks ?? "",
    staffId: existing.staffId ?? "",
    staffName: existing.staff ?? "",
  };
}

function appointmentTrimmingOverrides(
  prev: Partial<TrimmingFormData>,
  existing: TrimmingHydrateSource,
): Partial<TrimmingFormData> {
  return {
    ...prev,
    reservationTypeId: existing.reservationTypeId || prev.reservationTypeId || "",
    styleRequest: existing.styleRequest || prev.styleRequest || "",
    courseId: existing.courseId || prev.courseId || "",
    optionIds:
      (existing.optionIds?.length ?? 0) > 0 ? (existing.optionIds ?? []) : (prev.optionIds ?? []),
    bw: existing.bw || prev.bw || "",
    bwUnit: toBwUnit(existing.bwUnit || prev.bwUnit),
    bt: existing.bt || prev.bt || "",
    usedShampoo: existing.usedShampoo || prev.usedShampoo || "",
    usedRibbon: existing.usedRibbon || prev.usedRibbon || "",
    remarks: existing.remarks || prev.remarks || "",
    staffId: existing.staffId || prev.staffId || "",
    staffName: existing.staff || prev.staffName || "",
  };
}

export function useTrimmingFormHydration(input: {
  isEdit: boolean;
  existingTrimming: TrimmingHydrateSource | undefined;
  existingAppointmentTrimming: TrimmingHydrateSource | undefined;
  setLocalOverrides: (
    next:
      Partial<TrimmingFormData> | ((prev: Partial<TrimmingFormData>) => Partial<TrimmingFormData>),
  ) => void;
  setStyleImagePreview: (value: string | null) => void;
  setCompletedImagePreview: (value: string | null) => void;
}) {
  const {
    isEdit,
    existingTrimming,
    existingAppointmentTrimming,
    setLocalOverrides,
    setStyleImagePreview,
    setCompletedImagePreview,
  } = input;
  const serverDataLoadedRef = useRef(false);
  useEffect(() => {
    if (isEdit && existingTrimming && !serverDataLoadedRef.current) {
      serverDataLoadedRef.current = true;
      setLocalOverrides(existingTrimmingOverrides(existingTrimming));
      if (existingTrimming.styleImage) {
        setStyleImagePreview(existingTrimming.styleImage);
      }
      if (existingTrimming.completedImage) {
        setCompletedImagePreview(existingTrimming.completedImage);
      }
    }
  }, [isEdit, existingTrimming, setLocalOverrides, setStyleImagePreview, setCompletedImagePreview]);

  const appointmentDataLoadedRef = useRef(false);
  useEffect(() => {
    if (isEdit || !existingAppointmentTrimming || appointmentDataLoadedRef.current) return;
    appointmentDataLoadedRef.current = true;
    setLocalOverrides((prev) => appointmentTrimmingOverrides(prev, existingAppointmentTrimming));
    if (existingAppointmentTrimming.styleImage) {
      setStyleImagePreview(existingAppointmentTrimming.styleImage);
    }
    if (existingAppointmentTrimming.completedImage) {
      setCompletedImagePreview(existingAppointmentTrimming.completedImage);
    }
  }, [
    isEdit,
    existingAppointmentTrimming,
    setLocalOverrides,
    setStyleImagePreview,
    setCompletedImagePreview,
  ]);
}

// markSubmittedKeys が返す送信時点の session 所有キー。
// restoreSubmittedKeys へそのまま渡して所有権を巻き戻す。
export type SubmittedTrimmingImageKeys = {
  styleImage: string | null;
  completedImage: string | null;
};

export function useTrimmingFormImages(
  setLocalOverrides: TrimmingImageOverrides,
  markDirty: () => void,
) {
  const styleImage = useTrimmingImageField("styleImage", setLocalOverrides, markDirty);
  const completedImage = useTrimmingImageField("completedImage", setLocalOverrides, markDirty);
  const markStyleSubmitted = styleImage.markSubmitted;
  const markCompletedSubmitted = completedImage.markSubmitted;
  const restoreStyleSubmitted = styleImage.restoreSubmitted;
  const restoreCompletedSubmitted = completedImage.restoreSubmitted;

  const markSubmittedKeys = useCallback((): SubmittedTrimmingImageKeys => {
    return {
      styleImage: markStyleSubmitted(),
      completedImage: markCompletedSubmitted(),
    };
  }, [markStyleSubmitted, markCompletedSubmitted]);

  // 確定 pre-commit 失敗（4xx 応答）のときだけ呼ぶ — 退避したキーを session
  // 所有へ戻し、remove/離脱クリーンアップで削除可能にする。
  const restoreSubmittedKeys = useCallback(
    (keys: SubmittedTrimmingImageKeys) => {
      restoreStyleSubmitted(keys.styleImage);
      restoreCompletedSubmitted(keys.completedImage);
    },
    [restoreStyleSubmitted, restoreCompletedSubmitted],
  );

  return {
    styleImagePreview: styleImage.preview,
    setStyleImagePreview: styleImage.setPreview,
    completedImagePreview: completedImage.preview,
    setCompletedImagePreview: completedImage.setPreview,
    handleStyleImageChange: styleImage.handleChange,
    handleCompletedImageChange: completedImage.handleChange,
    removeStyleImage: styleImage.remove,
    removeCompletedImage: completedImage.remove,
    isImageUploading: styleImage.pendingCount > 0 || completedImage.pendingCount > 0,
    markSubmittedKeys,
    restoreSubmittedKeys,
  };
}

export function useTrimmingFormPetSync(input: {
  isEdit: boolean;
  petId: string | null;
  petFromQuery: Pet | undefined;
  petFromEdit?: Pet;
  isPetLoading: boolean;
  setSelectedPets: (pets: Pet[]) => void;
}) {
  const { isEdit, petId, petFromQuery, petFromEdit, isPetLoading, setSelectedPets } = input;
  const navigate = useNavigate();

  useEffect(() => {
    if (isEdit && petFromEdit) {
      setSelectedPets([petFromEdit]);
    }
  }, [isEdit, petFromEdit, setSelectedPets]);

  useEffect(() => {
    if (!isEdit) {
      if (petFromQuery) {
        setSelectedPets([petFromQuery]);
      } else if (!petId && !isPetLoading) {
        navigate(paths.trimming.selectPet.getHref());
      }
    }
  }, [isEdit, petId, petFromQuery, isPetLoading, setSelectedPets, navigate]);
}

export function createTrimmingDeleteHandler(input: {
  isEdit: boolean;
  id: string | undefined;
  isMutationAllowed: (action: "canCreate" | "canEdit" | "canDelete") => boolean;
  isEditPetReady: () => boolean;
  isPetDeceased: () => boolean;
  deleteTrimming: (
    id: string,
    opts: { onSuccess: () => void; onError: (error: unknown) => void },
  ) => void;
}): (onSuccess?: () => void) => void {
  return (onSuccess?: () => void) => {
    if (!input.isEdit || !input.id) return;
    if (!input.isMutationAllowed("canDelete")) {
      toast.error("この操作を行う権限がありません");
      return;
    }
    if (!input.isEditPetReady()) {
      toast.error("ペット情報の読み込みが完了してから削除してください");
      return;
    }
    if (input.isPetDeceased()) {
      toast.error("死亡したペットのトリミング記録は削除できません");
      return;
    }
    input.deleteTrimming(input.id, {
      onSuccess: () => {
        toast.success("トリミング情報を削除しました");
        onSuccess?.();
      },
      onError: (error) => {
        handleApiError(error, "削除");
      },
    });
  };
}
