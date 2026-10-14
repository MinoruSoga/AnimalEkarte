/**
 * PR665: 権限リソースの文字列リテラルを機能内で 1 箇所に集約する定数。
 * TASK-444-S1 により @/types/generated/models の新規 import は禁止されているため、
 * generated の ResourceMasterPermission（"master-permission"）と同値の定数を
 * 機能内で定義する。値がずれた場合は権限判定が変わるため変更禁止。
 */
export const RESOURCE_MASTER_PERMISSION = "master-permission";
