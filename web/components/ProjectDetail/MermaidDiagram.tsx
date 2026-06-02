/**
 * MermaidDiagram — CSR-only component that renders a mermaid diagram as inline SVG.
 *
 * Lazy-load contract:
 * - `mermaid` is imported via a dynamic `import('mermaid')` inside `useEffect`, so the
 *   ~300 KB gzipped mermaid chunk is only loaded when this component mounts (i.e. when a
 *   `language-mermaid` fenced block is actually present in the rendered document).
 * - The component must never run during SSR / Next.js build because mermaid references
 *   `window` / `document` at import time. This is guaranteed by the dynamic import inside
 *   `useEffect` (which never executes on the server).
 *
 * Error-handling contract:
 * - If `import('mermaid')` fails OR `mermaid.render(id, source)` throws / rejects, the
 *   component renders a `<div role="alert">Could not render diagram</div>` followed by the
 *   raw `source` in a `<pre><code>` fallback. Errors never propagate — the surrounding
 *   markdown continues to render.
 *
 * Unique-id contract:
 * - `useId()` (React 18+) provides a per-instance stable id, prefixed `mermaid-`, so that
 *   multiple diagrams in the same document do not collide on mermaid's internal SVG id
 *   namespace.
 *
 * Architecture ref: §"Components → Frontend" row MermaidDiagram.tsx; D-004; §"Mermaid mechanics".
 */
import React, { useEffect, useId, useRef, useState } from 'react';

/** Cached mermaid module — loaded once and reused on subsequent renders. */
let mermaidModuleCache: (typeof import('mermaid'))['default'] | null = null;
/** Flag to avoid initializing mermaid more than once per module lifetime. */
let mermaidInitialized = false;

interface MermaidDiagramProps {
  /** Raw mermaid diagram source (the content between the fenced code triple-backticks). */
  source: string;
}

type RenderState =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'success'; svg: string; ariaLabel: string }
  | { status: 'error' };

/**
 * Extract an accessible label from the mermaid source.
 *
 * Uses the first non-empty line as a description if the rendered SVG
 * does not carry a `<title>` element.
 */
function deriveAriaLabel(source: string): string {
  const firstLine = source
    .split('\n')
    .find((l) => l.trim().length > 0);
  return firstLine ? firstLine.trim() : 'Mermaid diagram';
}

/**
 * Check whether an SVG string already contains a `<title>` element.
 * If it does, the wrapper does not need to supply an `aria-label`.
 */
function svgHasTitle(svg: string): boolean {
  return /<title[^>]*>/.test(svg);
}

/**
 * MermaidDiagram component — renders a mermaid source string as an inline SVG diagram.
 *
 * See module-level doc comment for the full lazy-load / error-handling / unique-id contracts.
 */
export const MermaidDiagram = ({ source }: MermaidDiagramProps): React.ReactElement => {
  // Per-instance stable id for mermaid.render(id, source) — prevents namespace collisions
  // when multiple MermaidDiagram instances exist in the same document.
  const rawId = useId();
  // useId() may produce characters like ":" that are invalid in HTML ids.
  const mermaidId = `mermaid-${rawId.replace(/[^a-zA-Z0-9-_]/g, '-')}`;

  const [renderState, setRenderState] = useState<RenderState>({ status: 'idle' });
  const containerRef = useRef<HTMLDivElement | null>(null);

  // Effect that runs on transition into 'success'. Parses the mermaid-produced SVG
  // string via DOMParser and appends the parsed <svg> node into the ref'd wrapper div.
  // Cleanup removes the appended node so React 18 strict-mode double-invoke produces
  // exactly one <svg> child (R7 mitigation per architecture §11.1.2).
  useEffect(() => {
    if (renderState.status !== 'success') return;
    const host = containerRef.current;
    if (!host) return;

    try {
      const parsed = new DOMParser().parseFromString(renderState.svg, 'image/svg+xml');
      const svgNode = parsed.documentElement;

      // Clear any previous child (covers the source-change re-render path) and attach.
      while (host.firstChild) host.removeChild(host.firstChild);
      host.appendChild(svgNode);

      return () => {
        if (host.firstChild === svgNode) host.removeChild(svgNode);
      };
    } catch {
      // DOMParser failure is defensive; mermaid v11 emits well-formed SVG strings.
      // Silently skip to avoid crashing the component on malformed input.
    }
  }, [renderState]);

  useEffect(() => {
    let cancelled = false;
    setRenderState({ status: 'loading' });

    const run = async (): Promise<void> => {
      try {
        // Lazy-load mermaid (cached after first successful load).
        if (!mermaidModuleCache) {
          const mod = await import('mermaid');
          mermaidModuleCache = mod.default;
        }
        const mermaid = mermaidModuleCache;

        // Initialize once per module lifetime (securityLevel:'strict' per architecture).
        if (!mermaidInitialized) {
          mermaid.initialize({ startOnLoad: false, securityLevel: 'strict' });
          mermaidInitialized = true;
        }

        // mermaid v11 render() is async and returns { svg, bindFunctions }.
        const { svg } = await mermaid.render(mermaidId, source);

        if (cancelled) return;

        const ariaLabel = svgHasTitle(svg) ? '' : deriveAriaLabel(source);
        setRenderState({ status: 'success', svg, ariaLabel });
      } catch {
        if (cancelled) return;
        setRenderState({ status: 'error' });
      }
    };

    void run();

    return () => {
      cancelled = true;
    };
  }, [source, mermaidId]);

  if (renderState.status === 'error') {
    return (
      <>
        <div role="alert" className="text-sm text-red-600 inline-block">
          Could not render diagram
        </div>
        <pre className="mt-2 text-xs text-gray-600 overflow-auto">
          <code>{source}</code>
        </pre>
      </>
    );
  }

  if (renderState.status === 'success') {
    const { ariaLabel } = renderState;
    return (
      <div
        ref={containerRef}
        role="img"
        aria-label={ariaLabel || undefined}
        style={{ maxWidth: '100%', overflowX: 'auto' }}
      />
    );
  }

  // idle / loading: render nothing (brief flash is acceptable per US003 UX expectations)
  return <></>;
};
