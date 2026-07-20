// React/Framework
import { useState, memo } from "react";

// External
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from "recharts";

// Internal
import { C, ICON, PALETTE } from "@/lib/design-tokens";

// Relative
import type { Vital } from "../../types";

// ── 表示可能な指標定義 ────────────────────────────────────────────────

type MetricKey = "temperature" | "heart_rate" | "respiration_rate" | "weight";

interface MetricDef {
  key: MetricKey;
  label: string;
  unit: string;
  color: string;
  yAxisId: "left" | "right";
  domain?: [number | "auto", number | "auto"];
}

const METRICS: MetricDef[] = [
  { key: "temperature",      label: "体温",   unit: "℃",    color: PALETTE.chartTemperature, yAxisId: "left",  domain: [30, 45] },
  { key: "heart_rate",       label: "心拍数", unit: "bpm",  color: PALETTE.accent,             yAxisId: "right", domain: [0, "auto"] },
  { key: "respiration_rate", label: "呼吸数", unit: "/min", color: PALETTE.chartRespiratory,   yAxisId: "right", domain: [0, "auto"] },
  { key: "weight",      label: "体重",   unit: "",     color: PALETTE.chartWeight,        yAxisId: "left",  domain: [0, "auto"] },
];

// ── ユーティリティ ────────────────────────────────────────────────────

/** ISO8601 → "MM/DD HH:mm" 短縮表示 */
function formatXAxis(iso: string): string {
  const d = new Date(iso);
  if (isNaN(d.getTime())) return iso;
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${pad(d.getMonth() + 1)}/${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

// ── Tooltip カスタム ──────────────────────────────────────────────────

interface TooltipPayloadItem {
  name: string;
  value: number;
  color: string;
  dataKey: string;
}

interface CustomTooltipProps {
  active?: boolean;
  payload?: TooltipPayloadItem[];
  label?: string;
}

const CustomTooltip = memo(function CustomTooltip({ active, payload, label }: CustomTooltipProps) {
  if (!active || !payload?.length) return null;
  const metric = METRICS.find((m) => m.key === payload[0]?.dataKey);
  return (
    <div className={`${C.bgWhite} border ${C.borderMedium} rounded-xs shadow-level1 px-3 py-2 text-sm`}>
      <p className={`${C.text60} text-xs mb-1`}>{label}</p>
      {payload.map((item) => {
        const def = METRICS.find((m) => m.key === item.dataKey);
        return (
          <p key={item.dataKey} style={{ color: item.color }} className="font-medium">
            {def?.label ?? item.name}: {item.value}{def?.unit ?? ""}
          </p>
        );
      })}
      {metric ? null : null}
    </div>
  );
});

// ── Props ─────────────────────────────────────────────────────────────

interface VitalsGraphProps {
  vitals: Vital[];
}

// ── Component ─────────────────────────────────────────────────────────

export const VitalsGraph = memo(function VitalsGraph({ vitals }: VitalsGraphProps) {
  // 初期状態: 体温と体重を表示
  const [activeMetrics, setActiveMetrics] = useState<Set<MetricKey>>(
    () => new Set<MetricKey>(["temperature", "weight"])
  );

  const toggleMetric = (key: MetricKey) => {
    setActiveMetrics((prev) => {
      const next = new Set(prev);
      if (next.has(key)) {
        // 最低1つは残す
        if (next.size > 1) next.delete(key);
      } else {
        next.add(key);
      }
      return next;
    });
  };

  // recharts 用データ変換
  const chartData = vitals.map((v) => ({
    recorded_at:      formatXAxis(v.recorded_at),
    temperature:      v.temperature ?? undefined,
    heart_rate:       v.heart_rate ?? undefined,
    respiration_rate: v.respiration_rate ?? undefined,
    weight:      v.weight ?? undefined,
  }));

  // 右 Y 軸に表示する指標があるか
  const hasRight = METRICS.some(
    (m) => m.yAxisId === "right" && activeMetrics.has(m.key)
  );

  return (
    <div className={`border ${C.borderLight} rounded-xs ${C.bgWhite} p-4 flex flex-col gap-3`}>
      {/* 指標切り替えボタン */}
      <div className="flex flex-wrap gap-2">
        {METRICS.map((m) => {
          const active = activeMetrics.has(m.key);
          return (
            <button
              key={m.key}
              type="button"
              onClick={() => toggleMetric(m.key)}
              className={[
                "flex items-center gap-1.5 px-3 h-7 rounded-full text-xs font-medium border transition-colors",
                active
                  ? `${C.textWhite} border-transparent`
                  : `${C.text60} ${C.borderLight} ${C.bgWhite} ${C.hoverBgLight}`,
              ].join(" ")}
              style={active ? { backgroundColor: m.color, borderColor: m.color } : {}}
            >
              <span
                className={`${ICON.dot} rounded-full shrink-0`}
                style={{ backgroundColor: active ? PALETTE.whiteAlpha80 : m.color }}
              />
              {m.label}
              {m.unit ? <span className="opacity-70">({m.unit})</span> : null}
            </button>
          );
        })}
      </div>

      {/* グラフ本体 */}
      {chartData.length < 2 ? (
        <div className={`flex items-center justify-center h-40 text-sm ${C.text40}`}>
          グラフを表示するには2件以上のバイタル記録が必要です
        </div>
      ) : (
        <ResponsiveContainer width="100%" height={260}>
          <LineChart data={chartData} margin={{ top: 4, right: 20, left: 0, bottom: 4 }}>
            <CartesianGrid strokeDasharray="3 3" stroke={PALETTE.chartGrid} />
            <XAxis
              dataKey="recorded_at"
              tick={{ fontSize: 11, fill: PALETTE.chartAxisText }}
              tickLine={false}
              axisLine={{ stroke: PALETTE.chartGrid }}
            />
            {/* 左 Y 軸（体温・体重） */}
            <YAxis
              yAxisId="left"
              tick={{ fontSize: 11, fill: PALETTE.chartAxisText }}
              tickLine={false}
              axisLine={{ stroke: PALETTE.chartGrid }}
              width={36}
            />
            {/* 右 Y 軸（心拍数・呼吸数） */}
            {hasRight ? (
              <YAxis
                yAxisId="right"
                orientation="right"
                tick={{ fontSize: 11, fill: PALETTE.chartAxisText }}
                tickLine={false}
                axisLine={{ stroke: PALETTE.chartGrid }}
                width={36}
              />
            ) : null}
            <Tooltip content={<CustomTooltip />} />
            <Legend
              iconType="circle"
              iconSize={8}
              wrapperStyle={{ fontSize: 12, paddingTop: 8 }}
              formatter={(value: string) => {
                const def = METRICS.find((m) => m.key === value);
                return def ? `${def.label}${def.unit ? ` (${def.unit})` : ""}` : value;
              }}
            />
            {METRICS.flatMap((m) =>
              activeMetrics.has(m.key) ? [
                <Line
                  key={m.key}
                  yAxisId={m.yAxisId}
                  type="monotone"
                  dataKey={m.key}
                  name={m.key}
                  stroke={m.color}
                  strokeWidth={2}
                  dot={{ r: 4, fill: m.color, strokeWidth: 0 }}
                  activeDot={{ r: 6, fill: m.color }}
                  connectNulls
                />
              ] : []
            )}
          </LineChart>
        </ResponsiveContainer>
      )}
    </div>
  );
});
