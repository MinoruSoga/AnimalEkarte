#!/usr/bin/env node
'use strict';

const { spawnSync } = require('child_process');
const path = require('path');

const DEFAULT_SCOPE = 'repo';
const DEFAULT_THRESHOLD = 75;
const DEFAULT_MAX_ITERATIONS = 5;
const DEFAULT_DUE_DAYS = 2;

function usage() {
  console.log(`
Usage: node scripts/harness-refinement-loop.js [options]

Run harness refinement loops and generate operator-level action prompts per iteration.

Options:
  --scope <repo|hooks|skills|commands|agents>   Scope to audit (default: ${DEFAULT_SCOPE})
  --threshold <n>                               Score threshold for completion (default: ${DEFAULT_THRESHOLD})
  --max-iterations <n>                          Maximum iterations (default: ${DEFAULT_MAX_ITERATIONS})
  --root <path>                                 Repository root to inspect (default: cwd)
  --owner <name>                                Owner in generated blockers (default: unassigned)
  --due-days <n>                                Due days for generated blockers (default: ${DEFAULT_DUE_DAYS})
  --format <text|json>                          Output format (default: text)
  --help, -h                                    Show help
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
    scope: DEFAULT_SCOPE,
    threshold: DEFAULT_THRESHOLD,
    maxIterations: DEFAULT_MAX_ITERATIONS,
    root: path.resolve(process.cwd()),
    owner: 'unassigned',
    dueDays: DEFAULT_DUE_DAYS,
    format: 'text',
    help: false,
  };

  for (let index = 0; index < args.length; index += 1) {
    const arg = args[index];

    if (arg === '--help' || arg === '-h') {
      parsed.help = true;
      continue;
    }

    if (arg === '--scope') {
      parsed.scope = readValue(args, index, '--scope');
      index += 1;
      continue;
    }

    if (arg === '--threshold') {
      const value = Number(readValue(args, index, '--threshold'));
      if (!Number.isFinite(value) || value < 0 || value > 100) {
        throw new Error('--threshold must be 0-100');
      }
      parsed.threshold = value;
      index += 1;
      continue;
    }

    if (arg === '--max-iterations') {
      const value = Number(readValue(args, index, '--max-iterations'));
      if (!Number.isFinite(value) || value < 1 || value > 30) {
        throw new Error('--max-iterations must be 1-30');
      }
      parsed.maxIterations = Math.trunc(value);
      index += 1;
      continue;
    }

    if (arg === '--root') {
      parsed.root = path.resolve(readValue(args, index, '--root'));
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

    if (arg === '--format') {
      parsed.format = readValue(args, index, '--format').toLowerCase();
      index += 1;
      continue;
    }

    if (arg.startsWith('--scope=')) {
      parsed.scope = arg.slice('--scope='.length);
      continue;
    }

    if (arg.startsWith('--threshold=')) {
      const value = Number(arg.slice('--threshold='.length));
      if (!Number.isFinite(value) || value < 0 || value > 100) {
        throw new Error('--threshold must be 0-100');
      }
      parsed.threshold = value;
      continue;
    }

    if (arg.startsWith('--max-iterations=')) {
      const value = Number(arg.slice('--max-iterations='.length));
      if (!Number.isFinite(value) || value < 1 || value > 30) {
        throw new Error('--max-iterations must be 1-30');
      }
      parsed.maxIterations = Math.trunc(value);
      continue;
    }

    if (arg.startsWith('--root=')) {
      parsed.root = path.resolve(arg.slice('--root='.length));
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

    if (arg.startsWith('--format=')) {
      parsed.format = arg.slice('--format='.length).toLowerCase();
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

function runHarnessHealth(args) {
  const script = path.join(__dirname, 'harness-health.js');
  const cmdArgs = [
    script,
    '--scope', args.scope,
    '--threshold', String(args.threshold),
    '--format', 'json',
    '--root', args.root,
    '--owner', args.owner,
    '--due-days', String(args.dueDays),
  ];

  const result = spawnSync('node', cmdArgs, {
    encoding: 'utf8',
    stdio: ['ignore', 'pipe', 'pipe'],
  });

  if (result.status !== 0) {
    const message = (result.stderr || '').trim() || (result.stdout || '').trim() || 'Unknown failure';
    throw new Error(`harness-health failed: ${message}`);
  }

  try {
    return JSON.parse(result.stdout || '{}');
  } catch (error) {
    throw new Error(`Failed to parse harness-health JSON output: ${error.message}`);
  }
}

function actionsEqual(left, right) {
  if (!left || !right) {
    return left === right;
  }

  const l = JSON.stringify(left);
  const r = JSON.stringify(right);
  return l === r;
}

function buildIterationPrompt(iteration, report) {
  const actions = report.top_actions || [];
  const lines = [];

  lines.push(`Iteration ${iteration}: Run this in the same thread:`);
  lines.push('1) /plan');
  lines.push('   - Prioritize:');
  actions.slice(0, 3).forEach((action, index) => {
    lines.push(`   ${index + 1}. ${action.category}: ${action.action}`);
    if (action.path) {
      lines.push(`      path=${action.path}`);
    }
  });
  lines.push('2) /tdd');
  lines.push('3) Implement fixes prioritizing the highest-point failing checks.');
  lines.push('4) /code-review');
  lines.push('5) Re-run: /harness-loop --scope ' + report.scope + ` --threshold ${report.threshold} --max-iterations 1`);
  lines.push('');

  if (report.blockers && report.blockers.length > 0) {
    lines.push('Generated blockers:');
    report.blockers.forEach(blocker => {
      lines.push(`- [${blocker.id}] owner=${blocker.owner}, due=${blocker.due}`);
      lines.push(`  ${blocker.action}`);
      lines.push(`  path=${blocker.path}`);
    });
  }

  return lines.join('\n');
}

function buildLoopReport(args, iterations) {
  const last = iterations[iterations.length - 1];

  let finalStatus = 'OK';
  let reason = 'Completed';

  if (last.health.status !== 'OK') {
    if (iterations.length >= args.maxIterations) {
      finalStatus = 'MAX_REACHED';
      reason = `threshold ${args.threshold} not reached within ${args.maxIterations} iteration(s)`;
    } else if (iterations.length >= 2 && actionsEqual(
      iterations[iterations.length - 1].health.top_actions,
      iterations[iterations.length - 2].health.top_actions
    )) {
      finalStatus = 'STALL';
      reason = 'Top actions did not change between iterations';
    } else {
      finalStatus = 'BLOCKED';
      reason = 'Iteration limit not reached yet';
    }
  }

  return {
    harness: 'refinement-loop',
    generated_at: new Date().toISOString(),
    scope: args.scope,
    threshold: args.threshold,
    root: args.root,
    max_iterations: args.maxIterations,
    status: finalStatus,
    reason,
    iterations,
  };
}

function formatText(report) {
  const lines = [];
  lines.push('Harness Refinement Loop');
  lines.push(`Scope: ${report.scope}`);
  lines.push(`Threshold: ${report.threshold}`);
  lines.push(`Status: ${report.status}`);
  lines.push(`Reason: ${report.reason}`);
  lines.push(`Iterations: ${report.iterations.length}/${report.max_iterations}`);
  lines.push('');

  report.iterations.forEach(entry => {
    lines.push(`Iteration ${entry.iteration}`);
    lines.push(`  Score: ${entry.health.score}/${entry.health.max_score}`);
    lines.push(`  Status: ${entry.health.status}`);
    lines.push(`  Failed checks: ${entry.health.failed_checks_count}`);
    if (entry.operator_prompt) {
      lines.push('  Next actions:');
      entry.operator_prompt.split('\n').forEach(line => lines.push(`    ${line}`));
    }
    lines.push('');
  });

  if (report.iterations.length > 0 && report.iterations[report.iterations.length - 1].health.status === 'OK') {
    lines.push('Goal achieved: harness reached threshold.');
  }

  return `${lines.join('\n')}\n`;
}

function main() {
  try {
    const args = parseArgs(process.argv);

    if (args.help) {
      usage();
    }

    const iterations = [];

    for (let iteration = 1; iteration <= args.maxIterations; iteration += 1) {
      const health = runHarnessHealth(args);
      const operatorPrompt = buildIterationPrompt(iteration, health);

      iterations.push({
        iteration,
        health,
        operator_prompt: operatorPrompt,
      });

      if (health.status === 'OK') {
        break;
      }

      if (iteration < args.maxIterations) {
        // continue until max iterations unless stalled.
        const isStall = iterations.length >= 2 && actionsEqual(
          iterations[iterations.length - 1].health.top_actions,
          iterations[iterations.length - 2].health.top_actions
        );

        if (isStall) {
          break;
        }
      }
    }

    const report = buildLoopReport(args, iterations);

    if (args.format === 'json') {
      process.stdout.write(`${JSON.stringify(report, null, 2)}\n`);
      process.exit(report.status === 'OK' ? 0 : 2);
      return;
    }

    process.stdout.write(formatText(report));
    process.exit(report.status === 'OK' ? 0 : 2);
  } catch (error) {
    process.stderr.write(`Error: ${error.message}\n`);
    process.exit(1);
  }
}

main();
