// EMR-474: 再診問診 submit の Idempotency-Key 生成。
// 契約（api.yaml RevisitIntakeOperationKey）: lowercase hyphenated / non-nil UUID。暗黙正規化されないため
// 生成側で小文字化する。crypto.randomUUID が使えない非 secure context では getRandomValues で v4 を組み立てる。

function uuidV4FromRandomValues(): string {
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  bytes[6] = (bytes[6] & 0x0f) | 0x40;
  bytes[8] = (bytes[8] & 0x3f) | 0x80;
  const hex = Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}

export function createRevisitIntakeOperationKey(): string {
  const key =
    typeof crypto.randomUUID === "function" ? crypto.randomUUID() : uuidV4FromRandomValues();
  return key.toLowerCase();
}
