import { Trash2, Users } from "lucide-react";

import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { SearchableSelect } from "@/components/ui/searchable-select";
import { C, ICON, STYLE } from "@/lib/design-tokens";

import { usePetSubOwnersSection } from "../hooks/use-pet-sub-owners-section";

interface PetSubOwnersSectionProps {
  petId: string;
  canEdit: boolean;
}

export function PetSubOwnersSection({ petId, canEdit }: PetSubOwnersSectionProps) {
  const {
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
  } = usePetSubOwnersSection(petId, canEdit);

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

        <div className={`overflow-hidden rounded-lg border ${C.borderMedium} ${C.bgWhite}`}>
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
                      <p className={`truncate text-sm font-medium ${C.text}`}>{row.name}</p>
                      <p className={`truncate text-xs ${C.text60}`}>{row.nameKana}</p>
                    </div>
                    <div className="space-y-1">
                      <Label htmlFor={relationshipId} className={STYLE.sectionLabel}>
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
                        aria-invalid={invalidRelationshipOwnerId === row.ownerId}
                        aria-describedby={
                          invalidRelationshipOwnerId === row.ownerId
                            ? "pet-sub-owners-save-error"
                            : undefined
                        }
                        onChange={(event) =>
                          handleRelationshipChange(row.ownerId, event.target.value)
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
          <p className={`text-sm ${C.textSuccess}`} role="status" aria-live="polite">
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
