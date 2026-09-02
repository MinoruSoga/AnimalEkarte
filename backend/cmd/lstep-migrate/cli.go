package main

import (
	"flag"
	"fmt"
	"os"
)

type migrateCLI struct {
	clinicID      uint64
	dryRun        bool
	batchSize     int
	rateLimitPerS int
	skipTier2     bool
	ownerIDs      []uint64
	resumeFrom    uint64
	reportPath    string
}

func parseMigrateCLI() (migrateCLI, int) {
	var (
		clinicID      = flag.Uint64("clinic-id", 0, "対象クリニックID（必須）")
		dryRun        = flag.Bool("dry-run", false, "ドライラン（DB 書き込みなし）")
		batchSize     = flag.Int("batch-size", 5, "並列実行数")
		rateLimitPerS = flag.Int("rate-limit-per-sec", 10, "1秒あたりの最大同期数")
		skipTier2     = flag.Bool("skip-tier-2", false, "Tier2 同期スキップ")
		ownerIDsFlag  = flag.String("owner-ids", "", "対象 ownerID のカンマ区切りリスト（省略時は全員）")
		resumeFrom    = flag.Uint64("resume-from", 0, "このID以降の飼い主のみ処理")
		reportPath    = flag.String("report", "lstep_migration_report.csv", "CSVレポート出力先")
	)
	flag.Parse()

	if *clinicID == 0 {
		fmt.Fprintln(os.Stderr, "error: --clinic-id は必須です")
		flag.Usage()
		return migrateCLI{}, 1
	}
	if *batchSize < 1 {
		fmt.Fprintln(os.Stderr, "error: --batch-size は1以上を指定してください")
		return migrateCLI{}, 1
	}
	if *rateLimitPerS < 1 {
		fmt.Fprintln(os.Stderr, "error: --rate-limit-per-sec は1以上を指定してください")
		return migrateCLI{}, 1
	}

	ownerIDs, err := parseOwnerIDs(*ownerIDsFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: --owner-ids の解析に失敗しました: %v\n", err)
		return migrateCLI{}, 1
	}
	return migrateCLI{
		clinicID:      *clinicID,
		dryRun:        *dryRun,
		batchSize:     *batchSize,
		rateLimitPerS: *rateLimitPerS,
		skipTier2:     *skipTier2,
		ownerIDs:      ownerIDs,
		resumeFrom:    *resumeFrom,
		reportPath:    *reportPath,
	}, 0
}
