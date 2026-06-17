import { useCallback } from "react";
import { Navigate, useParams, useSearchParams } from "react-router";

import { C } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { useAuth } from "@/hooks/use-auth";
import { useTitle } from "@/hooks/use-title";
import { RequirePermission } from "@/components/shared/RequirePermission";
import { ResourceMedicalRecords } from "@/types/generated/models";
import { useGetOwner } from "@/features/owners";
import { useGetPets } from "@/features/pets";

import { OwnerReportPanel } from "../components/OwnerReportPanel";
import { PetSwitcher } from "../components/PetSwitcher";
import { PetDetailSection } from "../components/PetDetailSection";
import { VaccinationHistorySection } from "../components/VaccinationHistorySection";
import { ExaminationHistorySection } from "../components/ExaminationHistorySection";
import { TreatmentHistorySection } from "../components/TreatmentHistorySection";

/**
 * #158 飼主単位カルテレポート。
 * 別ウィンドウで開く Layout 外スタンドアロンルート。自前で認証ガード + medical-records:view ゲートを持つ。
 */
export function OwnerReport() {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) return null;
  if (!isAuthenticated) {
    return <Navigate to={paths.auth.login.getHref()} replace />;
  }

  return (
    <RequirePermission resource={ResourceMedicalRecords}>
      <OwnerReportContent />
    </RequirePermission>
  );
}

function OwnerReportContent() {
  const { id: ownerId = "" } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();
  const { data: owner, isLoading: ownerLoading } = useGetOwner(ownerId);
  const { data: pets = [], isLoading: petsLoading } = useGetPets(ownerId);

  // 別ウィンドウのタブ識別用にタイトルを設定する。
  useTitle(`飼主レポート - ${owner?.ownerName ?? ownerId}`);

  // D1/D2: 初期ペット = ?petId= が有効なら採用、なければ最初の関連ペット。
  const petIdParam = searchParams.get("petId");
  const selectedPetId =
    petIdParam && pets.some((p) => p.id === petIdParam) ? petIdParam : pets[0]?.id;
  const selectedPet = pets.find((p) => p.id === selectedPetId);

  // R5: ページ遷移なしの state 更新 + URL ?petId= 同期。
  const handleSelectPet = useCallback(
    (petId: string) => {
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev);
          next.set("petId", petId);
          return next;
        },
        { replace: true },
      );
    },
    [setSearchParams],
  );

  if (ownerLoading || petsLoading) {
    return (
      <div className={`min-h-screen ${C.bgPage} p-6`}>
        <p className={`text-sm ${C.text50}`}>読み込み中...</p>
      </div>
    );
  }

  return (
    <div className={`min-h-screen ${C.bgPage}`}>
      <div className="mx-auto flex max-w-4xl flex-col gap-4 p-4 sm:p-6">
        {/* R4: 飼主パネルは常時固定表示（ペット切替で消えない） */}
        {owner ? (
          <OwnerReportPanel owner={owner} />
        ) : (
          <div className={`rounded-lg border ${C.borderLight} ${C.bgWhite} p-4`}>
            <p className={`text-sm ${C.danger}`}>飼主情報を取得できませんでした</p>
          </div>
        )}

        {pets.length > 0 ? (
          <PetSwitcher pets={pets} selectedPetId={selectedPetId} onSelect={handleSelectPet} />
        ) : null}

        {selectedPet ? (
          <div className="flex flex-col gap-4">
            <PetDetailSection pet={selectedPet} />
            <VaccinationHistorySection petId={selectedPet.id} />
            <ExaminationHistorySection petId={selectedPet.id} />
            <TreatmentHistorySection
              petId={selectedPet.id}
              title="投薬履歴"
              filter="medicine"
              emptyMessage="投薬の履歴はありません"
            />
            <TreatmentHistorySection
              petId={selectedPet.id}
              title="手術・処置履歴"
              filter="procedure"
              showAnesthesia
              emptyMessage="手術・処置の履歴はありません"
            />
            <TreatmentHistorySection
              petId={selectedPet.id}
              title="治療履歴"
              filter="all"
              emptyMessage="治療の履歴はありません"
            />
          </div>
        ) : (
          <div className={`rounded-lg border ${C.borderLight} ${C.bgWhite} p-4`}>
            <p className={`text-sm ${C.text50}`}>この飼主に登録されたペットがありません</p>
          </div>
        )}
      </div>
    </div>
  );
}
