import { useActionState, useMemo, useState } from "react";

import type { SearchableSelectOption } from "@/components/ui/searchable-select";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { handleApiError } from "@/lib/handle-api-error";

import {
  useGetPetSubOwnerMetadata,
  useGetPetSubOwners,
  useGetSubOwnerCandidates,
} from "../api/get-pet-sub-owners";
import { useReplacePetSubOwners } from "../api/replace-pet-sub-owners";
import {
  getSaveErrorMessage,
  INITIAL_SAVE_STATE,
  SUB_OWNER_SEARCH_DEBOUNCE_MS,
  toEditableSubOwners,
  type SaveState,
} from "../lib/pet-sub-owners-section-model";

export function usePetSubOwnersSection(petId: string, canEdit: boolean) {
  const subOwnersQuery = useGetPetSubOwners(petId);
  const metadataQuery = useGetPetSubOwnerMetadata(petId);
  const replaceMutation = useReplacePetSubOwners();
  const [draftRows, setDraftRows] = useState<ReturnType<typeof toEditableSubOwners> | null>(null);
  const [candidateSearch, setCandidateSearch] = useState("");
  const [selectedOwnerId, setSelectedOwnerId] = useState("");
  const [invalidRelationshipOwnerId, setInvalidRelationshipOwnerId] =
    useState<number | null>(null);
  const debouncedCandidateSearch = useDebouncedValue(
    candidateSearch,
    SUB_OWNER_SEARCH_DEBOUNCE_MS,
  );
  const normalizedCandidateSearch = candidateSearch.trim();
  const normalizedDebouncedSearch = debouncedCandidateSearch.trim();
  const isCandidateSearchPending =
    normalizedCandidateSearch !== normalizedDebouncedSearch;
  const candidatesQuery = useGetSubOwnerCandidates(
    normalizedDebouncedSearch,
    canEdit,
  );
  const rows = useMemo(
    () =>
      draftRows ??
      toEditableSubOwners(subOwnersQuery.data?.sub_owners ?? []),
    [draftRows, subOwnersQuery.data],
  );
  const isDirty = draftRows !== null;

  const candidateOptions = useMemo<SearchableSelectOption[]>(() => {
    const primaryOwnerId = metadataQuery.data?.owner_id;
    const selectedIds = new Set(rows.map((row) => row.ownerId));
    return (candidatesQuery.data ?? [])
      .filter(
        (candidate) =>
          candidate.ownerId !== primaryOwnerId &&
          !selectedIds.has(candidate.ownerId),
      )
      .map((candidate) => ({
        value: String(candidate.ownerId),
        label: candidate.name,
        keywords: [candidate.nameKana],
      }));
  }, [candidatesQuery.data, metadataQuery.data?.owner_id, rows]);

  const handleAddSubOwner = (value: string) => {
    const ownerId = Number(value);
    const candidate = candidatesQuery.data?.find(
      (item) => item.ownerId === ownerId,
    );
    if (candidate === undefined) {
      return;
    }
    setDraftRows((currentRows) => [
      ...(currentRows ?? rows),
      {
        ownerId: candidate.ownerId,
        name: candidate.name,
        nameKana: candidate.nameKana,
        relationship: "",
      },
    ]);
    setCandidateSearch("");
    setSelectedOwnerId("");
    setInvalidRelationshipOwnerId(null);
  };

  const handleRelationshipChange = (ownerId: number, relationship: string) => {
    setDraftRows((currentRows) =>
      (currentRows ?? rows).map((row) =>
        row.ownerId === ownerId ? { ...row, relationship } : row,
      ),
    );
    setInvalidRelationshipOwnerId(null);
  };

  const handleRemoveSubOwner = (ownerId: number) => {
    setDraftRows((currentRows) =>
      (currentRows ?? rows).filter((row) => row.ownerId !== ownerId),
    );
    setInvalidRelationshipOwnerId(null);
    document.getElementById("pet-sub-owner-search")?.focus();
  };

  const [saveState, saveAction, isSavePending] = useActionState<
    SaveState,
    FormData
  >(
    async () => {
      if (
        subOwnersQuery.data === undefined ||
        subOwnersQuery.error !== null
      ) {
        return {
          kind: "error",
          message:
            "副飼主情報を取得できないため保存できません。再読み込みしてください。",
        };
      }
      const invalidRelationship = rows.find((row) => {
        const length = Array.from(row.relationship.trim()).length;
        return length < 1 || length > 50;
      });
      if (invalidRelationship !== undefined) {
        setInvalidRelationshipOwnerId(invalidRelationship.ownerId);
        return {
          kind: "error",
          message: `${invalidRelationship.name}の続柄は1〜50文字で入力してください。`,
        };
      }
      if (metadataQuery.data === undefined) {
        return {
          kind: "error",
          message:
            "ペットの更新情報を取得できませんでした。再読み込みしてください。",
        };
      }

      try {
        await replaceMutation.mutateAsync({
          petId,
          request: {
            version: metadataQuery.data.version,
            sub_owners: rows.map((row) => ({
              owner_id: row.ownerId,
              relationship: row.relationship.trim(),
            })),
          },
        });
        setInvalidRelationshipOwnerId(null);
        setDraftRows(null);
        return {
          kind: "success",
          message: "副飼主を保存しました",
        };
      } catch (error: unknown) {
        handleApiError(error, "副飼主保存");
        return {
          kind: "error",
          message: getSaveErrorMessage(error),
        };
      }
    },
    INITIAL_SAVE_STATE,
  );

  const loadError =
    subOwnersQuery.error ?? metadataQuery.error ?? candidatesQuery.error;

  return {
    subOwnersQuery,
    metadataQuery,
    candidatesQuery,
    rows,
    isDirty,
    candidateSearch,
    setCandidateSearch,
    selectedOwnerId,
    invalidRelationshipOwnerId,
    isCandidateSearchPending,
    normalizedCandidateSearch,
    candidateOptions,
    handleAddSubOwner,
    handleRelationshipChange,
    handleRemoveSubOwner,
    saveState,
    saveAction,
    isSavePending,
    loadError,
  };
}
