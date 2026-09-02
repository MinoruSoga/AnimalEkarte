import { useCallback, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router";
import { toast } from "sonner";

import { usePermission } from "@/hooks/use-permission";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { ResourceLabImport } from "@/types/generated/models";

import { useGetAllExaminationTypes } from "../api/exam-types-master";
import {
  useCreateLabDevice,
  useGetLabDevices,
  useSaveLabDeviceConfiguration,
} from "../api/lab-devices";
import {
  useEnsureLabDeviceItemMasters,
  useGetLabDeviceItemMasters,
} from "../api/lab-device-item-masters";
import {
  availableLabDeviceSourceTypes,
  itemsForLabDevice,
  parseLabDeviceSourceQuery,
  toLabDeviceRows,
  type LabDeviceFormData,
  type LabDeviceItemDraft,
  type LabDeviceRow,
} from "./lab-device-item-master-settings-model";
import { persistLabDeviceItemMaster } from "./lab-device-item-master-persist";

export function useLabDeviceItemMasterSettings() {
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit } = usePermission(ResourceLabImport);
  const { data: devices = [], isFetched: devicesFetched } = useGetLabDevices();
  const { data: items = [] } = useGetLabDeviceItemMasters();
  const { data: examTypes = [] } = useGetAllExaminationTypes();
  const ensureMutation = useEnsureLabDeviceItemMasters();
  const createMutation = useCreateLabDevice();
  const saveConfigurationMutation = useSaveLabDeviceConfiguration();
  const dirty = useSidePeekDirty();
  const sourceFromQuery = parseLabDeviceSourceQuery(searchParams.get("source"));
  const fromBoard = searchParams.get("from") === "board";
  const [selectedId, setSelectedId] = useState<string | "new" | null>(null);

  const rows = useMemo(
    () => toLabDeviceRows(devices, items, examTypes),
    [devices, examTypes, items],
  );
  const unusedSourceTypes = useMemo(
    () => availableLabDeviceSourceTypes(devices),
    [devices],
  );
  const selectedRow = selectedId === null || selectedId === "new"
    ? null
    : rows.find((row) => row.id === selectedId) ?? null;
  const isCreating = selectedId === "new";
  const selectedItems = useMemo(
    () => (selectedRow === null ? [] : itemsForLabDevice(items, selectedRow.sourceType)),
    [items, selectedRow],
  );
  const showPanel = isCreating || selectedRow !== null;
  const readOnly = isCreating ? canCreate !== true : canEdit !== true;

  useEffect(() => {
    if (sourceFromQuery === null) {
      return;
    }
    const match = rows.find((row) => row.sourceType === sourceFromQuery);
    if (match !== undefined) {
      setSelectedId(match.id);
    }
  }, [rows, sourceFromQuery]);

  const clearSourceParam = useCallback(() => {
    if (searchParams.has("source")) {
      const next = new URLSearchParams(searchParams);
      next.delete("source");
      setSearchParams(next, { replace: true });
    }
  }, [searchParams, setSearchParams]);

  const handleClose = useCallback(() => {
    dirty.runWithDiscardCheck(() => {
      setSelectedId(null);
      clearSourceParam();
      dirty.markClean();
    });
  }, [clearSourceParam, dirty.markClean, dirty.runWithDiscardCheck]);

  const handleEdit = useCallback((row: LabDeviceRow) => {
    dirty.runWithDiscardCheck(() => {
      setSelectedId(row.id);
    });
  }, [dirty.runWithDiscardCheck]);

  const handleNew = useCallback(() => {
    dirty.runWithDiscardCheck(() => {
      if (unusedSourceTypes.length === 0) {
        toast.error("対応プロトコルはすべて登録済みです");
        return;
      }
      setSelectedId("new");
    });
  }, [dirty.runWithDiscardCheck, unusedSourceTypes.length]);

  const handleDirtyChange = useCallback((nextDirty: boolean) => {
    if (nextDirty) {
      dirty.markDirty();
      return;
    }
    dirty.markClean();
  }, [dirty]);

  const handleSave = useCallback(async (form: LabDeviceFormData, drafts: LabDeviceItemDraft[]) => {
    const result = await persistLabDeviceItemMaster({
      form,
      drafts,
      isCreating,
      canCreate,
      canEdit,
      selectedRow,
      selectedItems,
      createDevice: (req) => createMutation.mutateAsync(req),
      saveConfiguration: (args) => saveConfigurationMutation.mutateAsync(args),
    });
    if (result !== "saved") {
      return;
    }
    dirty.markClean();
    setSelectedId(null);
    clearSourceParam();
  }, [
    canCreate,
    canEdit,
    clearSourceParam,
    createMutation,
    dirty,
    isCreating,
    selectedItems,
    selectedRow,
    saveConfigurationMutation,
  ]);

  return {
    canCreate,
    canEdit,
    devicesFetched,
    ensureMutation,
    createMutation,
    saveConfigurationMutation,
    dirty,
    sourceFromQuery,
    fromBoard,
    selectedId,
    rows,
    unusedSourceTypes,
    selectedRow,
    selectedItems,
    showPanel,
    readOnly,
    examTypes,
    handleClose,
    handleEdit,
    handleNew,
    handleDirtyChange,
    handleSave,
  };
}
