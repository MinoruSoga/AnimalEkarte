#!/usr/bin/env node
'use strict';

const { spawnSync } = require('child_process');
const path = require('path');
const fs = require('fs');

const DEFAULT_THRESHOLD = 75;

function usage() {
  console.log(`
Usage: node scripts/harness-health.js [options]

Generate an actionable harness health summary.

Options:
  --scope <repo|hooks|skills|commands|agents>  Scope to audit (default: repo)
  --root <path>                               Root directory to audit (default: cwd)
  --format <text|json>                        Output format (default: text)
  --threshold <n>                             Blocker threshold (default: ${DEFAULT_THRESHOLD})
  --owner <name>                              Default owner for blockers when absent (default: unassigned)
  --due-days <n>                              Default due-days for generated blockers (default: 3)
  --output <path>                             Write JSON output to file (json mode only)
  --exit-code                                  Exit with code 2 when status is BLOCKED
  --json                                       Alias for --format json
  --help, -h                                  Show help
`);
  process.exit(0);
}

function readValue(args, index, name) {
  const value = args[index + 1];
  if (!value || value.startsWith('--')) {
    throw new Error(`${name} requires a value`);
  }

  return value;
}

function parseArgs(argv) {
  const args = argv.slice(2);
  const parsed = {
    scope: 'repo',
    root: path.resolve(process.cwd()),
    format: 'text',
    threshold: DEFAULT_THRESHOLD,
    owner: 'unassigned',
    dueDays: 3,
    output: '',
    exitCode: false,
    help: false,
  };

  for (let index = 0; index < args.length; index += 1) {
    const arg = args[index];

    if (arg === '--help' || arg === '-h') {
      parsed.help = true;
      continue;
    }

    if (arg === '--json') {
      parsed.format = 'json';
      continue;
    }

    if (arg === '--scope') {
      parsed.scope = readValue(args, index, '--scope');
      index += 1;
      continue;
    }

    if (arg === '--root') {
      parsed.root = path.resolve(readValue(args, index, '--root'));
      index += 1;
      continue;
    }

    if (arg === '--format') {
      parsed.format = readValue(args, index, '--format').toLowerCase();
      index += 1;
      continue;
    }

    if (arg === '--threshold') {
      const value = Number(readValue(args, index, '--threshold'));
      if (!Number.isFinite(value) || value < 0 || value > 100) {
        throw new Error('--threshold must be a number between 0 and 100');
      }
      parsed.threshold = value;
      index += 1;
      continue;
    }

    if (arg === '--owner') {
      parsed.owner = readValue(args, index, '--owner');
      index += 1;
      continue;
    }

    if (arg === '--due-days') {
      const value = Number(readValue(args, index, '--due-days'));
      if (!Number.isFinite(value) || value < 0) {
        throw new Error('--due-days must be 0 or greater');
      }
      parsed.dueDays = value;
      index += 1;
      continue;
    }

    if (arg === '--output') {
      parsed.output = path.resolve(readValue(args, index, '--output'));
      index += 1;
      continue;
    }

    if (arg === '--exit-code') {
      parsed.exitCode = true;
      continue;
    }

    if (arg.startsWith('--scope=')) {
      parsed.scope = arg.slice('--scope='.length);
      continue;
    }

    if (arg.startsWith('--root=')) {
      parsed.root = path.resolve(arg.slice('--root='.length));
      continue;
    }

    if (arg.startsWith('--format=')) {
      parsed.format = arg.slice('--format='.length).toLowerCase();
      continue;
    }

    if (arg.startsWith('--threshold=')) {
      const value = Number(arg.slice('--threshold='.length));
      if (!Number.isFinite(value) || value < 0 || value > 100) {
        throw new Error('--threshold must be a number between 0 and 100');
      }
      parsed.threshold = value;
      continue;
    }

    if (arg.startsWith('--owner=')) {
      parsed.owner = arg.slice('--owner='.length);
      continue;
    }

    if (arg.startsWith('--due-days=')) {
      const value = Number(arg.slice('--due-days='.length));
      if (!Number.isFinite(value) || value < 0) {
        throw new Error('--due-days must be 0 or greater');
      }
      parsed.dueDays = value;
      continue;
    }

    if (arg.startsWith('--output=')) {
      parsed.output = path.resolve(arg.slice('--output='.length));
      continue;
    }

    throw new Error(`Unknown argument: ${arg}`);
  }

  if (!['repo', 'hooks', 'skills', 'commands', 'agents'].includes(parsed.scope)) {
    throw new Error(`Invalid scope: ${parsed.scope}`);
  }

  if (!['text', 'json'].includes(parsed.format)) {
    throw new Error(`Invalid format: ${parsed.format}. Use text or json.`);
  }

  return parsed;
}

function nowDateOffset(days) {
  const base = new Date();
  const date = new Date(base.getTime() + (days * 24 * 60 * 60 * 1000));
  return date.toISOString().slice(0, 10);
}

function runHarnessAudit(root, scope) {
  const script = path.join(__dirname, 'harness-audit.js');
  const result = spawnSync('node', [script, scope, '--format', 'json', '--root', root], {
    encoding: 'utf8',
    stdio: ['ignore', 'pipe', 'pipe'],
  });

  if (result.status !== 0) {
    const message = (result.stderr || '').trim() || (result.stdout || '').trim() || 'Unknown failure';
    throw new Error(`harness-audit failed: ${message}`);
  }

  try {
    return JSON.parse(result.stdout || '{}');
  } catch (error) {
    throw new Error(`Failed to parse harness-audit JSON output: ${error.message}`);
  }
}

function buildHarnessReport(auditReport, options) {
  const blocked = auditReport.overall_score < options.threshold;
  const failed = auditReport.checks.filter(check => !check.pass);
  const dueDate = nowDateOffset(options.dueDays);

  return {
    harness: 'daily-harness-health',
    generated_at: new Date().toISOString(),
    scope: auditReport.scope,
    root_dir: auditReport.root_dir,
    threshold: options.threshold,
    score: auditReport.overall_score,
    max_score: auditReport.max_score,
    category_count: auditReport.category_count,
    applicable_categories: auditReport.applicable_categories,
    failed_checks_count: failed.length,
    status: blocked ? 'BLOCKED' : 'OK',
    top_actions: auditReport.top_actions,
    blockers: auditReport.top_actions.slice(0, 3).map((action, index) => ({
      id: `blocker-${index + 1}`,
      owner: options.owner,
      due: dueDate,
      category: action.category,
      path: action.path,
      action: action.action,
      points: action.points,
    })),
  };
}

function formatText(report) {
  const lines = [];

  lines.push('Harness Health Report');
  lines.push(`Scope: ${report.scope}`);
  lines.push(`Root: ${report.root_dir}`);
  lines.push(`Score: ${report.score}/${report.max_score} (threshold ${report.threshold})`);
  lines.push(`Applicable categories: ${report.category_count}`);
  lines.push(`Failed checks: ${report.failed_checks_count}`);
  lines.push(`Status: ${report.status}`);
  lines.push('');

  if (report.top_actions.length > 0) {
    lines.push('Top 3 Actions:');
    report.top_actions.forEach((action, index) => {
      lines.push(`${index + 1}) [${action.category}] ${action.action} (${action.path})`);
    });
    lines.push('');
  }

  if (report.status === 'BLOCKED') {
    lines.push('Action Blockers:');
    report.blockers.forEach((blocker) => {
      lines.push(`- ${blocker.id}: owner=${blocker.owner} | due=${blocker.due} | action=${blocker.action}`);
    });
  } else {
    lines.push('Blockers: none');
  }

  return `${lines.join('\n')}\n`;
}

function main() {
  try {
    const options = parseArgs(process.argv);

    if (options.help) {
      usage();
    }

    const auditReport = runHarnessAudit(options.root, options.scope);
    const report = buildHarnessReport(auditReport, options);

    if (options.format === 'json') {
      const payload = JSON.stringify(report, null, 2);
      process.stdout.write(`${payload}\n`);

      if (options.output) {
        fs.writeFileSync(options.output, `${payload}\n`, 'utf8');
      }
      if (options.exitCode && report.status === 'BLOCKED') {
        process.exit(2);
      }
      return;
    }

    process.stdout.write(formatText(report));

    if (options.exitCode && report.status === 'BLOCKED') {
      process.exit(2);
    }
  } catch (error) {
    process.stderr.write(`Error: ${error.message}\n`);
    process.exit(1);
  }
}

main();
