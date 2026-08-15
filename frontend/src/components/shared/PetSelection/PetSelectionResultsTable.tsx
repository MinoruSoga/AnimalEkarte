import { memo } from "react";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { Check, Octagon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Pagination } from "@/components/shared/Pagination";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { formatDate } from "@/lib/format/date";
import type { PetSelectionResultPage } from "@/hooks/use-pet-selection-page";
import type { Pet } from "@/types";

interface PetSelectionResultsTableProps {
  pets: PetSelectionResultPage;
  onSelect: (pet: Pet) => void;
  /**
   * ペット一覧の取得に失敗した状態。既定 false（従来どおり「該当0件」表示）。
   * true の場合は「0件」を主張せず、取得失敗であることを明示する。
   */
  isError?: boolean;
  /** ペット一覧の取得中。既定 false。true の場合も「0件」を主張しない。 */
  isLoading?: boolean;
}

export const PetSelectionResultsTable = memo(function PetSelectionResultsTable({ pets, onSelect, isError = false, isLoading = false }: PetSelectionResultsTableProps) {
  const {
    items,
    totalCount,
    currentPage,
    totalPages,
    startIndex,
    endIndex,
    onPageChange,
  } = pets;
  const rangeText = `${totalCount.toLocaleString()}件中 ${startIndex.toLocaleString()}-${endIndex.toLocaleString()}件`;
  // total と描画行が食い違う状態（応答に total が含まれない、範囲が総数を超える等）では
  // 件数を主張しない。「0件」と言いながら行を出すのは BUG-451 と同じ「嘘の件数」である。
  const isCountTrustworthy =
    totalCount > 0 ? startIndex <= totalCount : items.length === 0;
  // 取得失敗・取得中に現在行数を全体件数として提示しない。
  // キャッシュ行が残る場合は、backend 由来の前回総数・範囲だと明示する。
  const showCachedRange = items.length > 0 && (isError || isLoading);
  const statusText = showCachedRange
    ? `前回取得: ${rangeText}`
    : isError || isLoading || !isCountTrustworthy
      ? ""
      : totalCount === 0
        ? "0件"
        : rangeText;
  // 多ページ時は Pagination が同じ範囲を可視表示するため、こちらは読み上げ専用にする。
  // 領域自体を出し入れすると AT が更新として扱わず、単一ページへ絞り込まれた検索が
  // 無通知になるため、領域は常設しテキストだけを差し替える。
  const isRangeShownByPagination = !isError && totalPages > 1;
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <h2 className={`text-sm font-medium ${C.text}`}>検索結果</h2>
        <span
          role="status"
          aria-live="polite"
          aria-atomic="true"
          className={
            isRangeShownByPagination
              ? "sr-only"
              : `text-sm ${C.text60}`
          }
        >
          {statusText}
        </span>
      </div>

      {isError && items.length > 0 ? (
        <p role="alert" className={`text-sm ${C.danger}`}>
          ペット一覧の取得に失敗しました。前回取得した患者は選択できません
        </p>
      ) : null}

      <div className={`rounded-lg bg-white overflow-hidden border ${C.borderMedium}`}>
        <Table>
          <TableHeader>
            <TableRow
              className={`hover:bg-transparent ${C.bgPage} border-b ${C.borderMedium} h-12`}
            >
              <TableHead className={`min-w-[80px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>飼主No</TableHead>
              <TableHead className={`min-w-[140px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>飼主名</TableHead>
              <TableHead className={`min-w-[80px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>ペット番号</TableHead>
              <TableHead className={`min-w-[100px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>ペット名</TableHead>
              <TableHead className={`min-w-[50px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>生死</TableHead>
              <TableHead className={`min-w-[50px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>種</TableHead>
              <TableHead className={`min-w-[90px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>生年月日</TableHead>
              <TableHead className={`min-w-[60px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>体重</TableHead>
              <TableHead className={`min-w-[100px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>環境</TableHead>
              <TableHead className={`min-w-[90px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>前回来院</TableHead>
              <TableHead className={`min-w-[80px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((pet, index) => {
              const isDeceased = pet.status === "死亡";
              const isSelectable =
                pet.status === "生存" && !isError && !isLoading;
              return (
                <TableRow
                  key={pet.id}
                  className={`transition-colors ${C.hoverBgMedium} h-12 ${
                    index < items.length - 1 ? `border-b ${C.borderLight}` : "border-none"
                  } ${isDeceased ? "opacity-60 grayscale-[0.5]" : ""}`}
                >
                  <TableCell className={`text-sm ${C.text} whitespace-nowrap`}>{pet.ownerId}</TableCell>
                  <TableCell className={`text-sm ${C.text} whitespace-nowrap`}>{pet.ownerName}</TableCell>
                  <TableCell className={`font-mono text-sm ${C.text} whitespace-nowrap`}>{pet.petNumber || "-"}</TableCell>
                  <TableCell className={`text-sm ${C.text} whitespace-nowrap`}>
                    <span className="flex items-center gap-1.5">
                      <span>{pet.name}</span>
                      {pet.dangerLevel === "高" ? (
                        <Popover>
                          <PopoverTrigger asChild>
                            <button
                              type="button"
                              aria-label={`${pet.name}の危険理由を表示`}
                              className={`inline-flex items-center rounded px-1.5 py-0.5 text-xs font-semibold ${C.bgDanger10} ${C.danger} ${C.borderDanger20} outline-none focus-visible:ring-2 ${C.focusRingAccent40}`}
                            >
                              ⚠ 危険
                            </button>
                          </PopoverTrigger>
                          <PopoverContent
                            align="start"
                            aria-label={`${pet.name}の危険理由`}
                            onOpenAutoFocus={(event) => event.preventDefault()}
                            className="w-64"
                          >
                            <p className={`text-sm font-semibold ${C.danger}`}>危険理由</p>
                            <p className={`mt-1 whitespace-pre-wrap break-words text-sm ${C.textInkSecondary}`}>
                              {pet.dangerReason?.trim() || "理由未登録"}
                            </p>
                          </PopoverContent>
                        </Popover>
                      ) : null}
                    </span>
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {pet.status ? (
                      <Badge
                        variant="secondary"
                        className={
                          pet.status === "生存"
                            ? `${C.bgStatusGreen} ${C.textStatusGreen} ${C.borderStatusGreen} ${C.hoverBgStatusGreen} text-xs px-2 py-0 h-7`
                            : `${C.bgStatusGray} ${C.textStatusGray} ${C.borderStatusGray} ${C.hoverBgStatusGray} text-xs px-2 py-0 h-7`
                        }
                      >
                        {pet.status}
                      </Badge>
                    ) : null}
                  </TableCell>
                  <TableCell className={`text-sm ${C.text} whitespace-nowrap`}>{pet.species}</TableCell>
                  <TableCell className={`font-mono text-sm ${C.text} whitespace-nowrap`}>{formatDate(pet.birthDate)}</TableCell>
                  <TableCell className={`font-mono text-sm ${C.text} whitespace-nowrap`}>{pet.weight || "-"}</TableCell>
                  <TableCell className={`text-sm ${C.text} whitespace-nowrap`}>{pet.environment || "-"}</TableCell>
                  <TableCell className={`font-mono text-sm ${C.text} whitespace-nowrap`}>{formatDate(pet.lastVisit)}</TableCell>
                  <TableCell className="whitespace-nowrap">
                    <Button
                      size="sm"
                      variant={isSelectable ? "outline" : "ghost"}
                      disabled={!isSelectable}
                      aria-label={
                        isError
                          ? `取得失敗・選択不可: ${pet.name} (ID ${pet.id})`
                          : isLoading
                            ? `読み込み中・選択不可: ${pet.name} (ID ${pet.id})`
                          : isDeceased
                          ? `死亡・選択不可: ${pet.name} (ID ${pet.id})`
                          : pet.status === "生存"
                            ? `選択: ${pet.name} (ID ${pet.id})`
                            : `状態不明・選択不可: ${pet.name} (ID ${pet.id})`
                      }
                      // docs/spec/design-system.md button-primary: brand と同じ primary teal + pill
                      className={`h-11 min-w-11 gap-1 ${isSelectable ? `${C.bgActionPrimary} ${C.textOnActionPrimary} ${C.hoverBgActionPrimary} ${C.hoverTextOnActionPrimary} rounded-full` : C.textStatusGray} text-sm px-4`}
                      onClick={() => onSelect(pet)}
                    >
                      {isSelectable ? (
                        <>
                          <Check className={ICON.action} />
                          選択
                        </>
                      ) : (
                        <>
                          <Octagon className={ICON.action} />
                          選択不可
                        </>
                      )}
                    </Button>
                  </TableCell>
                </TableRow>
              );
            })}
            {items.length === 0 ? (
              <TableRow>
                {/* 取得失敗 > 取得中 > 該当0件 の優先順。色に依存せず文言で状態を区別する。 */}
                <TableCell colSpan={11} className={STYLE.tableEmpty}>
                  {isError ? (
                    <span role="alert" className={C.danger}>
                      ペット一覧の取得に失敗しました
                    </span>
                  ) : isLoading ? (
                    "ペット一覧を読み込み中です"
                  ) : (
                    "該当するペットが見つかりません"
                  )}
                </TableCell>
              </TableRow>
            ) : null}
          </TableBody>
        </Table>
      </div>
      {/* 操作可能なコントロールを live region へ入れない（全ボタンが読み上げられる）。
          読み込み中も disabled にしない — フォーカス中のボタンが無効化されると
          ブラウザが強制 blur し、ページ送りのたびにキーボード焦点が失われる。 */}
      {!isError && totalPages > 1 ? (
        <fieldset aria-busy={isLoading} className="border-0 p-0 m-0">
          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            totalCount={totalCount}
            startIndex={startIndex}
            endIndex={endIndex}
            onPageChange={onPageChange}
            onPrev={() => onPageChange(currentPage - 1)}
            onNext={() => onPageChange(currentPage + 1)}
          />
        </fieldset>
      ) : null}
    </div>
  );
});
