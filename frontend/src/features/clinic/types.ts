export interface ClinicInfo {
  name: string;
  branchName: string;
  postalCode: string;
  address: string;
  phoneNumber: string;
  faxNumber?: string;
  registrationNumber?: string;
  directorName?: string;
  email?: string;
  website?: string;
  logoUrl?: string; // Optional: for image data or URL
}

export const DEFAULT_CLINIC_INFO: ClinicInfo = {
  name: "ノア動物病院",
  branchName: "八王子院",
  postalCode: "192-0083",
  address: "東京都八王子市旭町1-1",
  phoneNumber: "042-123-4567",
  registrationNumber: "東京都獣医師会 第12345号",
};
