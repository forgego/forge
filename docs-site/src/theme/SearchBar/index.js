import React, { useEffect, useMemo, useRef, useState } from 'react';
import { useHistory } from '@docusaurus/router';
import useBaseUrl from '@docusaurus/useBaseUrl';
import styles from './SearchBar.module.css';

export default function SearchBar() {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [index, setIndex] = useState(null);
  const [activeIdx, setActiveIdx] = useState(0);
  const inputRef = useRef(null);
  const history = useHistory();
  const indexUrl = useBaseUrl('/search-index.json');
  const baseUrl = useBaseUrl('/');

  useEffect(() => {
    const onKeyDown = (e) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setOpen((prev) => !prev);
      } else if (e.key === 'Escape') {
        setOpen(false);
      }
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, []);

  useEffect(() => {
    if (!open) return;
    setTimeout(() => inputRef.current?.focus(), 0);
    if (index !== null) return;
    fetch(indexUrl)
      .then((res) => (res.ok ? res.json() : []))
      .then((data) => setIndex(Array.isArray(data) ? data : []))
      .catch(() => setIndex([]));
  }, [open, index, indexUrl]);

  const results = useMemo(() => {
    if (!index || index.length === 0) return [];
    const q = query.trim().toLowerCase();
    if (!q) return index.slice(0, 8);
    const terms = q.split(/\s+/).filter(Boolean);
    return index
      .map((item) => {
        const title = item.title.toLowerCase();
        const section = (item.section || '').toLowerCase();
        const desc = (item.description || '').toLowerCase();
        const headings = (item.headings || []).join(' ').toLowerCase();
        const excerpt = (item.excerpt || '').toLowerCase();
        let score = 0;
        for (const t of terms) {
          if (title.includes(t)) score += 10;
          if (headings.includes(t)) score += 5;
          if (section.includes(t) || desc.includes(t)) score += 3;
          if (excerpt.includes(t)) score += 1;
        }
        return { item, score };
      })
      .filter((r) => r.score > 0)
      .sort((a, b) => b.score - a.score)
      .slice(0, 10)
      .map((r) => r.item);
  }, [index, query]);

  useEffect(() => {
    setActiveIdx(0);
  }, [query, open]);

  const navigateTo = (urlPath) => {
    const cleanBase = baseUrl.endsWith('/') ? baseUrl.slice(0, -1) : baseUrl;
    const target = `${cleanBase}${urlPath}`;
    setOpen(false);
    setQuery('');
    history.push(target);
  };

  const handleInputKeyDown = (e) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setActiveIdx((i) => Math.min(i + 1, Math.max(results.length - 1, 0)));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setActiveIdx((i) => Math.max(i - 1, 0));
    } else if (e.key === 'Enter' && results[activeIdx]) {
      e.preventDefault();
      navigateTo(results[activeIdx].url);
    }
  };

  return (
    <>
      <button
        type="button"
        className={styles.trigger}
        onClick={() => setOpen(true)}
        aria-label="Search documentation">
        <span className={styles.triggerText}>Search docs...</span>
        <kbd className={styles.kbd}>⌘K</kbd>
      </button>

      {open && (
        <div
          className={styles.backdrop}
          role="dialog"
          aria-modal="true"
          aria-label="Search documentation"
          onClick={() => setOpen(false)}>
          <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
            <div className={styles.inputRow}>
              <input
                ref={inputRef}
                type="search"
                className={styles.input}
                placeholder="Search Forge docs (models, ORM, CLI, migrations, API...)"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                onKeyDown={handleInputKeyDown}
              />
              <button
                type="button"
                className={styles.closeBtn}
                onClick={() => setOpen(false)}>
                Esc
              </button>
            </div>

            <div className={styles.results}>
              {index === null && <div className={styles.empty}>Loading index...</div>}
              {index !== null && results.length === 0 && (
                <div className={styles.empty}>No matching pages found.</div>
              )}
              {results.map((item, idx) => (
                <button
                  key={item.id}
                  type="button"
                  className={`${styles.resultItem} ${idx === activeIdx ? styles.resultActive : ''}`}
                  onMouseEnter={() => setActiveIdx(idx)}
                  onClick={() => navigateTo(item.url)}>
                  <div className={styles.resultTop}>
                    <span className={styles.resultTitle}>{item.title}</span>
                    <span className={styles.resultSection}>{item.section}</span>
                  </div>
                  {item.description && (
                    <div className={styles.resultDesc}>{item.description}</div>
                  )}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}
    </>
  );
}
