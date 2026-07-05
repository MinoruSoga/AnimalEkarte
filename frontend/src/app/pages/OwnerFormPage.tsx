/**
 * OwnerFormPage - app層での feature 合成
 * owners feature と pets feature と line-reservation feature を app層でのみ合成し、
 * feature 間 import を排除する。
 */
import { useState } from "react";
import { useParams } from "react-router";
import { Send } from "lucide-react";

// features（app層なので複数 feature を import 可能）
import {
  OwnerForm,
  LineIntegrationCard,
  LineSendPanel,
  useGetOwner,
} from "@/features/owners";
import { createPet, useCreatePet, useUpdatePet, useDeletePet } from "@/features/pets";
import { LinkedLineCustomers } from "@/features/line-reservation";
import { useAuth } from "@/hooks/use-auth";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { ICON } from "@/lib/design-tokens";
import type { PetMutations } from "@/types/pet";

export function OwnerFormPage() {
  const { id: ownerId } = useParams();
  const { user } = useAuth();
  const clinicId = user?.mainClinicId ?? null;

  const { mutate: createPetMutate } = useCreatePet();
  const { mutate: updatePetMutate } = useUpdatePet();
  const { mutate: deletePetMutate } = useDeletePet();

  const { data: owner } = useGetOwner(ownerId ?? "");
  const ownerName = owner?.ownerName ?? "";

  const [sendPanelOpen, setSendPanelOpen] = useState(false);

  const petMutations: PetMutations = {
    createPetFn: createPet,
    createPetMutate: (req, { onSuccess, onError }) =>
      createPetMutate(req, { onSuccess, onError }),
    updatePetMutate: (args, { onSuccess, onError }) =>
      updatePetMutate(args, { onSuccess, onError }),
    deletePetMutate: (id, { onSuccess, onError }) =>
      deletePetMutate(id, { onSuccess, onError }),
  };

  // LINE連携セクション（編集モード=ownerIdがある時のみ意味がある）
  const lineSection = ownerId ? (
    <div className="space-y-4">
      <LinkedLineCustomers clinicId={clinicId} ownerId={Number(ownerId)} />
      <LineIntegrationCard ownerId={ownerId} ownerName={ownerName} owner={owner} />
      <div className="flex justify-end">
        <PrimaryButton
          type="button"
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

  return <OwnerForm petMutations={petMutations} lineSection={lineSection} />;
}
