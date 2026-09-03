import { Button } from "@/components/ui/button";
import { C } from "@/lib/design-tokens";
import type { Owner } from "@/types/owner";

import {
  DeliveryStatusBanner,
  LineDeliveryControls,
  LineIdConfirmSection,
} from "./LineIntegrationCardSections";
import {
  LineIntegrationCardFrame,
  LineIntegrationDialogs,
  LineIntegrationError,
  LineIntegrationLoading,
  LineLinkTokenSection,
  LinkedStatusRow,
  LstepTagsSection,
  UnlinkedLineIdForm,
  UnlinkedStatusRow,
} from "./LineIntegrationCardParts";
import { useLineIntegrationCardState } from "../hooks/use-line-integration-card-state";

interface LineIntegrationCardProps {
  ownerId: string;
  ownerName: string;
  owner?: Owner;
}

export function LineIntegrationCard({ ownerId, ownerName, owner }: LineIntegrationCardProps) {
  const state = useLineIntegrationCardState({ ownerId, owner });

  if (state.isLoading) {
    return <LineIntegrationLoading />;
  }

  if (state.isError || !state.data) {
    return <LineIntegrationError />;
  }

  const { is_linked, lstep_opt_out } = state.data;
  const bannerSection = (
    <DeliveryStatusBanner
      isStopped={state.isDeliveryStopped}
      stopReason={state.deliveryStopReason}
      caution={owner?.deliveryCaution}
      cautionReason={owner?.deliveryCautionReason ?? undefined}
    />
  );

  const lineIdConfirmSection = (
    <LineIdConfirmSection
      isLinked={is_linked}
      lineUserId={state.lineUserId}
      lineIdConfirmedAt={state.lineIdConfirmedAt}
      canEdit={state.canEdit}
      isPending={state.isConfirmingLineId}
      onConfirm={() => state.confirmLineId()}
    />
  );

  const deliveryControls = (
    <LineDeliveryControls
      canEdit={state.canEdit}
      owner={owner}
      deliveryReasonInputRef={state.deliveryReasonInputRef}
      deliveryCautionReasonInputRef={state.deliveryCautionReasonInputRef}
      isUpdatingDeliveryExclusion={state.isUpdatingDeliveryExclusion}
      isUpdatingDeliveryCaution={state.isUpdatingDeliveryCaution}
      isUpdatingTransferStatus={state.isUpdatingTransferStatus}
      onUpdateDeliveryExclusion={state.updateDeliveryExclusion}
      onUpdateDeliveryCaution={state.updateDeliveryCaution}
      onUpdateTransferStatus={state.updateTransferStatus}
      onTransferEnableRequest={() => state.setConfirmTransferOpen(true)}
    />
  );

  const dialogs = (
    <LineIntegrationDialogs
      ownerId={ownerId}
      ownerName={ownerName}
      tagAddDialogOpen={state.tagAddDialogOpen}
      confirmUnlinkOpen={state.confirmUnlinkOpen}
      confirmOptOutOpen={state.confirmOptOutOpen}
      confirmTransferOpen={state.confirmTransferOpen}
      isDeletingLine={state.isDeletingLine}
      isUpdatingDeliveryExclusion={state.isUpdatingDeliveryExclusion}
      isUpdatingTransferStatus={state.isUpdatingTransferStatus}
      onTagAddOpenChange={state.setTagAddDialogOpen}
      onUnlinkClose={() => state.setConfirmUnlinkOpen(false)}
      onUnlinkConfirm={state.confirmUnlink}
      onOptOutClose={() => state.setConfirmOptOutOpen(false)}
      onOptOutConfirm={state.confirmOptOut}
      onTransferClose={() => state.setConfirmTransferOpen(false)}
      onTransferConfirm={state.confirmTransfer}
    />
  );

  if (is_linked && lstep_opt_out) {
    return (
      <LineIntegrationCardFrame>
        {bannerSection}

        {state.canEdit ? (
          <div className="flex justify-end">
            <Button
              type="button"
              size="sm"
              variant="outline"
              className="h-8 px-3 text-xs shrink-0"
              disabled={state.isUpdatingDeliveryExclusion}
              onClick={state.resumeDelivery}
            >
              {state.isUpdatingDeliveryExclusion ? "処理中..." : "配信を再開"}
            </Button>
          </div>
        ) : null}

        {lineIdConfirmSection}
        <LstepTagsSection
          ownerId={ownerId}
          tags={state.tags}
          canEdit={false}
          removeTagName={null}
          disabled
          onRemoveTagNameChange={() => undefined}
          onAddClick={() => undefined}
        />
        {deliveryControls}
        {dialogs}
      </LineIntegrationCardFrame>
    );
  }

  if (is_linked) {
    return (
      <LineIntegrationCardFrame>
        {bannerSection}
        <LinkedStatusRow
          lineUserId={state.lineUserId}
          canEdit={state.canEdit}
          onUnlinkClick={() => state.setConfirmUnlinkOpen(true)}
        />
        {lineIdConfirmSection}
        <LstepTagsSection
          ownerId={ownerId}
          tags={state.tags}
          canEdit={state.canEdit}
          removeTagName={state.removeTagName}
          onRemoveTagNameChange={state.setRemoveTagName}
          onAddClick={() => state.setTagAddDialogOpen(true)}
        />

        {state.canEdit ? (
          <div className={`border-t ${C.borderLight} pt-3`}>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className={`h-8 px-2 text-xs ${C.text40} ${C.hoverText60}`}
              onClick={() => state.setConfirmOptOutOpen(true)}
            >
              配信を停止する
            </Button>
          </div>
        ) : null}

        {deliveryControls}
        {dialogs}
      </LineIntegrationCardFrame>
    );
  }

  return (
    <LineIntegrationCardFrame>
      {bannerSection}
      <UnlinkedStatusRow />
      <UnlinkedLineIdForm
        canEdit={state.canEdit}
        lineIdFormAction={state.lineIdFormAction}
        lineIdState={state.lineIdState}
      />
      <LineLinkTokenSection
        canEdit={state.canEdit}
        liffUrl={state.linkTokenResult?.liffUrl}
        isGenerating={state.isGeneratingLinkToken}
        onGenerate={() => state.generateLinkToken()}
      />
      {deliveryControls}
      {dialogs}
    </LineIntegrationCardFrame>
  );
}
