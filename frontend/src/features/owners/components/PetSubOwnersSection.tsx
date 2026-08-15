import { isAxiosError } from "axios";
import { Trash2, Users } from "lucide-react";
import {
  useActionState,
  useMemo,
  useState,
} from "react";

import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  SearchableSelect,
  type SearchableSelectOption,
} from "@/components/ui/searchable-select";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { handleApiError } from "@/lib/handle-api-error";

import {
  useGetPetSubOwnerMetadata,
  useGetPetSubOwners,
  useGetSubOwnerCandidates,
} from "../api/get-pet-sub-owners";
import { useReplacePetSubOwners } from "../api/replace-pet-sub-owners";

interface PetSubOwnersSectionProps {
  petId: string;
  canEdit: boolean;
}

interface EditableSubOwner {
  ownerId: number;
  name: string;
  nameKana: string;
  relationship: string;
}

interface SaveState {
  kind: "idle" | "success" | "error";
  message: string;
}

interface ApiErrorResponse {
  error?: unknown;
}

const INITIAL_SAVE_STATE: SaveState = {
  kind: "idle",
  message: "",
};

const VERSION_CONFLICT_MESSAGE =
  "他の端末でペット情報が変更されました。再読み込みしてから、もう一度保存してください。";
const SUB_OWNER_SEARCH_DEBOUNCE_MS = 300;

function toEditableSubOwners(
  subOwners: ReadonlyArray<{
    owner_id: number;
    name: string;
    name_kana: string;
    relationship: string;
  }>,
): EditableSubOwner[] {
  return subOwners.map((subOwner) => ({
    ownerId: subOwner.owner_id,
    name: subOwner.name,
    nameKana: subOwner.name_kana,
    relationship: subOwner.relationship,
  }));
}

function getSaveErrorMessage(error: unknown): string {
  if (!isAxiosError<ApiErrorResponse>(error)) {
    return "副飼主を保存できませんでした。時間をおいて再度お試しください。";
  }
  if (error.response?.status === 409) {
    return VERSION_CONFLICT_MESSAGE;
  }
  if (error.response?.status === 400) {
    const serverMessage = error.response.data?.error;
    return typeof serverMessage === "string" && serverMessage.trim() !== ""
      ? `副飼主を保存できませんでした。${serverMessage}`
      : "副飼主を保存できませんでした。入力内容を確認してください。";
  }
  return "副飼主を保存できませんでした。時間をおいて再度お試しください。";
}

export function PetSubOwnersSection({
  petId,
  canEdit,
}: PetSubOwnersSectionProps) {
  const subOwnersQuery = useGetPetSubOwners(petId);
  const metadataQuery = useGetPetSubOwnerMetadata(petId);
  const replaceMutation = useReplacePetSubOwners();
  const [draftRows, setDraftRows] = useState<EditableSubOwner[] | null>(null);
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

  return (
    <section
      aria-labelledby="pet-sub-owners-title"
      className="col-span-1 space-y-3 md:col-span-2 lg:col-span-3"
    >
      <h2
        id="pet-sub-owners-title"
        className={`flex items-center gap-2 text-sm font-bold ${C.text}`}
      >
        <Users className={`${ICON.action} ${C.text60}`} aria-hidden="true" />
        副飼主
      </h2>

      {loadError !== null ? (
        <p className={`text-sm ${C.danger}`} role="alert">
          副飼主情報を取得できませんでした。
        </p>
      ) : null}

      <form action={saveAction} className="space-y-3">
        <div className="space-y-1">
          <Label htmlFor="pet-sub-owner-search" className={STYLE.sectionLabel}>
            副飼主を検索
          </Label>
          <Input
            id="pet-sub-owner-search"
            type="search"
            value={candidateSearch}
            disabled={
              !canEdit ||
              isSavePending ||
              subOwnersQuery.data === undefined ||
              subOwnersQuery.error !== null
            }
            placeholder="飼主名・よみ・電話番号"
            onChange={(event) => setCandidateSearch(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                event.preventDefault();
              }
            }}
          />
        </div>

        <div className="space-y-1">
          <Label htmlFor="pet-sub-owner-select" className={STYLE.sectionLabel}>
            副飼主を追加
          </Label>
          <SearchableSelect
            id="pet-sub-owner-select"
            ariaLabel="副飼主を追加"
            value={selectedOwnerId}
            onValueChange={handleAddSubOwner}
            options={candidateOptions}
            disabled={
              !canEdit ||
              isSavePending ||
              normalizedCandidateSearch === "" ||
              isCandidateSearchPending ||
              candidatesQuery.isLoading ||
              candidatesQuery.error !== null ||
              subOwnersQuery.data === undefined ||
              subOwnersQuery.error !== null
            }
            placeholder={
              normalizedCandidateSearch === ""
                ? "検索語を入力してください"
                : isCandidateSearchPending || candidatesQuery.isLoading
                ? "飼主を読み込み中..."
                : "飼主を選択してください"
            }
            searchPlaceholder="飼主名・よみで検索..."
            emptyMessage="追加できる飼主が見つかりません。"
          />
        </div>

        <div
          className={`overflow-hidden rounded-lg border ${C.borderMedium} ${C.bgWhite}`}
        >
          {rows.length === 0 ? (
            <p className={`px-4 py-6 text-center text-sm ${C.text60}`}>
              副飼主は登録されていません。
            </p>
          ) : (
            <ul className="divide-y">
              {rows.map((row) => {
                const relationshipId = `pet-sub-owner-relationship-${row.ownerId}`;
                return (
                  <li
                    key={row.ownerId}
                    className={`grid grid-cols-1 gap-3 p-3 md:grid-cols-[minmax(0,1fr)_minmax(12rem,1fr)_auto] md:items-end ${C.borderDivider}`}
                  >
                    <div className="min-w-0">
                      <p className={`truncate text-sm font-medium ${C.text}`}>
                        {row.name}
                      </p>
                      <p className={`truncate text-xs ${C.text60}`}>
                        {row.nameKana}
                      </p>
                    </div>
                    <div className="space-y-1">
                      <Label
                        htmlFor={relationshipId}
                        className={STYLE.sectionLabel}
                      >
                        {`続柄（${row.name}）`}
                      </Label>
                      <Input
                        id={relationshipId}
                        value={row.relationship}
                        disabled={
                          !canEdit ||
                          isSavePending ||
                          subOwnersQuery.data === undefined ||
                          subOwnersQuery.error !== null
                        }
                        aria-invalid={
                          invalidRelationshipOwnerId === row.ownerId
                        }
                        aria-describedby={
                          invalidRelationshipOwnerId === row.ownerId
                            ? "pet-sub-owners-save-error"
                            : undefined
                        }
                        onChange={(event) =>
                          handleRelationshipChange(
                            row.ownerId,
                            event.target.value,
                          )
                        }
                      />
                    </div>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      disabled={
                        !canEdit ||
                        isSavePending ||
                        subOwnersQuery.data === undefined ||
                        subOwnersQuery.error !== null
                      }
                      aria-label={`副飼主 ${row.name}を削除`}
                      onClick={() => handleRemoveSubOwner(row.ownerId)}
                      className={`${C.danger} ${C.borderDanger}`}
                    >
                      <Trash2 className={ICON.action} aria-hidden="true" />
                      削除
                    </Button>
                  </li>
                );
              })}
            </ul>
          )}
        </div>

        {saveState.kind === "error" ? (
          <p
            id="pet-sub-owners-save-error"
            className={`text-sm ${C.danger}`}
            role="alert"
            aria-live="assertive"
          >
            {saveState.message}
          </p>
        ) : null}
        {saveState.kind === "success" && !isDirty ? (
          <p
            className={`text-sm ${C.textSuccess}`}
            role="status"
            aria-live="polite"
          >
            {saveState.message}
          </p>
        ) : null}

        <div className="flex justify-end">
          <SubmitButton
            disabled={
              !canEdit ||
              subOwnersQuery.isLoading ||
              subOwnersQuery.data === undefined ||
              subOwnersQuery.error !== null ||
              metadataQuery.isLoading ||
              metadataQuery.data === undefined ||
              metadataQuery.error !== null
            }
            loadingText="副飼主を保存中..."
            className="text-sm"
          >
            副飼主を保存
          </SubmitButton>
        </div>
      </form>
    </section>
  );
}
