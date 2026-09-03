import { Navigate } from "react-router";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { C } from "@/lib/design-tokens";
import { usePermission } from "@/hooks/use-permission";
import { ResourceIdentityLinks } from "@/types/generated/models";

import { OwnerLinkSection } from "../components/OwnerLinkSection";
import { PetLinkSection } from "../components/PetLinkSection";
import { useIdentityLinksWorkbench } from "../hooks/use-identity-links-workbench";

/**
 * Phase 1 manual identity-link UI: search → select → link/unlink → minimal history.
 * view gates page; edit gates mutation controls (DEC-15 manage → edit action).
 */
export function IdentityLinksPage() {
  const { canView, canEdit } = usePermission(ResourceIdentityLinks);
  if (!canView) {
    return <Navigate to="/" replace />;
  }

  return <IdentityLinksWorkbench canEdit={canEdit} />;
}

function IdentityLinksWorkbench({ canEdit }: { canEdit: boolean }) {
  const workbench = useIdentityLinksWorkbench(canEdit);

  return (
    <PageLayout
      title="同一飼主・ペット連携"
      description="所属医院内の手動リンクのみ。権限のない医院の ID はサーバ側で拒否されます。"
      resource={ResourceIdentityLinks}
    >
      <div className="space-y-6">
        {!canEdit ? (
          <p className={`text-sm ${C.textWarning}`} role="status">
            閲覧のみ（連携の変更権限がありません）
          </p>
        ) : null}

        {workbench.errorMessage ? (
          <div
            className={`rounded border p-3 text-sm ${C.borderDanger} ${C.bgDanger10} ${C.textWarning}`}
            role="alert"
          >
            {workbench.errorMessage}
          </div>
        ) : null}

        <OwnerLinkSection
          canEdit={canEdit}
          pending={workbench.pending}
          ownerQuery={workbench.ownerQuery}
          setOwnerQuery={workbench.setOwnerQuery}
          ownerHits={workbench.ownerHits}
          selectedOwners={workbench.selectedOwners}
          ownerGroupId={workbench.ownerGroupId}
          toggleOwner={workbench.toggleOwner}
          onLinkOwners={workbench.onLinkOwners}
          onUnlinkOwner={workbench.onUnlinkOwner}
          resolveOwnerGroupId={workbench.resolveOwnerGroupId}
        />

        <PetLinkSection
          canEdit={canEdit}
          pending={workbench.pending}
          petQuery={workbench.petQuery}
          setPetQuery={workbench.setPetQuery}
          petHits={workbench.petHits}
          selectedPets={workbench.selectedPets}
          petGroupId={workbench.petGroupId}
          canLinkPets={workbench.canLinkPets}
          historyText={workbench.historyText}
          togglePet={workbench.togglePet}
          onLinkPets={workbench.onLinkPets}
          onUnlinkPet={workbench.onUnlinkPet}
          onLoadHistory={workbench.onLoadHistory}
          resolvePetGroupId={workbench.resolvePetGroupId}
        />
      </div>
    </PageLayout>
  );
}
