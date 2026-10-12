import { formatJSTWallDate, toJSTWallDate } from "@/lib/jst-date";
import { requireStoredClinicId } from "@/lib/current-clinic";
import { axios } from "@/lib/axios";

interface ReceptionCandidate {
  id: number;
  clinic_id: number;
  pet_id: number;
  owner_id: number;
  start_time: string;
  updated_at: string;
  status: string;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function isID(value: unknown): value is number {
  return typeof value === "number" && Number.isSafeInteger(value) && value > 0;
}

function isTimestamp(value: unknown): value is string {
  return (
    typeof value === "string" &&
    /^\d{4}-\d{2}-\d{2}T(?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d(?:\.\d{1,9})?(?:Z|[+-](?:[01]\d|2[0-3]):[0-5]\d)$/.test(
      value,
    ) &&
    Number.isFinite(Date.parse(value)) &&
    new Date(`${value.slice(0, 10)}T00:00:00Z`).toISOString().slice(0, 10) === value.slice(0, 10)
  );
}

function isCandidate(value: unknown): value is ReceptionCandidate {
  return (
    isRecord(value) &&
    isID(value.id) &&
    isID(value.clinic_id) &&
    isID(value.pet_id) &&
    isID(value.owner_id) &&
    isTimestamp(value.start_time) &&
    isTimestamp(value.updated_at) &&
    typeof value.status === "string" &&
    [
      "confirmed",
      "pending",
      "checked_in",
      "in_consultation",
      "accounting",
      "completed",
      "cancelled",
      "no_show",
      "tentative",
    ].includes(value.status)
  );
}

function isAlivePet(value: unknown, petId: string, ownerId: string, clinicId: string): boolean {
  return (
    isRecord(value) &&
    isID(value.id) &&
    String(value.id) === petId &&
    isID(value.clinic_id) &&
    String(value.clinic_id) === clinicId &&
    isID(value.owner_id) &&
    String(value.owner_id) === ownerId &&
    value.status === "alive" &&
    (value.deceased_at === undefined || value.deceased_at === null || value.deceased_at === "")
  );
}

/** Validate the exact read target before check-in; never create from an incomplete response. */
export async function getReceptionCandidates(date: string, petId: string, ownerId: string) {
  const clinicId = requireStoredClinicId();
  const config = { headers: { "X-Clinic-ID": clinicId } };
  const { data: pet } = await axios.get<unknown>(`/v1/pets/${encodeURIComponent(petId)}`, config);
  if (requireStoredClinicId() !== clinicId || !isAlivePet(pet, petId, ownerId, clinicId)) {
    throw new Error("ペットの状態または医院情報が変わりました。再読み込みしてください");
  }
  let rows: ReceptionCandidate[] = [];
  let page = 1;
  let expectedTotal: number | undefined;
  while (true) {
    const { data } = await axios.get<unknown>("/v1/reservations", {
      ...config,
      params: { date, pet_id: petId, owner_id: ownerId, page, limit: 100 },
    });
    if (
      !isRecord(data) ||
      !Array.isArray(data.data) ||
      typeof data.total !== "number" ||
      !Number.isSafeInteger(data.total) ||
      data.total < 0 ||
      (expectedTotal !== undefined && data.total !== expectedTotal) ||
      rows.length + data.data.length > data.total
    ) {
      throw new Error("予約一覧の件数が変わりました。再読み込みしてください");
    }
    expectedTotal = data.total;
    const validated = data.data.filter(isCandidate);
    if (
      requireStoredClinicId() !== clinicId ||
      validated.length !== data.data.length ||
      validated.some(
        (row) =>
          String(row.clinic_id) !== clinicId ||
          String(row.pet_id) !== petId ||
          String(row.owner_id) !== ownerId ||
          formatJSTWallDate(toJSTWallDate(new Date(row.start_time))) !== date,
      ) ||
      new Set(rows.concat(validated).map((row) => row.id)).size !== rows.length + validated.length
    ) {
      throw new Error("予約一覧の対象が変わりました。再読み込みしてください");
    }
    rows = rows.concat(validated);
    if (rows.length >= data.total) break;
    if (validated.length === 0) throw new Error("予約一覧を取得できませんでした");
    page += 1;
  }
  return rows.filter(
    (row) => !["cancelled", "tentative", "completed", "no_show"].includes(row.status),
  );
}
