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
import {
  PetSwitcher,
  OWNER_REPORT_TABPANEL_ID,
  ownerReportPetTabId,
} from "../components/PetSwitcher";
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
      <div className={`flex min-h-dvh items-center justify-center ${C.bgPage} p-6`}>
        <p className={`text-sm ${C.text50}`} role="status" aria-live="polite">
          読み込み中...
        </p>
      </div>
    );
  }

  // 業務ツール向け密集レイアウト:
  // - root を固定ビューポート高(h-dvh)にする。グローバル CSS が html/body/#root を
  //   height:100% + overflow:hidden に固定しているため、root 自身をスクロールコンテナにする。
  // - lg+ では overflow-hidden で root も固定し、各履歴パネルだけが内部スクロールする（ページ非スクロール）。
  // - lg 未満（タブレット/モバイル）は root が overflow-y-auto でスクロールし、パネルは自然高さで縦積みする。
  // - 上部 <header> = 常時固定（sticky）の飼主コンテキスト + ペット切替（R4/R5）。
  // - <main> = 6 セクションを敷き詰めるグリッド（xl:3列×2行 / lg:2列×3行 / それ未満:1列）。
  return (
    <div className={`flex h-dvh flex-col overflow-y-auto ${C.bgPage} lg:overflow-hidden`}>
      {/* R4/R5: 飼主は固定表示、ペット切替は即アクセス可能。モバイルでも sticky で残す。 */}
      <header
        className={`sticky top-0 z-20 shrink-0 border-b ${C.borderLight} ${C.bgWhite} px-3 py-2`}
      >
        <div className="flex flex-wrap items-center justify-between gap-x-4 gap-y-2">
          {owner ? (
            <OwnerReportPanel owner={owner} />
          ) : (
            <p className={`text-sm ${C.danger}`}>飼主情報を取得できませんでした</p>
          )}
          {pets.length > 0 ? (
            <PetSwitcher pets={pets} selectedPetId={selectedPetId} onSelect={handleSelectPet} />
          ) : null}
        </div>
      </header>

      <main className="flex-1 lg:min-h-0 lg:overflow-hidden">
        {selectedPet ? (
          <div
            id={OWNER_REPORT_TABPANEL_ID}
            role="tabpanel"
            aria-labelledby={ownerReportPetTabId(selectedPet.id)}
            className="grid grid-cols-1 gap-2 p-2 lg:h-full lg:min-h-0 lg:grid-cols-2 lg:grid-rows-3 xl:grid-cols-3 xl:grid-rows-2"
          >
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
          <div className="p-3">
            <p className={`text-sm ${C.text50}`}>この飼主に登録されたペットがありません</p>
          </div>
        )}
      </main>
    </div>
  );
}
