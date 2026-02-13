#!/usr/bin/env node
'use strict';

const { spawn } = require('child_process');

const PLAYWRIGHT_CLI = 'H:/home/hamid/.npm-global/node_modules/@playwright/mcp/cli.js';
const child = spawn(process.execPath, [PLAYWRIGHT_CLI], {
  stdio: ['pipe', 'pipe', 'pipe'],
  windowsHide: true,
});

child.stderr.on('data', (d) => {
  // Keep stderr visible to Codex logs for debugging.
  process.stderr.write(d);
});

child.on('exit', (code, signal) => {
  if (!process.stdout.destroyed) process.stdout.end();
  process.exit(code ?? (signal ? 1 : 0));
});

process.on('exit', () => {
  try { child.kill(); } catch {}
});

function writeContentLengthMessage(jsonText) {
  const payload = Buffer.from(jsonText, 'utf8');
  process.stdout.write(`Content-Length: ${payload.length}\r\n\r\n`);
  process.stdout.write(payload);
}

function tryParseJson(text) {
  const t = text.trim();
  if (!t) return null;
  if (!t.startsWith('{') && !t.startsWith('[')) return null;
  try {
    JSON.parse(t);
    return t;
  } catch {
    return null;
  }
}

// Child stdout is line-delimited JSON. Convert to Content-Length frames.
let childBuffer = '';
child.stdout.on('data', (chunk) => {
  childBuffer += chunk.toString('utf8');
  while (true) {
    const nl = childBuffer.indexOf('\n');
    if (nl < 0) break;
    const line = childBuffer.slice(0, nl);
    childBuffer = childBuffer.slice(nl + 1);
    const json = tryParseJson(line);
    if (json) writeContentLengthMessage(json);
  }
});

// Codex stdin is Content-Length framed JSON-RPC. Convert to line-delimited JSON.
let inBuffer = '';
function forwardLine(jsonText) {
  child.stdin.write(jsonText.trim() + '\n');
}

function consumeInput() {
  while (true) {
    const headerEnd = inBuffer.indexOf('\r\n\r\n');
    if (headerEnd >= 0) {
      const header = inBuffer.slice(0, headerEnd);
      const match = /Content-Length:\s*(\d+)/i.exec(header);
      if (match) {
        const len = Number(match[1]);
        const bodyStart = headerEnd + 4;
        if (inBuffer.length < bodyStart + len) return;
        const body = inBuffer.slice(bodyStart, bodyStart + len);
        inBuffer = inBuffer.slice(bodyStart + len);
        forwardLine(body);
        continue;
      }
    }

    const nl = inBuffer.indexOf('\n');
    if (nl >= 0) {
      const line = inBuffer.slice(0, nl);
      inBuffer = inBuffer.slice(nl + 1);
      const json = tryParseJson(line);
      if (json) forwardLine(json);
      continue;
    }

    return;
  }
}

process.stdin.setEncoding('utf8');
process.stdin.on('data', (chunk) => {
  inBuffer += chunk;
  consumeInput();
});

process.stdin.on('end', () => {
  try { child.stdin.end(); } catch {}
});
