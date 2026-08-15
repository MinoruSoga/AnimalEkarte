import { formatDate } from "@/lib/format/date";
import type { Owner, Pet } from "@/types";

import { formatDMPreference } from "../lib/dm-preference";
import { formatPetAge } from "../lib/pet-age";
import { ClinicalBriefingPanel } from "./ClinicalBriefingPanel";
import { DetailField } from "./ClinicalBriefingFields";

interface BasicDetailsProps {
  owner: Owner;
  pet: Pet;
  firstVisitDate?: string | null;
  firstVisitLoading: boolean;
  firstVisitError: boolean;
}

function PetBasicDetails({
  pet,
  firstVisitDate,
  firstVisitLoading,
  firstVisitError,
}: Omit<BasicDetailsProps, "owner">) {
  const firstVisitValue = firstVisitLoading
    ? "読み込み中..."
    : firstVisitError
      ? "取得失敗"
      : formatDate(firstVisitDate);
  return (
    <>
      <DetailField label="ふりがな" value={pet.petNameKana} />
      <DetailField
        label="年齢"
        value={pet.birthDate ? formatPetAge(pet.birthDate) : null}
      />
      <DetailField label="生年月日" value={formatDate(pet.birthDate)} />
      <DetailField label="性別" value={pet.gender} />
      <DetailField
        label="種類・品種"
        value={[pet.species, pet.breed].filter(Boolean).join("・")}
      />
      <DetailField label="毛色" value={pet.color} />
      <DetailField
        label="体重"
        value={pet.weight ? `${pet.weight} kg` : undefined}
      />
      <DetailField label="前回来院" value={formatDate(pet.lastVisit)} />
      <DetailField label="初診日" value={firstVisitValue} />
      <DetailField label="血液型" value={pet.bloodType} />
      <DetailField label="マイクロチップ" value={pet.microchipNumber} />
      <DetailField label="去勢・避妊日" value={formatDate(pet.neuteredDate)} />
      <DetailField label="入手方法" value={pet.acquisitionType} />
      <DetailField label="フード" value={pet.food} />
      <DetailField label="飼育環境" value={pet.environment} />
      <DetailField
        label="保険"
        value={[pet.insuranceName, pet.insuranceDetails]
          .filter(Boolean)
          .join("・")}
      />
    </>
  );
}

function OwnerBasicDetails({ owner }: Pick<BasicDetailsProps, "owner">) {
  const address = [
    owner.postalCode ? `〒${owner.postalCode}` : "",
    owner.address1,
    owner.address2,
  ]
    .filter(Boolean)
    .join(" ");
  return (
    <>
      <DetailField label="電話" value={owner.phone} />
      <DetailField label="会員区分" value={owner.membershipType} />
      {address ? <DetailField label="住所" value={address} /> : null}
      {owner.company ? (
        <DetailField label="勤務先" value={owner.company} />
      ) : null}
      {owner.companyPhone ? (
        <DetailField label="勤務先TEL" value={owner.companyPhone} />
      ) : null}
      {owner.email ? <DetailField label="メール" value={owner.email} /> : null}
      {owner.dmPreference != null ? (
        <DetailField
          label="DM"
          value={formatDMPreference(owner.dmPreference)}
        />
      ) : null}
    </>
  );
}

export function BasicInformationPanel(props: BasicDetailsProps) {
  return (
    <ClinicalBriefingPanel
      title="基本情報"
      description="ペットと飼主の基本情報"
      areaClassName="owner-report-area-basic"
      bodyClassName="px-2"
    >
      <dl className="grid min-w-0 grid-cols-2 gap-x-3 max-[520px]:grid-cols-1">
        <PetBasicDetails
          pet={props.pet}
          firstVisitDate={props.firstVisitDate}
          firstVisitLoading={props.firstVisitLoading}
          firstVisitError={props.firstVisitError}
        />
        <OwnerBasicDetails owner={props.owner} />
      </dl>
    </ClinicalBriefingPanel>
  );
}
