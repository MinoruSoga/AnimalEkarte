import { useState, useCallback, useMemo } from "react";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { usePermission } from "@/hooks/use-permission";
import { ResourceMasterReservationType } from "@/types/generated/models";
import {
  useGetReservationTypes,
  useCreateReservationType,
  useUpdateReservationType,
  useDeleteReservationType,
} from "../api/reservation-types";
import type { ReservationType } from "../api/reservation-types";
import {
  useGetReservationTypeGroups,
  useCreateReservationTypeGroup,
  useUpdateReservationTypeGroup,
  useDeleteReservationTypeGroup,
} from "../api/reservation-type-groups";
import type { ReservationTypeGroup } from "../api/reservation-type-groups";
import type {
  CreateReservationTypeGroupRequest,
  UpdateReservationTypeGroupRequest,
} from "../api/reservation-type-groups";
import type {
  CreateReservationTypeRequest,
  UpdateReservationTypeRequest,
} from "../api/reservation-types";
import type { GroupFormData } from "../components/ReservationTypeGroupSidePanel";
import type { CategoryFormData } from "../components/ReservationTypeSidePanel";
import {
  buildReservationTypeCreateRequest,
  buildReservationTypeGroupCreateRequest,
  buildReservationTypeGroupUpdateRequest,
  buildReservationTypeUpdateRequest,
  getActiveReservationTypeGroupOptions,
  matchesReservationTypeSearch,
  matchesReservationTypeStatusFilter,
} from "./reservation-type-settings-model";

export function useReservationTypeSettings() {
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterReservationType);

  const { data: groupsRaw = [] } = useGetReservationTypeGroups();
  const { data: categoriesRaw = [] } = useGetReservationTypes();

  const activeGroups = useMemo(() => getActiveReservationTypeGroupOptions(groupsRaw), [groupsRaw]);

  const createGroupMutation = useCreateReservationTypeGroup();
  const updateGroupMutation = useUpdateReservationTypeGroup();
  const deleteGroupMutation = useDeleteReservationTypeGroup();
  const createCategoryMutation = useCreateReservationType();
  const updateCategoryMutation = useUpdateReservationType();
  const deleteCategoryMutation = useDeleteReservationType();

  const dirty = useSidePeekDirty();
  const handleDirtyChange = useCallback(
    (d: boolean) => {
      if (d) dirty.markDirty();
      else dirty.markClean();
    },
    [dirty],
  );

  const groupCrud = useMasterCRUD<ReservationTypeGroup>({
    data: groupsRaw,
    deleteMutation: deleteGroupMutation,
    entityLabel: "予約区分グループ",
    dirtyGuard: dirty,
    permissions: { canDelete },
  });

  const categoryCrud = useMasterCRUD<ReservationType>({
    data: categoriesRaw,
    deleteMutation: deleteCategoryMutation,
    entityLabel: "予約区分",
    dirtyGuard: dirty,
    permissions: { canDelete },
    searchFilter: matchesReservationTypeSearch,
    activeFilterApply: matchesReservationTypeStatusFilter,
  });

  const [categoryDefaultGroupId, setCategoryDefaultGroupId] = useState<string | undefined>(
    undefined,
  );

  const groupSave = useMasterSave<
    ReservationTypeGroup,
    GroupFormData,
    CreateReservationTypeGroupRequest,
    UpdateReservationTypeGroupRequest
  >({
    crud: groupCrud,
    createMutation: createGroupMutation,
    updateMutation: updateGroupMutation,
    validate: (data) => (data.name.trim() ? null : "名称を入力してください"),
    toCreateRequest: buildReservationTypeGroupCreateRequest,
    toUpdateRequest: buildReservationTypeGroupUpdateRequest,
    permissions: { canCreate, canEdit },
  });

  const categorySave = useMasterSave<
    ReservationType,
    CategoryFormData,
    CreateReservationTypeRequest,
    UpdateReservationTypeRequest
  >({
    crud: categoryCrud,
    createMutation: createCategoryMutation,
    updateMutation: updateCategoryMutation,
    validate: (data) => (data.name.trim() ? null : "名称を入力してください"),
    toCreateRequest: buildReservationTypeCreateRequest,
    toUpdateRequest: buildReservationTypeUpdateRequest,
    permissions: { canCreate, canEdit },
  });

  const handleGroupEdit = useCallback(
    (group: ReservationTypeGroup) => {
      groupCrud.handleEdit(group);
    },
    [groupCrud],
  );

  const handleGroupAdd = useCallback(() => {
    groupCrud.handleNew();
  }, [groupCrud]);

  const handleGroupDeleteRequest = useCallback(
    (item: ReservationTypeGroup) => {
      groupCrud.handleDeleteRequest(item);
    },
    [groupCrud],
  );

  const handleCategoryEdit = useCallback(
    (cat: ReservationType) => {
      categoryCrud.handleEdit(cat);
      setCategoryDefaultGroupId(undefined);
    },
    [categoryCrud],
  );

  const handleCategoryAddInGroup = useCallback(
    (groupId: string | undefined) => {
      categoryCrud.handleNew();
      setCategoryDefaultGroupId(groupId);
    },
    [categoryCrud],
  );

  const handleCategoryDeleteRequest = useCallback(
    (item: ReservationType) => {
      categoryCrud.handleDeleteRequest(item);
    },
    [categoryCrud],
  );

  return {
    canCreate,
    canEdit,
    canDelete,
    groupsRaw,
    activeGroups,
    dirty,
    handleDirtyChange,
    groupCrud,
    categoryCrud,
    categoryDefaultGroupId,
    groupSave,
    categorySave,
    handleGroupEdit,
    handleGroupAdd,
    handleGroupDeleteRequest,
    handleCategoryEdit,
    handleCategoryAddInGroup,
    handleCategoryDeleteRequest,
  };
}
