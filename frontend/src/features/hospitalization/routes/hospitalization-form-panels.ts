// FE-RC-045: 518行あった panels ファイルを chrome hook / status-view / header-extra / fields / body に分割。
// このファイルは既存の import 経路を維持するための re-export barrel。
export { HospitalizationFormStatusView } from "./HospitalizationFormStatusView";
export { HospitalizationFormBody } from "./HospitalizationFormBody";
