import React from 'react';
import { DocumentListItem } from '../../lib/api/types';

interface DocumentSidebarProps {
  /** Documents to display, in the order received from the server (updatedAt DESC). */
  documents: DocumentListItem[];
  /** The id of the currently selected document, or undefined for none. */
  selectedId: string | undefined;
  /** Callback invoked when the user selects a document. */
  onSelect: (id: string) => void;
}

/**
 * DocumentSidebar component.
 *
 * Renders a scrollable sidebar listing document titles with:
 * - A header showing "Documents (N)" count.
 * - A listbox of option buttons, one per document.
 * - Active item visually distinct and aria-selected="true".
 * - Title truncation with ellipsis (single line); full title on hover via `title` attribute.
 * - Native `<button>` items so Tab/Enter/Space work out of the box.
 *
 * The component renders in the order received — no client-side re-sort.
 */
export const DocumentSidebar: React.FC<DocumentSidebarProps> = ({
  documents,
  selectedId,
  onSelect,
}) => {
  return (
    <nav
      aria-label="Documents"
      className="w-64 flex-shrink-0 border-r border-gray-200 flex flex-col"
    >
      <div className="p-3 border-b border-gray-200">
        <h3 className="text-sm font-semibold text-gray-700">
          Documents ({documents.length})
        </h3>
      </div>

      <ul
        role="listbox"
        aria-label="Documents"
        className="overflow-y-auto flex-1"
      >
        {documents.map((doc) => {
          const isSelected = doc.id === selectedId;
          return (
            <li key={doc.id}>
              <button
                role="option"
                aria-selected={isSelected}
                title={doc.title}
                onClick={() => onSelect(doc.id)}
                className={[
                  'w-full text-left px-3 py-2 text-sm truncate block',
                  'focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset',
                  isSelected
                    ? 'bg-blue-50 text-blue-700 font-semibold'
                    : 'text-gray-700 hover:bg-gray-100',
                ].join(' ')}
              >
                {doc.title}
              </button>
            </li>
          );
        })}
      </ul>
    </nav>
  );
};
