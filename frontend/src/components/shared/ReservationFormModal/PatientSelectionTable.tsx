// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import { useState, useMemo, useCallback, memo } from "react";

// External
import { Check, Search, SearchX, RotateCcw } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { EmptyState, ErrorFallback } from "@/components/shared/DataStates";
import { Pagination } from "@/components/shared/Pagination";
import { useGetPets } from "@/hooks/use-pet";
import { useAnimalSpecies } from "@/hooks/use-animal-species";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { SEARCH_DEBOUNCE_MS } from "@/hooks/use-pet-selection-page";

// Types
import type { Pet } from "@/types";

interface PatientSelectionTableProps {
  onSelect: (pet: Pet) => void;
  selectedPets: Pet[];
  /** 直接記録の患者変更では死亡 sentinel も表示し、選択不可を明示する。 */
  includeDeceased?: boolean;
}

/**
 * 検索条件は backend の述語と1対1で対応させる（BUG-452）。
 * `search` は pets.name / pets.name_kana / owners.name / owners.name_kana /
 * owners.phone を横断する ILIKE 検索（backend/internal/pet/repository.go:123-139）、
 * `species` は animal_species_id、`ownerId` は owner_id。
 *
 * 旧実装は「飼主名 / 飼主名よみ / 電話番号 / ペット名 / ペット名よみ」の5欄を持ち、
 * 取得済みの先頭20件に対してクライアント側で絞り込んでいた。5欄は backend の
 * `search` 1個でカバーされる。backend に述語の無い条件（住所など）を足してはならない
 * — FE 側で補うと総件数と描画行が食い違い、利用者は「その条件で全件を検索した」と
 * 誤解する。
 */
interface PatientSearchParams {
  search: string;
  ownerId: string;
  species: string;
}

const INITIAL_SEARCH_PARAMS: PatientSearchParams = {
  search: "",
  ownerId: "",
  species: "",
};

const PAGE_SIZE = 20;

const ALL_SPECIES_VALUE = "__all__";

const TEXT_FIELDS = [
  {
    key: "search",
    label: "検索（ペット名・飼主名・よみ・電話）",
    placeholder: "例: もも、山田、090",
  },
  { key: "ownerId", label: "飼主No", placeholder: "例: 30042" },
] as const;

function SpeciesSelectField({
  value,
  onChange,
}: {
  value: string;
  onChange: (species: string) => void;
}) {
  const { activeSpecies, isLoading, isError } = useAnimalSpecies();
  // TASK-SPECIES-01: error → loading → success かつ0件 → options を区別する。
  // 取得できていない状態を「該当なし」と同じ見た目にすると、利用者は種別を
  // 絞り込めていないことに気づけない。
  const isUnavailable = isLoading || isError;
  // 状態文言をトリガーへ結び付ける。結ばないと、無効化されたセレクトへ到達した
  // 利用者に「なぜ選べないのか」が届かない（PetSelectionSearchForm と同じ扱い）。
  const hasStatusMessage = isError || isLoading || activeSpecies.length === 0;

  return (
    <div className="space-y-0.5">
      <Label htmlFor="species" className={`text-xs ${C.text60}`}>
        種別
      </Label>
      <Select
        value={value || ALL_SPECIES_VALUE}
        onValueChange={(next) =>
          onChange(next === ALL_SPECIES_VALUE ? "" : next)
        }
        disabled={isUnavailable}
      >
        <SelectTrigger
          id="species"
          aria-describedby={hasStatusMessage ? "species-status" : undefined}
          className={`text-xs h-11 bg-white ${C.borderMediumLight} ${C.text}`}
        >
          <SelectValue placeholder="すべて" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={ALL_SPECIES_VALUE}>すべて</SelectItem>
          {activeSpecies.map((species) => (
            <SelectItem key={species.id} value={String(species.id)}>
              {species.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {isError ? (
        <p
          id="species-status"
          role="alert"
          aria-atomic="true"
          className={`text-xs ${C.danger}`}
        >
          動物種の取得に失敗しました。
        </p>
      ) : isLoading ? (
        <p
          id="species-status"
          role="status"
          aria-live="polite"
          aria-atomic="true"
          className={`text-xs ${C.text50}`}
        >
          動物種を読み込み中です。
        </p>
      ) : activeSpecies.length === 0 ? (
        <p
          id="species-status"
          role="status"
          aria-live="polite"
          aria-atomic="true"
          className={`text-xs ${C.text50}`}
        >
          動物種マスタが登録されていません。
        </p>
      ) : null}
    </div>
  );
}

export const PatientSelectionTable = memo(function PatientSelectionTable({
  onSelect,
  selectedPets,
  includeDeceased = false,
}: PatientSelectionTableProps) {
  const [searchParams, setSearchParams] = useState<PatientSearchParams>(
    INITIAL_SEARCH_PARAMS,
  );

  // 入力停止後にだけ問い合わせる。押しても何も起きない検索ボタンは持たない。
  // useDebouncedValue は参照の同一性で判定するため、3条件は必ず1個の state
  // オブジェクトにまとめて渡す（レンダーごとに作り直すと確定しなくなる）。
  const debouncedSearchParams = useDebouncedValue(
    searchParams,
    SEARCH_DEBOUNCE_MS,
  );

  // 入力からデバウンス確定までの間、表示中の一覧は「まだ古い検索条件の結果」である。
  const isSearchPending = searchParams !== debouncedSearchParams;

  // 条件が1つも無いうちは問い合わせない。モーダルを開いた直後に一覧の先頭ページを
  // 出すと、それが全件だと誤解される（BUG-452 の障害モードと同型）。判定は必ず
  // 確定済み(debounced)の条件で行う — 入力中の値で有効化すると、1打鍵目で
  // 「条件なしの一覧」を取りに行ってしまう。
  const hasSearchConditions = useMemo(
    () =>
      Object.values(debouncedSearchParams).some((value) => value.trim() !== ""),
    [debouncedSearchParams],
  );

  // ページ番号は「それを選んだ時点で確定していた検索条件」と対で保持する。
  // 対にしないと、古い totalPages を根拠に選んだ page が新条件のクエリへ持ち越され、
  // 「5件中 21-5件」のような反転した範囲になる（BUG-451）。
  const [pageBinding, setPageBinding] = useState<{
    page: number;
    params: PatientSearchParams;
  }>({ page: 1, params: INITIAL_SEARCH_PARAMS });
  const page =
    pageBinding.params === debouncedSearchParams ? pageBinding.page : 1;

  // 検索条件は全て backend の述語へ委譲する。ここで絞り込みを足すと、総件数を
  // 根拠に「N件中 1-20件」と表示しながらその20件から黙って行を消すことになり、
  // 利用者は実在患者を未登録と誤認する。
  //
  // error を破棄してはならない。破棄すると API 失敗が「該当0件」と区別できず、
  // 利用者に嘘の検索結果を見せる（BUG-2026-07-27-01 で pets 一覧全滅が
  // 「0件」として現れ、原因特定が遅れた）。
  const {
    data: pets = [],
    total,
    page: rawResponsePage,
    limit: rawResponseLimit,
    error,
    isLoading,
    isPlaceholderData,
  } = useGetPets(
    debouncedSearchParams.ownerId || undefined,
    {
      ...(includeDeceased ? { includeDeceased: true } : {}),
      page,
      limit: PAGE_SIZE,
      ...(debouncedSearchParams.search
        ? { search: debouncedSearchParams.search }
        : {}),
      ...(debouncedSearchParams.species
        ? { species: debouncedSearchParams.species }
        : {}),
    },
    { enabled: hasSearchConditions, preservePreviousData: true },
  );

  // デバウンス待ち・前ページ保持中を確定済みとして扱うと、古い結果から患者を選ばせてしまう。
  const isBusy = Boolean(isLoading || isPlaceholderData || isSearchPending);

  // 応答の page / limit は外部境界の値であり、欠落・0・負値を素通しすると
  // 「1,204件中 -19-0件」や totalPages=Infinity といった破綻表示になる。
  // 正の整数でなければ要求値へ落として範囲計算を健全に保つ。
  const responsePage =
    typeof rawResponsePage === "number" && rawResponsePage > 0
      ? rawResponsePage
      : page;
  const responseLimit =
    typeof rawResponseLimit === "number" && rawResponseLimit > 0
      ? rawResponseLimit
      : PAGE_SIZE;

  const hasTotal = typeof total === "number";
  const totalCount = total ?? 0;
  const totalPages = Math.max(1, Math.ceil(totalCount / responseLimit));
  const startIndex =
    totalCount === 0 ? 0 : (responsePage - 1) * responseLimit + 1;
  const endIndex =
    totalCount === 0 ? 0 : Math.min(responsePage * responseLimit, totalCount);
  const rangeText = `${totalCount.toLocaleString()}件中 ${startIndex.toLocaleString()}-${endIndex.toLocaleString()}件`;

  // total と描画行が食い違う状態（応答に total が含まれない、範囲が総数を超える等）では
  // 件数を主張しない。「0件」と言いながら行を出すのは BUG-451 と同じ「嘘の件数」である。
  const isCountTrustworthy =
    hasTotal && (totalCount > 0 ? startIndex <= totalCount : pets.length === 0);
  // 取得失敗・取得中に現在行数を全体件数として提示しない。
  const showCachedRange =
    pets.length > 0 && Boolean(error || isBusy) && isCountTrustworthy;
  const statusText = !hasSearchConditions
    ? ""
    : showCachedRange
      ? `前回取得: ${rangeText}`
      : error || isBusy || !isCountTrustworthy
        ? ""
        : totalCount === 0
          ? "0件"
          : rangeText;
  // 多ページ時は Pagination が同じ範囲を可視表示するため、こちらは読み上げ専用にする。
  // 領域自体を出し入れすると AT が更新として扱わず、単一ページへ絞り込まれた検索が
  // 無通知になるため、領域は常設しテキストだけを差し替える（BUG-451 の a11y CRITICAL）。
  const isRangeShownByPagination = !error && totalPages > 1;

  const handlePageChange = useCallback(
    (nextPage: number) => {
      setPageBinding({
        page: Math.min(Math.max(1, nextPage), totalPages),
        params: debouncedSearchParams,
      });
    },
    [debouncedSearchParams, totalPages],
  );

  const handleTextFieldChange = useCallback(
    (key: "search" | "ownerId", value: string) => {
      setSearchParams((prev) => ({ ...prev, [key]: value }));
    },
    [],
  );

  const handleSpeciesChange = useCallback((species: string) => {
    setSearchParams((prev) => ({ ...prev, species }));
  }, []);

  // ページは debouncedSearchParams との対応で導出するため binding も初期値へ戻す。
  const handleClear = useCallback(() => {
    setSearchParams(INITIAL_SEARCH_PARAMS);
    setPageBinding({ page: 1, params: INITIAL_SEARCH_PARAMS });
  }, []);

  const selectedPetIds = useMemo(
    () => new Set(selectedPets.map((p) => p.id)),
    [selectedPets],
  );
  const isSelected = (pet: Pet) => selectedPetIds.has(pet.id);

  // 選択は親（usePetSelection）が Pet 実体で保持するためページに依存しない。
  // ただしページを移ると選択済みの行が画面から消えるため、現ページに無い選択の
  // 存在を明示する — 見えないだけで選択が外れたと誤解させない。
  //
  // ただし現ページの行と突き合わせられないうち（未検索・取得中で行が無い）は
  // 「別ページに在る」と断定しない。空の一覧を根拠に全選択を別ページ扱いすると、
  // 存在しないページを示唆する嘘の内訳になる。
  const currentPageIds = useMemo(() => new Set(pets.map((p) => p.id)), [pets]);
  const canLocateSelectionOnPage = hasSearchConditions && pets.length > 0;
  const offPageSelectedCount = useMemo(
    () =>
      canLocateSelectionOnPage
        ? selectedPets.reduce(
            (count, pet) => (currentPageIds.has(pet.id) ? count : count + 1),
            0,
          )
        : 0,
    [canLocateSelectionOnPage, selectedPets, currentPageIds],
  );
  const selectionText =
    selectedPets.length === 0
      ? ""
      : offPageSelectedCount > 0
        ? `選択中: ${selectedPets.length}件（別ページの${offPageSelectedCount}件を含む）`
        : `選択中: ${selectedPets.length}件`;

  return (
    <div className="flex flex-col gap-4 h-full">
      {/* Search Criteria */}
      <div
        className={`rounded-lg bg-white p-3 border ${C.borderMedium} shrink-0`}
      >
        {/* 入力停止後に自動検索するため、押しても何も起きない検索ボタンは置かない。 */}
        <p className={`mb-2 text-xs ${C.text60}`}>入力すると自動で検索します</p>
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-2 mb-3">
          {TEXT_FIELDS.map((field) => (
            <div key={field.key} className="space-y-0.5">
              <Label htmlFor={field.key} className={`text-xs ${C.text60}`}>
                {field.label}
              </Label>
              <Input
                id={field.key}
                placeholder={field.placeholder}
                value={searchParams[field.key]}
                onChange={(e) =>
                  handleTextFieldChange(field.key, e.target.value)
                }
                className={`text-xs h-11 bg-white ${C.borderMediumLight} ${C.text}`}
              />
            </div>
          ))}
          <SpeciesSelectField
            value={searchParams.species}
            onChange={handleSpeciesChange}
          />
        </div>
        {/* 条件付きでマウントしない — 押した瞬間にボタンが消えるとフォーカスが
            body へ落ちる（BUG-451 の a11y HIGH と同型）。条件が空のときの押下は
            no-op で害が無い。 */}
        <Button
          size="sm"
          variant="outline"
          onClick={handleClear}
          className={`h-11 min-w-11 text-sm ${C.borderMediumLight}`}
        >
          <RotateCcw className={`mr-1.5 ${ICON.xs}`} />
          クリア
        </Button>
      </div>

      {/* 選択状況と件数。どちらも領域を常設しテキストだけを差し替える。 */}
      <div className="flex items-center justify-between gap-2 shrink-0">
        <span
          role="status"
          aria-live="polite"
          aria-atomic="true"
          className={`text-xs ${C.text60}`}
        >
          {selectionText}
        </span>
        <span
          role="status"
          aria-live="polite"
          aria-atomic="true"
          className={
            isRangeShownByPagination ? "sr-only" : `text-xs ${C.text60}`
          }
        >
          {statusText}
        </span>
      </div>

      {/* Results Table */}
      <div
        className={`flex-1 rounded-lg bg-white overflow-hidden border ${C.borderMedium} min-h-0`}
      >
        <div className="overflow-auto h-full flex flex-col">
          {!hasSearchConditions && !isSearchPending ? (
            <div className="flex flex-col items-center justify-center flex-1 text-center gap-3">
              <Search className={`${ICON.xl} ${C.text20}`} />
              <div className={`text-sm ${C.text40}`}>
                検索条件を入力してください
              </div>
            </div>
          ) : error ? (
            // 取得失敗は「0件」より優先する。0件と誤表示すると
            // 存在する患者が存在しないことになり、臨床上危険な誤認を生む。
            <div className="flex-1 flex items-center justify-center">
              <ErrorFallback message="患者一覧の取得に失敗しました" />
            </div>
          ) : isBusy && pets.length === 0 ? (
            <div className="flex flex-col items-center justify-center flex-1 text-center gap-3">
              <div className="animate-spin">
                <Search className={`${ICON.xl} ${C.text20}`} />
              </div>
              <div className={`text-sm ${C.text40}`}>検索中...</div>
            </div>
          ) : pets.length === 0 ? (
            <div className="flex-1 flex items-center justify-center">
              <EmptyState
                icon={<SearchX className={ICON.xl} />}
                message="該当する患者が見つかりませんでした"
              />
            </div>
          ) : (
            <Table>
              <TableHeader className={`${C.bgPage} sticky top-0 z-10`}>
                <TableRow
                  className={`border-b ${C.borderMedium} h-9 ${C.hoverBgPage}`}
                >
                  <TableHead className={`min-w-[80px] ${C.text40} h-9`}>
                    飼主No
                  </TableHead>
                  <TableHead className={`min-w-[120px] ${C.text40} h-9`}>
                    飼主名
                  </TableHead>
                  <TableHead className={`min-w-[100px] ${C.text40} h-9`}>
                    ペット名
                  </TableHead>
                  <TableHead className={`min-w-[60px] ${C.text40} h-9`}>
                    種別
                  </TableHead>
                  <TableHead className={`min-w-[60px] ${C.text40} h-9`}>
                    性別
                  </TableHead>
                  <TableHead className={`min-w-[80px] ${C.text40} h-9`}>
                    生年月日
                  </TableHead>
                  <TableHead className={`min-w-[60px] ${C.text40} h-9`}>
                    体重
                  </TableHead>
                  <TableHead className={`min-w-[60px] ${C.text40} h-9`}>
                    操作
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {pets.map((pet) => {
                  const isDeceased = pet.status === "死亡";
                  const isAlive = pet.status === "生存";
                  // ページ送り・デバウンス待ちの間も行は消さない（消すと preservePreviousData が
                  // 無意味になり「前回取得: N件中 1-20件」が空の表の横に出る自己矛盾になる）。
                  // ただし表示中の行は「まだ古い条件の結果」なので選択はさせない。
                  const isSelectable = isAlive && !isBusy;
                  return (
                    <TableRow
                      key={pet.id}
                      className={`transition-colors ${C.hoverBgLight} h-12 ${
                        isSelected(pet) ? C.bgPage : ""
                      } ${!isAlive ? "opacity-50 grayscale-[0.5]" : ""}`}
                    >
                      <TableCell className={`text-sm font-mono ${C.text}`}>
                        {pet.ownerId}
                      </TableCell>
                      <TableCell className={`text-sm font-medium ${C.text}`}>
                        {pet.ownerName}
                      </TableCell>
                      <TableCell className={`text-sm ${C.text}`}>
                        <span className="font-bold">{pet.name}</span>
                      </TableCell>
                      <TableCell className={`text-sm ${C.text}`}>
                        {pet.species}
                      </TableCell>
                      <TableCell className={`text-sm ${C.text}`}>
                        {pet.gender || "-"}
                      </TableCell>
                      <TableCell className={`text-sm font-mono ${C.text}`}>
                        {pet.birthDate || "-"}
                      </TableCell>
                      <TableCell className={`text-sm font-mono ${C.text}`}>
                        {pet.weight || "-"}
                      </TableCell>
                      <TableCell>
                        <Button
                          size="sm"
                          disabled={!isSelectable}
                          aria-label={
                            isDeceased
                              ? `死亡・選択不可: ${pet.name} (ID ${pet.id})`
                              : isBusy
                                ? `読み込み中・選択不可: ${pet.name} (ID ${pet.id})`
                                : !isAlive
                                  ? `状態不明・選択不可: ${pet.name} (ID ${pet.id})`
                                  : isSelected(pet)
                                    ? `選択中: ${pet.name} (ID ${pet.id})`
                                    : `選択: ${pet.name} (ID ${pet.id})`
                          }
                          className={`h-11 min-w-11 gap-1 text-sm px-2 transition-colors ${
                            !isSelectable
                              ? `${C.bgPage} ${C.textStatusGray} border-transparent cursor-not-allowed`
                              : isSelected(pet)
                                ? `${C.bgBrand} ${C.textOnBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand}`
                                : `bg-white border ${C.borderMediumLight} ${C.text} ${C.hoverBgSubtle}`
                          }`}
                          onClick={() => {
                            if (isSelectable) onSelect(pet);
                          }}
                        >
                          <Check
                            className={`${ICON.xs} ${isSelected(pet) ? "" : "opacity-0"}`}
                          />
                          {isDeceased
                            ? "死亡"
                            : !isAlive
                              ? "不明"
                              : isSelected(pet)
                                ? "選択中"
                                : "選択"}
                        </Button>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}
        </div>
      </div>

      {/* 操作可能なコントロールを live region へ入れない（全ボタンが読み上げられる）。
          読み込み中も disabled にしない — フォーカス中のボタンが無効化されると
          ブラウザが強制 blur し、ページ送りのたびにキーボード焦点が失われる。 */}
      {!error && totalPages > 1 ? (
        <fieldset aria-busy={isBusy} className="border-0 p-0 m-0 shrink-0">
          <Pagination
            currentPage={responsePage}
            totalPages={totalPages}
            totalCount={totalCount}
            startIndex={startIndex}
            endIndex={endIndex}
            onPageChange={handlePageChange}
            onPrev={() => handlePageChange(responsePage - 1)}
            onNext={() => handlePageChange(responsePage + 1)}
          />
        </fieldset>
      ) : null}
    </div>
  );
});
