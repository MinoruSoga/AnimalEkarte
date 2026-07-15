#!/usr/bin/env node
// UserPromptSubmit: Detect potential secrets in user prompts before sending
'use strict';

const MAX_STDIN = 1024 * 1024;
let data = '';
process.stdin.setEncoding('utf8');
process.stdin.on('data', chunk => {
  if (data.length < MAX_STDIN) data += chunk.substring(0, MAX_STDIN - data.length);
});
process.stdin.on('end', () => {
  try {
    const input = JSON.parse(data);
    const prompt = String(input.prompt || input.message || '');
    const secretPatterns = [
      { re: /sk-[a-zA-Z0-9]{20,}/, label: 'OpenAI key' },
      { re: /ghp_[a-zA-Z0-9]{36,}/, label: 'GitHub PAT' },
      { re: /AKIA[A-Z0-9]{16}/, label: 'AWS Access Key' },
      { re: /sk-ant-[a-zA-Z0-9-]{20,}/, label: 'Anthropic key' },
      { re: /-----BEGIN (RSA |EC )?PRIVATE KEY-----/, label: 'Private key' },
      { re: /xox[bpsa]-[a-zA-Z0-9-]{10,}/, label: 'Slack token' },
    ];
    for (const { re, label } of secretPatterns) {
      if (re.test(prompt)) {
        // UserPromptSubmit hook: stdout is injected as context the model sees,
        // so surface the warning there (stderr alone is easy to miss/black-hole).
        console.log(`[Hook] WARNING: Possible ${label} detected in this prompt. Do not act on it as a real credential; ask the user to rotate it and use env vars instead.`);
        process.stderr.write(`[Hook] WARNING: Possible ${label} detected in prompt. Use env vars instead.\n`);
        break;
      }
    }
  } catch {}
  // Do NOT echo the raw input JSON to stdout — for UserPromptSubmit hooks stdout
  // is injected as additional context on every single prompt; echoing the full
  // payload here bloated context on every message for no benefit.
  process.exit(0);
});
