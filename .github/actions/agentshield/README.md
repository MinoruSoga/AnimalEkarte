# Maintained AgentShield action

Upstream: [affaan-m/agentshield v1.6.0](https://github.com/affaan-m/agentshield/tree/b0891303bdcd6037376a94263d45cfd2ff3dfb98).
The MIT license is retained in LICENSE. action.yml is unchanged from that commit.
dist/action.js is the self-contained Node24 action bundle. provenance.json records
the upstream commit, patch digest and reviewed bundle digest.

## Local corrections
1. Exact .prime/agent/settings.json paths carry Prime schema metadata. The file
   remains settings-json; only the missing Claude permissions/PreToolUse rules
   become inapplicable. Existing secrets, MCP and all other rules remain active.
2. Explicit quoted prohibitions outside fenced code retain an info audit finding,
   including ID, line and evidence. Actual, ambiguous, exception-bearing and
   fenced destructive instructions remain high. Prose does not prove enforcement.

upstream.patch contains all source/test/built changes needed to reproduce this
bundle, including31 new regressions. On the pinned upstream checkout, apply that
patch, install from the fixed lock with npm ci --ignore-scripts, then run tsc,
ESLint and tsup. Do this in an isolated Node24 environment with no operational
credentials. The selected discovery/scanner/agents/permissions/hooks and
MCP/remote/secrets/poisoning/action suites total1,069 passing tests.
Verify the resulting bundle against provenance.json; do not hand-edit dist.

## Execution trust
The Security Scan workflow checks out this action from an explicitly reviewed
full repository commit into a separate path and verifies the bundle SHA-256.
The PR checkout is scanned as data; its action bundle is not executed.
No dependencies are installed during a Security Scan run.
Updates require a reviewed source patch, regression checks, rebuild, bundle hash,
provenance refresh and an intentional workflow commit-pin change. Never switch
the execution pin to a PR head, mutable branch, tag or event-provided ref.
No input defaults, permissions, secrets handling or severity gates are changed.

The corrections reduce the existing PR606 threshold count32→29; the scan still
fails when fail-on-findings=true. Capability26, backend settings2 and device1
remain unresolved. No suppression, device/backend runtime guard or role tool
reduction is included in this action.
