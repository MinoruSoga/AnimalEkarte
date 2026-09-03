/**
 * OwnerFormPage - app層での feature 合成
 * owners feature と pets feature と line-reservation feature を app層でのみ合成し、
 * feature 間 import を排除する。
 */
import { useState, lazy } from "react";
import { useLoaderData, useParams } from "react-router";
import { Send } from "lucide-react";

// features（app層なので複数 feature を import 可能）
import { OwnerForm, LineIntegrationCard, LineSendPanel } from "@/features/owners";
import type { OwnerLoaderData } from "@/features/owners";
import { createPet, useCreatePet, useUpdatePet, useDeletePet } from "@/features/pets";
import { useRevokePetDeath } from "@/hooks/use-revoke-pet-death";
import { LinkedLineCustomers } from "@/features/line-reservation";
import { useAuth } from "@/hooks/use-auth";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { ICON } from "@/lib/design-tokens";
import type { PetMutations } from "@/types/pet";

// Lazy-loaded — 編集モードでのみ必要、新規登録画面のバンドルから切り離す。
// feature 間 import (owners → accounting) を避けるため app 層でのみ import する。
const OwnerAccountingHistory = lazy(() =>
  import("@/features/accounting").then((m) => ({ default: m.OwnerAccountingHistory })),
);

export function OwnerFormPage() {
  const { id: ownerId } = useParams();
  const loaderData = useLoaderData<OwnerLoaderData | undefined>();
  const { user } = useAuth();
  const clinicId = user?.mainClinicId ?? null;

  const { mutate: createPetMutate } = useCreatePet();
  const { mutate: updatePetMutate } = useUpdatePet();
  const { mutate: deletePetMutate } = useDeletePet();
  const { mutate: revokePetDeathMutate } = useRevokePetDeath();

  const owner = loaderData?.owner;
  const ownerName = owner?.ownerName ?? "";

  const [sendPanelOpen, setSendPanelOpen] = useState(false);

  const petMutations: PetMutations = {
    createPetFn: createPet,
    createPetMutate: (req, { onSuccess, onError }) => createPetMutate(req, { onSuccess, onError }),
    updatePetMutate: (args, { onSuccess, onError }) =>
      updatePetMutate(args, { onSuccess, onError }),
    deletePetMutate: (id, { onSuccess, onError }) => deletePetMutate(id, { onSuccess, onError }),
    revokePetDeathMutate: (petId) => revokePetDeathMutate(petId),
  };

  // 会計履歴セクション（編集モード=ownerIdがある時のみ意味がある。表示可否は OwnerForm 側の権限判定に委ねる）
  const accountingSection = ownerId ? <OwnerAccountingHistory ownerId={ownerId} /> : null;

  // LINE連携セクション（編集モード=ownerIdがある時のみ意味がある）
  const lineSection = ownerId ? (
    <div className="space-y-4">
      <LinkedLineCustomers clinicId={clinicId} ownerId={Number(ownerId)} />
      <LineIntegrationCard ownerId={ownerId} ownerName={ownerName} owner={owner} />
      <div className="flex justify-end">
        <PrimaryButton
          type="button"
          colorVariant="primary"
          onClick={() => setSendPanelOpen(true)}
          className="px-4 text-base"
        >
          <Send className={ICON.sm} />
          個別LINE送信
        </PrimaryButton>
      </div>
      <LineSendPanel
        ownerId={ownerId}
        ownerName={ownerName}
        open={sendPanelOpen}
        onOpenChange={setSendPanelOpen}
      />
    </div>
  ) : null;

  return (
    <OwnerForm
      petMutations={petMutations}
      lineSection={lineSection}
      accountingSection={accountingSection}
    />
  );
}
