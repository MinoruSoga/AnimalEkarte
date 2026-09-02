// FE-RC-015: 実装は src/hooks へ昇格済み（medical-records/CheckupsTab から直接参照可能に
// するため）。checkups 内部の後方互換 re-export として維持する。
export {
  replaceCheckupFieldResults,
  type CheckupFieldResultInput,
} from "@/hooks/use-checkup-fields";
