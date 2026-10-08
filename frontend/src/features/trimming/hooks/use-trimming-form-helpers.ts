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
  const previewRef = useRef<string | null>(null);

  useEffect(() => {
    previewRef.current = preview;
  }, [preview]);

  // アンマウント時: in-flight を無効化し、blob URL と未保存アップロードを解放する。
  useEffect(
    () => () => {
      requestIdRef.current += 1;
      for (const url of blobUrlsRef.current) {
        URL.revokeObjectURL(url);
      }
      blobUrlsRef.current.clear();
      if (sessionKeyRef.current) {
        discardUploadedKey(sessionKeyRef.current);
        sessionKeyRef.current = null;
      }
    },
    [],
  );

  const releaseBlobIfUnused = useCallback((blobUrl: string) => {
    if (previewRef.current === blobUrl) return;
    if (blobUrlsRef.current.delete(blobUrl)) {
      URL.revokeObjectURL(blobUrl);
    }
  }, []);

  const handleChange = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const file = selectedImageFile(event);
      if (!file) return undefined;
      const requestId = requestIdRef.current + 1;
      requestIdRef.current = requestId;
      const previousPreview = previewRef.current;
      const blobUrl = URL.createObjectURL(file);
      blobUrlsRef.current.add(blobUrl);
      setPreview(blobUrl); // アップロード完了を待たず即時プレビュー
      setPendingCount((count) => count + 1);
      const isCurrent = () => requestIdRef.current === requestId;
      return uploadTrimmingImage(file)
        .then((uploaded) => {
          if (!isCurrent()) {
            discardUploadedKey(uploaded.key);
            return;
          }
          const replacedSessionKey = sessionKeyRef.current;
          sessionKeyRef.current = uploaded.key;
          setLocalOverrides((prev) => ({ ...prev, [field]: uploaded.key }));
          markDirty();
          if (uploaded.signed_url !== "") {
            setPreview(uploaded.signed_url);
          }
          // 未保存で置き換わった先行アップロードは到達不能 → 掃除
          if (replacedSessionKey && replacedSessionKey !== uploaded.key) {
            discardUploadedKey(replacedSessionKey);
          }
        })
        .catch((error) => {
          if (!isCurrent()) return;
          setPreview(previousPreview);
          handleApiError(error, "画像アップロード");
        })
        .finally(() => {
          setPendingCount((count) => count - 1);
          releaseBlobIfUnused(blobUrl);
        });
    },
    [field, setLocalOverrides, markDirty, releaseBlobIfUnused],
  );

  const remove = useCallback(() => {
    requestIdRef.current += 1;
    setLocalOverrides((prev) => ({ ...prev, [field]: "" }));
    markDirty();
    setPreview(null);
    const sessionKey = sessionKeyRef.current;
    sessionKeyRef.current = null;
    if (sessionKey) {
      discardUploadedKey(sessionKey);
    }
    for (const url of blobUrlsRef.current) {
      URL.revokeObjectURL(url);
    }
    blobUrlsRef.current.clear();
  }, [field, setLocalOverrides, markDirty]);

  // 保存成功でキーの所有権がレコードへ移った後に呼ぶ。以後の削除は backend の
  // superseded 処理が担うため、session 追跡を止める。
  const commitUploadedKey = useCallback(() => {
    sessionKeyRef.current = null;
  }, []);

  return { preview, setPreview, pendingCount, handleChange, remove, commitUploadedKey };
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

export function useTrimmingFormImages(
  setLocalOverrides: TrimmingImageOverrides,
  markDirty: () => void,
) {
  const styleImage = useTrimmingImageField("styleImage", setLocalOverrides, markDirty);
  const completedImage = useTrimmingImageField("completedImage", setLocalOverrides, markDirty);
  const commitStyleImageKey = styleImage.commitUploadedKey;
  const commitCompletedImageKey = completedImage.commitUploadedKey;

  const commitUploadedKeys = useCallback(() => {
    commitStyleImageKey();
    commitCompletedImageKey();
  }, [commitStyleImageKey, commitCompletedImageKey]);

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
    commitUploadedKeys,
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
