export interface ClinicalFixture {
  clinicId: number;
  ownerName: string;
  ownerSearch: string;
  petId: string;
  petName: string;
  outsideFirstPagePet: { id: string; name: string };
  estimateTitle: string;
  medicalRecordCount: number;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function readUint(value: unknown): number | null {
  if (typeof value === "number" && Number.isSafeInteger(value) && value > 0) return value;
  if (typeof value === "string" && /^\d+$/.test(value)) {
    const parsed = Number(value);
    return Number.isSafeInteger(parsed) && parsed > 0 ? parsed : null;
  }
  return null;
}

function readName(value: unknown): string | null {
  return typeof value === "string" && value.trim() !== "" ? value : null;
}

export function parseClinicalFixture(raw: string | undefined): ClinicalFixture {
  if (raw === undefined || raw.trim() === "") {
    throw new Error("clinical e2e fixture is missing");
  }
  let decoded: unknown;
  try {
    decoded = JSON.parse(raw) as unknown;
  } catch {
    throw new Error("clinical e2e fixture is not JSON");
  }
  if (!isRecord(decoded)) {
    throw new Error("clinical e2e fixture is not an object");
  }
  const clinicId = readUint(decoded.clinicId);
  const ownerName = readName(decoded.ownerName);
  const ownerSearch = readName(decoded.ownerSearch);
  const petId = readUint(decoded.petId);
  const petName = readName(decoded.petName);
  const outsideId = readUint(decoded.outsideFirstPagePetId);
  const outsideName = readName(decoded.outsideFirstPagePetName);
  const estimateTitle = readName(decoded.estimateTitle);
  const medicalRecordCount = readUint(decoded.medicalRecordCount);
  if (
    clinicId === null ||
    ownerName === null ||
    ownerSearch === null ||
    petId === null ||
    petName === null ||
    outsideId === null ||
    outsideName === null ||
    estimateTitle === null ||
    medicalRecordCount === null
  ) {
    throw new Error("clinical e2e fixture is incomplete");
  }
  if (clinicId === 1 || clinicId === 2) {
    throw new Error("clinical e2e clinic_id is reserved");
  }
  if (medicalRecordCount < 1) {
    throw new Error("clinical e2e fixture owner record is missing");
  }
  return {
    clinicId,
    ownerName,
    ownerSearch,
    petId: String(petId),
    petName,
    outsideFirstPagePet: { id: String(outsideId), name: outsideName },
    estimateTitle,
    medicalRecordCount,
  };
}

export function requireClinicalFixture(): ClinicalFixture {
  return parseClinicalFixture(process.env.E2E_CLINICAL_FIXTURE);
}
