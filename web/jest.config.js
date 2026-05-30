/* eslint-disable @typescript-eslint/no-require-imports */
const nextJest = require('next/jest')

const createJestConfig = nextJest({
  dir: './',
})

const customJestConfig = {
  setupFiles: ['<rootDir>/jest.polyfills.js'],
  setupFilesAfterEnv: ['<rootDir>/jest.setup.ts'],
  testEnvironment: 'jest-environment-jsdom',
  testEnvironmentOptions: {
    customExportConditions: [''],
  },
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/$1',
    // Map CSS imports to an empty module in tests (e.g. highlight.js theme CSS)
    '\\.(css|less|scss|sass)$': '<rootDir>/__mocks__/styleMock.js',
  },
}

module.exports = async () => {
  const config = await createJestConfig(customJestConfig)();

  // All ESM-only packages that need transpilation by the SWC transform.
  // The pattern also covers nested node_modules (e.g. mdast-util-find-and-replace/node_modules/escape-string-regexp).
  const esmPackages = [
    // MSW (pre-existing)
    'msw', '@mswjs', 'rettime', 'until-async', 'is-node-process', 'outvariant',
    'strict-event-emitter', 'headers-polyfill', '@open-draft', '@bundled-es-modules',
    // Markdown stack — top-level packages
    'react-markdown', 'remark-gfm', 'rehype-sanitize', 'rehype-highlight',
    'remark-parse', 'remark-rehype', 'remark-stringify',
    'unified', 'vfile', 'vfile-message', 'bail', 'trough', 'is-plain-obj',
    // hast utilities
    'hast-util-sanitize', 'hast-util-to-jsx-runtime', 'hast-util-whitespace',
    'hast-util-is-element', 'hast-util-to-text', 'hast-util-to-html',
    'hast-util-from-parse5', 'hast-util-raw',
    'property-information', 'html-url-attributes',
    'space-separated-tokens', 'comma-separated-tokens',
    // unist utilities
    'unist-util-visit', 'unist-util-visit-parents', 'unist-util-is', 'unist-util-position',
    'unist-util-stringify-position', 'unist-util-find-after',
    // mdast utilities
    'mdast-util-to-hast', 'mdast-util-to-markdown', 'mdast-util-to-string',
    'mdast-util-from-markdown', 'mdast-util-phrasing', 'mdast-util-find-and-replace',
    'mdast-util-gfm', 'mdast-util-gfm-autolink-literal', 'mdast-util-gfm-footnote',
    'mdast-util-gfm-strikethrough', 'mdast-util-gfm-table', 'mdast-util-gfm-task-list-item',
    'mdast-util-mdx-expression', 'mdast-util-mdx-jsx', 'mdast-util-mdxjs-esm',
    // micromark
    'micromark', 'micromark-core-commonmark', 'micromark-extension-gfm',
    'micromark-extension-gfm-autolink-literal', 'micromark-extension-gfm-footnote',
    'micromark-extension-gfm-strikethrough', 'micromark-extension-gfm-table',
    'micromark-extension-gfm-tagfilter', 'micromark-extension-gfm-task-list-item',
    'micromark-factory-destination', 'micromark-factory-label', 'micromark-factory-space',
    'micromark-factory-title', 'micromark-factory-whitespace',
    'micromark-util-character', 'micromark-util-chunked', 'micromark-util-classify-character',
    'micromark-util-combine-extensions', 'micromark-util-decode-numeric-character-reference',
    'micromark-util-decode-string', 'micromark-util-encode', 'micromark-util-html-tag-name',
    'micromark-util-normalize-identifier', 'micromark-util-resolve-all',
    'micromark-util-sanitize-uri', 'micromark-util-subtokenize', 'micromark-util-symbol',
    'micromark-util-types',
    // other utilities used in the remark/rehype ecosystem
    'ccount', 'decode-named-character-reference', 'character-entities',
    'character-entities-html4', 'character-entities-legacy', 'character-reference-invalid',
    'zwitch', 'longest-streak', 'trim-lines', 'devlop', 'lowlight',
    'parse-entities', 'stringify-entities', 'estree-util-is-identifier-name',
    'is-alphabetical', 'is-alphanumerical', 'is-decimal', 'is-hexadecimal',
    'entities', 'parse5', 'markdown-table',
    // Nested ESM packages (inside other packages' node_modules)
    'escape-string-regexp',
  ].join('|');

  // This pattern handles both top-level and nested node_modules for the listed packages.
  config.transformIgnorePatterns = [
    `/node_modules/(?!.*(${esmPackages}))`,
  ];
  return config;
}
