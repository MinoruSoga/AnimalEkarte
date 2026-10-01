// React/Framework
import { useCallback, useContext, useRef, useState } from "react";
import { QueryClientContext } from "@tanstack/react-query";

// External
import { toast } from "sonner";

// Internal
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import {
  transformConsultation,
  transformProcedure,
  type ConsultationItem,
  type ProcedureItem,
} from "@/lib/transforms/treatment";
import { transformBackendMedicineToFrontend, type Medicine } from "@/lib/transforms/medicine";

// Relative
import { createTreatment, getTreatments } from "../api/treatments";
import type { Treatment } from "../types";
import {
  buildCopiedTreatmentPayload,
  type CopyTreatmentMasters,
} from "../lib/treatments-tab-model";

/**
 * EMR-219（スプシ44 残作業）: 問診履歴「コピー」で元カルテの治療明細行も複製する。
 *
 * - 単価は当時価格を持ち込まず、現行マスタ一覧（/v1/masters/*）の価格へ再解決する。
 *   masters の query key と transform は src/hooks/use-treatment-master.ts の
 *   useGetAllConsultations / useGetAllProcedures / useGetAllMedicinesMaster と同一で、
 *   STATIC staleTime の共有キャッシュをそのまま利用する（lazy fetch のため hook は
 *   mount せず、適用時に fetchQuery で同一キーを引く）。
 * - wire 型は @/types/generated/models の新規 import 禁止（TASK-444-S1）に抵触しないよう
 *   `Parameters<typeof transform>[0]` から引き出す。fetcher はフル shape の transform を
 *   通すため、キャッシュ shape が shared hook の応答と常に一致する。
 * - QueryClientProvider 無しでも安全に呼べるよう useQueryClient / useMutation /
 *   useCreateTreatment は使わず、QueryClientContext を直接引く（provider 不在時は
 *   undefined → コールバック内で早期 return）。作成は同ファイルの createTreatment
 *   fetcher（useCreateTreatment と同一 request shape）をソース sort_order 昇順で
 *   逐次 POST する。create 失敗時は useCreateTreatment.onError と同じ
 *   handleApiError('治療追加') を1回出して中断し、作成済み行は残る（rollback API 無し）。
 *   成功分がある場合は onSuccess 相当の treatments 一覧 invalidate を完了時に1回だけ行う。
 * - 現カルテに id が無い（自動作成前）ときは明細複写をスキップする（問診項目の複写は
 *   呼び出し側で既に適用済み）。
 */

// use-treatment-master.ts の useGetAll*Master と同一 endpoint/transform/key。
type ConsultationWire = Parameters<typeof transformConsultation>[0];
const fetchConsultationsOnce = async (): Promise<ConsultationItem[]> => {
  const { data } = await axios.get<ConsultationWire[]>("/v1/masters/consultations");
  return data.map(transformConsultation);
};

type ProcedureWire = Parameters<typeof transformProcedure>[0];
const fetchProceduresOnce = async (): Promise<ProcedureItem[]> => {
  const { data } = await axios.get<ProcedureWire[]>("/v1/masters/procedures");
  return data.map(transformProcedure);
};

type MedicineWire = Parameters<typeof transformBackendMedicineToFrontend>[0];
const fetchMedicinesOnce = async (): Promise<Medicine[]> => {
  const { data } = await axios.get<MedicineWire[]>("/v1/masters/medicines");
  return data.map(transformBackendMedicineToFrontend);
};

export function useCopyTreatmentDetails(
  medicalRecordId: string | undefined,
  recordClinicId?: string,
) {
  // QueryClientProvider 不在でも throw しないよう context を直接引く
  // （provider が無い描画経路では undefined → コールバック内で早期 return）。
  const queryClient = useContext(QueryClientContext);
  const [isPending, setIsPending] = useState(false);
  // 連打で二重複写しないための実行中ガード（isPending は render 時点 snapshot のため ref）。
  const runningRef = useRef(false);

  const copyTreatmentsFromRecord = useCallback(
    async (sourceRecordId: string): Promise<void> => {
      if (!medicalRecordId || !queryClient || runningRef.current) return;
      runningRef.current = true;
      setIsPending(true);
      try {
        let sourceRows: Treatment[];
        let destRows: Treatment[];
        let masters: CopyTreatmentMasters;
        try {
          // 複写元は履歴一覧（グローバルクリニック文脈）と同じ文脈で引くため clinicId を
          // 明示しない。複写先は拠点横断対応済みの recordClinicId を使う（P2-15 準拠）。
          const [source, dest, consultations, procedures, medicines] = await Promise.all([
            queryClient.fetchQuery({
              queryKey: queryKeys.medicalRecords.treatments(sourceRecordId),
              queryFn: () => getTreatments(sourceRecordId),
              staleTime: QUERY_STALE_TIMES.REALTIME,
            }),
            queryClient.fetchQuery({
              queryKey: queryKeys.medicalRecords.treatments(medicalRecordId, recordClinicId),
              queryFn: () => getTreatments(medicalRecordId, recordClinicId),
              staleTime: QUERY_STALE_TIMES.REALTIME,
            }),
            queryClient.fetchQuery({
              queryKey: queryKeys.masters.category("consultations"),
              queryFn: fetchConsultationsOnce,
              staleTime: QUERY_STALE_TIMES.STATIC,
            }),
            queryClient.fetchQuery({
              queryKey: queryKeys.masters.category("procedures"),
              queryFn: fetchProceduresOnce,
              staleTime: QUERY_STALE_TIMES.STATIC,
            }),
            queryClient.fetchQuery({
              queryKey: queryKeys.masters.category("medicines"),
              queryFn: fetchMedicinesOnce,
              staleTime: QUERY_STALE_TIMES.STATIC,
            }),
          ]);
          sourceRows = source;
          destRows = dest;
          masters = { consultations, procedures, medicines };
        } catch (err) {
          handleApiError(err, "治療明細の複写");
          return;
        }

        const sortedSource = [...sourceRows].sort((a, b) => a.sort_order - b.sort_order);
        if (sortedSource.length === 0) {
          toast.info("複写する治療明細はありませんでした");
          return;
        }

        const sortOrderOffset =
          destRows.reduce((max, row) => Math.max(max, row.sort_order), -1) + 1;

        let copied = 0;
        let createFailed = false;
        let createFailure: unknown;
        for (const row of sortedSource) {
          try {
            await createTreatment(
              medicalRecordId,
              buildCopiedTreatmentPayload({ source: row, masters, sortOrderOffset }),
              recordClinicId,
            );
            copied += 1;
          } catch (err) {
            createFailed = true;
            createFailure = err;
            break;
          }
        }

        if (copied > 0) {
          // useCreateTreatment.onSuccess 相当の一覧無効化を完了時に1回だけ行う
          // （行毎 invalidate による再 fetch storm 防止）。途中失敗で部分作成された
          // 場合も作成済み行を一覧へ反映するため同じく1回実行する。
          await queryClient.invalidateQueries({
            queryKey: queryKeys.medicalRecords.treatments(medicalRecordId),
          });
        }
        if (createFailed) {
          // useCreateTreatment.onError（handleApiError '治療追加'）と同じ通知を1回だけ出す。
          handleApiError(createFailure, "治療追加");
          return;
        }
        toast.success(`前回の治療明細を ${copied.toLocaleString()} 件コピーしました`);
      } catch (err) {
        // payload 構築等の想定外失敗（create 経路の失敗は上で処理済み）。
        handleApiError(err, "治療明細の複写");
      } finally {
        runningRef.current = false;
        setIsPending(false);
      }
    },
    [medicalRecordId, queryClient, recordClinicId],
  );

  return { copyTreatmentsFromRecord, isPending };
}
