/** UI-facing clinic info (camelCase) @see {@link import("@/types/generated/models").Clinic} */
export interface ClinicInfo {
  name: string;
  postalCode: string;
  address: string;
  phoneNumber: string;
  faxNumber?: string;
  registrationNumber?: string;
  directorName?: string;
  email?: string;
  website?: string;
  logoUrl?: string;
}
