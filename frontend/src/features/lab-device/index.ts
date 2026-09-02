export { LabDeviceBoard } from "./routes/LabDeviceBoard";
// FE-RC-015: 実装は components/shared へ昇格済み（examinations/medical-records から
// 直接参照可能にするため）。lab-device 内部の後方互換 re-export として維持する。
export { LabDeviceUnlinkedBanner } from "@/components/shared/LabDeviceUnlinkedBanner/LabDeviceUnlinkedBanner";
