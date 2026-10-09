#!/usr/bin/env node
/**
 * Stdio-to-HTTP Plane MCP Bridge for Antigravity
 * Bridges JSON-RPC 2.0 messages from stdio to https://mcp.plane.so/http/api-key/mcp
 */
import fs from 'fs';
import os from 'os';
import path from 'path';
import readline from 'readline';

const keyFile = process.env.PLANE_MCP_KEY_FILE || path.join(os.homedir(), '.config/devin/plane-mcp-key');
let apiKey = '';
try {
  apiKey = fs.readFileSync(keyFile, 'utf-8').trim();
} catch (e) {
  process.stderr.write(`Failed to read key file: ${keyFile}\n`);
  process.exit(1);
}

const endpoint = 'https://mcp.plane.so/http/api-key/mcp';
const workspace = 'baritechllc';

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  terminal: false
});

rl.on('line', async (line) => {
  if (!line.trim()) return;
  try {
    const payload = JSON.parse(line);
    const res = await fetch(endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json, text/event-stream, */*',
        'Authorization': `Bearer ${apiKey}`,
        'X-API-Key': apiKey,
        'X-Workspace-slug': workspace
      },
      body: JSON.stringify(payload)
    });

    const text = await res.text();
    // Parse SSE lines or raw json
    const sseLines = text.split('\n');
    let handled = false;
    for (const s of sseLines) {
      if (s.startsWith('data: ')) {
        const jsonStr = s.slice(6);
        process.stdout.write(jsonStr + '\n');
        handled = true;
      }
    }
    if (!handled && text.trim().startsWith('{')) {
      process.stdout.write(text.trim() + '\n');
    }
  } catch (err) {
    process.stderr.write(`Bridge error: ${err.message}\n`);
  }
});
