import fs from 'node:fs';
import path from 'node:path';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const siteDir = path.resolve(__dirname, '..');
const docsDir = path.join(siteDir, 'docs');
const staticDir = path.join(siteDir, 'static');
const require = createRequire(import.meta.url);

const BASE_URL = 'https://forgego.github.io/forge';

function parseFrontmatter(raw) {
  const match = raw.match(/^---\r?\n([\s\S]*?)\r?\n---\r?\n([\s\S]*)$/);
  if (!match) {
    return { meta: {}, body: raw };
  }
  const meta = {};
  for (const line of match[1].split(/\r?\n/)) {
    const idx = line.indexOf(':');
    if (idx > 0) {
      const key = line.slice(0, idx).trim();
      const val = line.slice(idx + 1).trim().replace(/^['"]|['"]$/g, '');
      meta[key] = val;
    }
  }
  return { meta, body: match[2] };
}

function extractTitle(meta, body, docId) {
  if (meta.title) return meta.title;
  const h1 = body.match(/^#\s+(.+)$/m);
  if (h1) return h1[1].trim();
  const last = docId.split('/').pop() || docId;
  return last.replace(/[-_]/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
}

function extractHeadings(body) {
  const headings = [];
  let inFence = false;
  for (const line of body.split(/\r?\n/)) {
    if (/^\s*(```|~~~)/.test(line)) {
      inFence = !inFence;
      continue;
    }
    if (inFence) continue;
    const m = line.match(/^#{2,3}\s+(.+)$/);
    if (m) {
      headings.push(m[1].replace(/`/g, '').trim());
    }
  }
  return headings;
}

function toPlainExcerpt(body, maxLen = 400) {
  const stripped = body
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/`([^`]+)`/g, '$1')
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
    .replace(/^#+\s+.*$/gm, ' ')
    .replace(/[*_>|]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim();
  return stripped.length > maxLen ? `${stripped.slice(0, maxLen)}...` : stripped;
}

function docUrlPath(docId, meta) {
  if (meta.slug === '/' || docId === 'index') {
    return '/docs/';
  }
  return `/docs/${docId}/`;
}

function collectDocsFromSidebar() {
  const sidebars = require(path.join(siteDir, 'sidebars.js'));
  const entries = [];
  for (const category of sidebars.docs || []) {
    const section = category.label || 'Documentation';
    for (const docId of category.items || []) {
      const mdPath = path.join(docsDir, `${docId}.md`);
      if (!fs.existsSync(mdPath)) continue;
      const raw = fs.readFileSync(mdPath, 'utf8');
      const { meta, body } = parseFrontmatter(raw);
      const title = extractTitle(meta, body, docId);
      const description = meta.description || '';
      const urlPath = docUrlPath(docId, meta);
      entries.push({
        id: docId,
        section,
        title,
        description,
        urlPath,
        headings: extractHeadings(body),
        excerpt: toPlainExcerpt(body),
        body: body.trim(),
      });
    }
  }
  return entries;
}

function buildLlmsTxt(entries) {
  const lines = [
    '# Forge Documentation',
    '',
    '> Forge is a schema-first, batteries-included web framework for Go (v0.2.x): typed ORM queries, AST-driven SQL migrations, a DRF-style REST API, identity, and an embedded React admin panel.',
    '',
    'Canonical support tiers and platform matrix: ' + `${BASE_URL}/docs/status/`,
    '',
  ];

  let currentSection = '';
  for (const entry of entries) {
    if (entry.section !== currentSection) {
      currentSection = entry.section;
      lines.push(`## ${currentSection}`, '');
    }
    const desc = entry.description ? `: ${entry.description}` : '';
    lines.push(`- [${entry.title}](${BASE_URL}${entry.urlPath})${desc}`);
  }
  lines.push('');
  return lines.join('\n');
}

function buildLlmsFullTxt(entries) {
  const chunks = [
    '# Forge Complete Documentation',
    '',
    '> Generated from docs-site/docs/ for AI coding agents and offline context loading.',
    '',
  ];
  for (const entry of entries) {
    chunks.push(
      `---`,
      ``,
      `<!-- source: docs-site/docs/${entry.id}.md | url: ${BASE_URL}${entry.urlPath} -->`,
      ``,
      entry.body,
      ``,
    );
  }
  return chunks.join('\n');
}

function buildSearchIndex(entries) {
  return entries.map(({ id, section, title, description, urlPath, headings, excerpt }) => ({
    id,
    section,
    title,
    description,
    url: urlPath,
    headings,
    excerpt,
  }));
}

fs.mkdirSync(staticDir, { recursive: true });
const entries = collectDocsFromSidebar();

fs.writeFileSync(path.join(staticDir, 'llms.txt'), buildLlmsTxt(entries), 'utf8');
fs.writeFileSync(path.join(staticDir, 'llms-full.txt'), buildLlmsFullTxt(entries), 'utf8');
fs.writeFileSync(
  path.join(staticDir, 'search-index.json'),
  JSON.stringify(buildSearchIndex(entries)),
  'utf8',
);

console.log(`Generated llms.txt, llms-full.txt, and search-index.json (${entries.length} pages).`);
