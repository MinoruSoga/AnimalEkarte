import { formatJSTWallDate, toJSTWallDate } from "@/lib/jst-date";
import { axios } from "@/lib/axios";
import type { PetResponse } from "@/types/generated/pet-responses";
import type { ReservationApiResponse } from "@/hooks/reservation-intent-response";

/** Read again immediately before check-in; failed or incomplete reads must never create. */
export async function getReceptionCandidates(date: string, petId: string, ownerId: string) {
  const { data: pet } = await axios.get<PetResponse>(`/v1/pets/${petId}`);
  if (pet.status !== "alive" || pet.deceased_at || String(pet.owner_id) !== ownerId) {
    throw new Error("ペットの状態または飼主情報が変わりました。再読み込みしてください");
  }
  let rows: ReservationApiResponse[] = [];
  let page = 1;
  let expectedTotal: number | undefined;
  while (true) {
    const { data } = await axios.get<{ data: ReservationApiResponse[]; total: number }>(
      "/v1/reservations",
      {
        params: { date, pet_id: petId, owner_id: ownerId, page, limit: 100 },
      },
    );
    if (
      !Array.isArray(data.data) ||
      !Number.isSafeInteger(data.total) ||
      data.total < 0 ||
      (expectedTotal !== undefined && data.total !== expectedTotal) ||
      rows.length + data.data.length > data.total
    ) {
      throw new Error("予約一覧の件数が変わりました。再読み込みしてください");
    }
    expectedTotal = data.total;
    if (
      data.data.some(
        (row) =>
          String(row.pet_id) !== petId ||
          String(row.owner_id) !== ownerId ||
          !row.start_time ||
          !Number.isFinite(Date.parse(row.start_time)) ||
          formatJSTWallDate(toJSTWallDate(new Date(row.start_time))) !== date ||
          rows.some((previous) => previous.id === row.id),
      )
    ) {
      throw new Error("予約一覧の対象が変わりました。再読み込みしてください");
    }
    rows = rows.concat(data.data);
    if (rows.length >= data.total) break;
    if (data.data.length === 0) throw new Error("予約一覧を取得できませんでした");
    page += 1;
  }
  return rows.filter(
    (row) => !["cancelled", "tentative", "completed", "no_show"].includes(String(row.status)),
  );
}
