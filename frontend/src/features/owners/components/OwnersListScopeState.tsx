import { Button } from "@/components/ui/button";
import { C, STYLE } from "@/lib/design-tokens";
import type { OwnersListState } from "../loaders";

interface OwnersListScopeStateProps {
  listState: Exclude<OwnersListState, { kind: "ready" }>;
  onReload: () => void;
  isReloading: boolean;
}

/**
 * EMR-463: 一覧を取得しなかった理由の表示。
 * scope 取得失敗・許可ゼロ・選択超過を一覧 0 件と区別し、
 * 「何が無いか／条件／次の操作」の 3 点で示す（ui-design-guide §2 空状態）。
 */
export function OwnersListScopeState({
  listState,
  onReload,
  isReloading,
}: OwnersListScopeStateProps) {
  if (listState.kind === "scope-error") {
    return (
      <div role="alert" aria-atomic="true" className={STYLE.formCard}>
        <p className={`text-sm font-medium ${C.danger}`}>
          表示できる医院を確認できなかったため、飼主・ペット一覧を表示していません。
        </p>
        <p className={`mt-1 text-sm ${C.text60}`}>
          通信状態または権限情報の取得に問題があります。保存した表示医院の選択は変更していません。
        </p>
        <div className="mt-3">
          <Button
            type="button"
            variant="outline"
            onClick={onReload}
            disabled={isReloading}
            aria-busy={isReloading}
          >
            {isReloading ? "再読み込み中..." : "再読み込み"}
          </Button>
        </div>
      </div>
    );
  }

  if (listState.kind === "no-permitted-clinics") {
    return (
      <div role="status" aria-live="polite" aria-atomic="true" className={STYLE.formCard}>
        <p className={`text-sm font-medium ${C.text}`}>
          飼主・ペット一覧を表示できる医院がありません。
        </p>
        <p className={`mt-1 text-sm ${C.text60}`}>
          現在のアカウントには、所属医院のいずれにも飼主・ペットの閲覧権限がありません。
        </p>
        <p className={`mt-1 text-sm ${C.text60}`}>
          閲覧が必要な場合は、医院の管理者に権限の付与を依頼してください。
        </p>
      </div>
    );
  }

  return (
    <div role="alert" aria-atomic="true" className={STYLE.formCard}>
      <p className={`text-sm font-medium ${C.danger}`}>
        表示医院が{listState.selectedCount}院選択されているため、一覧を表示していません。
      </p>
      <p className={`mt-1 text-sm ${C.text60}`}>
        一度に表示できる医院は{listState.limit}院までです。
      </p>
      <p className={`mt-1 text-sm ${C.text60}`}>
        上の「表示医院」で選択を{listState.limit}院以下に減らしてください。
      </p>
    </div>
  );
}
