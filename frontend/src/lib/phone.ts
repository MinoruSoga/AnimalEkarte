/** Characters accepted by backend owner phone validation. */
const OWNER_PHONE_CHARACTERS = /^[0-9+ ()-]+$/;

/**
 * True when phone matches backend-compatible character set and has ≥10 digits.
 * Shared by ReservationFormModal (line-reserve keeps a local copy for app boundary).
 */
export function isValidOwnerPhone(phone: string): boolean {
  return OWNER_PHONE_CHARACTERS.test(phone) && phone.replace(/\D/g, "").length >= 10;
}
