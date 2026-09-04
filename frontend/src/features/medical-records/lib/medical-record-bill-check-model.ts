export type BillCheckExtraKind = "exam" | "vaccination";

export interface BillCheckExtraSource {
  id: number | string;
  name: string;
  price?: number | null;
  medicalRecordId?: string;
}

export interface BillCheckExtraLine {
  id: string;
  kind: BillCheckExtraKind;
  name: string;
  unitPrice: number | null;
}

export function isUnbillableMasterPrice(price: number | null | undefined): boolean {
  return price == null || !Number.isFinite(price) || price < 0;
}

export function billCheckExtraLines(
  kind: BillCheckExtraKind,
  sources: ReadonlyArray<BillCheckExtraSource>,
  medicalRecordId: string,
): BillCheckExtraLine[] {
  if (!medicalRecordId) {
    return [];
  }
  return sources
    .filter((source) => source.medicalRecordId === medicalRecordId)
    .map((source) => ({
      id: `${kind}_${source.id}`,
      kind,
      name: source.name,
      unitPrice: source.price ?? null,
    }));
}

export function billCheckPricedExtras(
  lines: ReadonlyArray<BillCheckExtraLine>,
): BillCheckExtraLine[] {
  return lines.filter((line) => !isUnbillableMasterPrice(line.unitPrice));
}
