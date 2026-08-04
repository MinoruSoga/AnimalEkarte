import { useCallback } from "react";
import { Navigate, useParams, useSearchParams } from "react-router";

import { RequirePermission } from "@/components/shared/RequirePermission";
import { paths } from "@/config/paths";
import { useAuth } from "@/hooks/use-auth";
import { useGetOwner } from "@/hooks/use-owner";
import { useTitle } from "@/hooks/use-title";
import { C } from "@/lib/design-tokens";
import type { Owner, Pet } from "@/types";
import {
  ResourceMedicalRecords,
  ResourceOwners,
} from "@/types/generated/models";

import { useGetOwnerReportPets } from "../api/get-owner-report-pets";
import { useGetPetFirstVisit } from "../api/get-pet-first-visit";
import { OwnerClinicalBriefing } from "../components/OwnerClinicalBriefing";
import { OwnerReportPanel } from "../components/OwnerReportPanel";
import { PetSwitcher } from "../components/PetSwitcher";
import { SelectedPetContext } from "../components/SelectedPetContext";
import { toPet } from "../lib/owner-report-pet";

/**
 * #158 飼主単位カルテレポート。
 * 別ウィンドウで開く Layout 外スタンドアロンルート。
 * 自前で認証ガード + medical-records:view + owners:view ゲートを持つ。
 */
export function OwnerReport() {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) return null;
  if (!isAuthenticated)
    return <Navigate to={paths.auth.login.getHref()} replace />;
  return (
    <RequirePermission resource={ResourceMedicalRecords}>
      <RequirePermission resource={ResourceOwners}>
        <OwnerReportContent />
      </RequirePermission>
    </RequirePermission>
  );
}

function useOwnerPetSelection(pets: ReadonlyArray<Pet>) {
  const [searchParams, setSearchParams] = useSearchParams();
  const petIdParam = searchParams.get("petId");
  const selectedPetId =
    petIdParam && pets.some((pet) => pet.id === petIdParam)
      ? petIdParam
      : pets[0]?.id;
  const selectedPet = pets.find((pet) => pet.id === selectedPetId);
  const handleSelectPet = useCallback(
    (petId: string) => {
      setSearchParams(
        (previous) => {
          const next = new URLSearchParams(previous);
          next.set("petId", petId);
          return next;
        },
        { replace: true },
      );
    },
    [setSearchParams],
  );
  return { selectedPetId, selectedPet, handleSelectPet };
}

function ReportStatus({ error = false }: { error?: boolean }) {
  return (
    <div
      className={`flex min-h-dvh items-center justify-center ${C.bgPage} p-6`}
    >
      <p
        className={`text-sm ${error ? C.danger : C.text50}`}
        role={error ? "alert" : "status"}
        aria-live={error ? undefined : "polite"}
      >
        {error ? "飼主・ペット情報の取得に失敗しました" : "読み込み中..."}
      </p>
    </div>
  );
}

interface ReportHeaderProps {
  owner?: Owner;
  pets: ReadonlyArray<Pet>;
  selectedPet?: Pet;
  selectedPetId?: string;
  onSelectPet: (petId: string) => void;
}

function ReportHeader({
  owner,
  pets,
  selectedPet,
  selectedPetId,
  onSelectPet,
}: ReportHeaderProps) {
  return (
    <header
      className={`z-20 shrink-0 border-b px-2 py-1 shadow-level1 ${C.borderLight} ${C.bgWhite}`}
    >
      <div className="flex min-w-0 flex-wrap items-center justify-between gap-x-3 gap-y-1">
        {owner ? (
          <OwnerReportPanel owner={owner} />
        ) : (
          <p className={`text-sm ${C.danger}`}>
            飼主情報を取得できませんでした
          </p>
        )}
        <div className="flex min-w-0 items-center gap-3">
          {selectedPet ? <SelectedPetContext pet={selectedPet} /> : null}
          {pets.length > 0 ? (
            <PetSwitcher
              pets={[...pets]}
              selectedPetId={selectedPetId}
              onSelect={onSelectPet}
            />
          ) : null}
        </div>
      </div>
    </header>
  );
}

interface ReportBodyProps {
  owner?: Owner;
  pet?: Pet;
  firstVisitDate?: string | null;
  firstVisitLoading: boolean;
  firstVisitError: boolean;
}

function ReportBody(props: ReportBodyProps) {
  if (!props.owner || !props.pet) {
    return (
      <main className="min-h-0 flex-1 p-2">
        <div className="p-3">
          <p className={`text-sm ${C.text50}`}>
            この飼主に登録されたペットがありません
          </p>
        </div>
      </main>
    );
  }
  return (
    <OwnerClinicalBriefing
      owner={props.owner}
      pet={props.pet}
      firstVisitDate={props.firstVisitDate}
      firstVisitLoading={props.firstVisitLoading}
      firstVisitError={props.firstVisitError}
    />
  );
}


function OwnerReportContent() {
  const { id: ownerId = "" } = useParams();
  const ownerQuery = useGetOwner(ownerId);
  const petsQuery = useGetOwnerReportPets(ownerId);
  const pets = (petsQuery.data ?? []).map((pet) => toPet(pet, ownerId));
  const selection = useOwnerPetSelection(pets);
  const firstVisitQuery = useGetPetFirstVisit(selection.selectedPetId);

  useTitle(`飼主レポート - ${ownerQuery.data?.ownerName ?? ownerId}`);
  if (ownerQuery.isLoading || petsQuery.isLoading) return <ReportStatus />;
  if (ownerQuery.isError || petsQuery.isError) return <ReportStatus error />;
  return (
    <div
      data-testid="owner-report-viewport"
      className={`fixed inset-0 flex min-w-[320px] flex-col overflow-hidden ${C.bgPage}`}
    >
      <ReportHeader
        owner={ownerQuery.data}
        pets={pets}
        selectedPet={selection.selectedPet}
        selectedPetId={selection.selectedPetId}
        onSelectPet={selection.handleSelectPet}
      />
      <ReportBody
        owner={ownerQuery.data}
        pet={selection.selectedPet}
        firstVisitDate={firstVisitQuery.data}
        firstVisitLoading={firstVisitQuery.isLoading}
        firstVisitError={firstVisitQuery.isError}
      />
    </div>
  );
}
