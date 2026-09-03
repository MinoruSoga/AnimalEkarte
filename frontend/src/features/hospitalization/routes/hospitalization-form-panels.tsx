// FE-RC-045: 518行あった panels ファイルを chrome hook / status-view / header-extra / fields / body に分割。
// このファイルは既存の import 経路 (`./hospitalization-form-panels`) を維持するための re-export barrel。
export { useHospitalizationFormChrome } from "./hospitalization-form-chrome";
export { HospitalizationFormStatusView } from "./hospitalization-form-status-view";
export { HospitalizationFormBody } from "./hospitalization-form-body";
