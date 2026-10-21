import { boundSucceededRefs } from "./certificate-binding";
import type {
  CertificateRecoveryItem,
  CertificateRecoveryQueryRequest,
  CertificateRecoveryQueryResponse,
} from "./certificate-recovery-schemas";

/**
 * EMR-487 Slice A: 回復 discovery の応答を「要求した対象」に束縛する純関数。
 * zod は形しか保証しない。別 intent・別接種・別版・別 key の結果を画面に載せない。
 */

type RecoveryTargetRequest = CertificateRecoveryQueryRequest;

function editionIdOf(request: RecoveryTargetRequest): string | undefined {
  return request.intent === "reprint" ? String(request.target_ref.edition_id) : undefined;
}

export function recoveryItemMatchesRequest(item: CertificateRecoveryItem, request: RecoveryTargetRequest): boolean {
  if (item.intent !== request.intent) return false;
  if (item.result === null) return true;
  if (item.result.operation_key !== item.operation_key) return false;
  if (item.execution_state !== "succeeded") return true;
  // 成功は、結果が指す接種（reprint なら版も）が要求した対象と一致するときだけ採用する。
  return (
    boundSucceededRefs(item.result, {
      operationKey: item.operation_key,
      vaccinationId: request.target_ref.vaccination_id,
      editionId: editionIdOf(request),
    }) !== null
  );
}

export function recoveryPageMatchesRequest(
  page: CertificateRecoveryQueryResponse,
  request: RecoveryTargetRequest,
): boolean {
  return page.items.every((item) => recoveryItemMatchesRequest(item, request));
}
