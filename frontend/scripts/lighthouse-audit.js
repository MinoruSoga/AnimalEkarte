#!/usr/bin/env node

/**
 * Lighthouse Audit Runner
 *
 * フロントエンドのパフォーマンス・アクセシビリティ・ベストプラクティスを測定
 *
 * 実行:
 * $ node frontend/scripts/lighthouse-audit.js
 * $ node frontend/scripts/lighthouse-audit.js --url http://localhost:3000
 */

import lighthouse from "lighthouse";
import * as chromeLauncher from "chrome-launcher";
import fs from "node:fs";
import path from "node:path";

const BASE_URL =
  process.argv.find((arg) => arg.startsWith("--url"))?.split("=")[1] || "http://localhost:3000";
const OUTPUT_DIR = "frontend/audit-results";
const JST_TIME_ZONE = "Asia/Tokyo";

function formatJSTTimestamp(date = new Date()) {
  const parts = new Intl.DateTimeFormat("ja-JP", {
    timeZone: JST_TIME_ZONE,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  }).formatToParts(date);
  const values = Object.fromEntries(parts.map((part) => [part.type, part.value]));
  return `${values.year}-${values.month}-${values.day}T${values.hour}:${values.minute}:${values.second}+09:00`;
}

async function launchChrome() {
  const chrome = await chromeLauncher.launch({ chromeFlags: ["--headless"] });
  return chrome;
}

async function runLighthouse(url, chrome) {
  const options = {
    logLevel: "info",
    output: "json",
    port: chrome.port,
    onlyCategories: ["performance", "accessibility", "best-practices", "seo"],
  };

  const runnerResult = await lighthouse(url, options);

  // Chrome を閉じる
  await chromeLauncher.kill(chrome.pid);

  return runnerResult;
}

function generateReport(results) {
  const categories = results.lhr.categories;
  const report = {
    url: results.lhr.fetchTime,
    timestamp: formatJSTTimestamp(),
    categories: {},
    metrics: {},
  };

  // カテゴリスコア
  for (const [name, category] of Object.entries(categories)) {
    report.categories[name] = {
      score: Math.round(category.score * 100),
      title: category.title,
    };
  }

  // Core Web Vitals
  const metrics = results.lhr.audits;
  report.metrics = {
    firstContentfulPaint: metrics["first-contentful-paint"]?.numericValue || null,
    largestContentfulPaint: metrics["largest-contentful-paint"]?.numericValue || null,
    cumulativeLayoutShift: metrics["cumulative-layout-shift"]?.numericValue || null,
    interactive: metrics["interactive"]?.numericValue || null,
  };

  return report;
}

function displayReport(report) {
  console.log("\n=== Lighthouse Audit Results ===\n");

  console.log("Categories:");
  for (const [name, data] of Object.entries(report.categories)) {
    const score = data.score;
    const status = score >= 90 ? "✅" : score >= 50 ? "⚠️" : "❌";
    console.log(`  ${status} ${data.title}: ${score}/100`);
  }

  console.log("\nCore Web Vitals:");
  console.log(`  First Contentful Paint: ${report.metrics.firstContentfulPaint?.toFixed(2)}ms`);
  console.log(`  Largest Contentful Paint: ${report.metrics.largestContentfulPaint?.toFixed(2)}ms`);
  console.log(`  Cumulative Layout Shift: ${report.metrics.cumulativeLayoutShift?.toFixed(3)}`);
  console.log(`  Time to Interactive: ${report.metrics.interactive?.toFixed(2)}ms`);

  console.log(`\nTimestamp: ${report.timestamp}`);
}

async function main() {
  try {
    // 出力ディレクトリを作成
    if (!fs.existsSync(OUTPUT_DIR)) {
      fs.mkdirSync(OUTPUT_DIR, { recursive: true });
    }

    console.log(`🚀 Starting Lighthouse audit for ${BASE_URL}...`);

    const chrome = await launchChrome();
    const results = await runLighthouse(BASE_URL, chrome);
    const report = generateReport(results);

    // ファイルに保存
    const timestamp = formatJSTTimestamp().replace(/[:+]/g, "-");
    const reportPath = path.join(OUTPUT_DIR, `lighthouse-${timestamp}.json`);
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));

    // レポートを表示
    displayReport(report);
    console.log(`\n✅ Report saved to ${reportPath}`);

    // 結果がしきい値を超えたかチェック
    const performance = report.categories.performance.score;
    if (performance < 75) {
      console.warn("\n⚠️ Performance score is below target (75). Consider optimization.");
      process.exit(1);
    }
  } catch (error) {
    console.error("Error running Lighthouse audit:", error);
    process.exit(1);
  }
}

main().catch((err) => {
  console.error("Fatal error:", err);
  process.exit(1);
});
