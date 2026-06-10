import React from 'react';
import { useProjectRequirements } from '../../hooks/useProjectRequirements';

interface RequirementSelectorProps {
  /** The project whose requirements are listed. */
  projectId: string;
  /** The currently selected requirement id (undefined = none selected). */
  selectedRequirementId: string | undefined;
  /** Called when the user selects a requirement item. */
  onSelect: (requirementId: string) => void;
}

/**
 * RequirementSelector component.
 *
 * Fetches and renders the requirements list for a project.
 * Handles loading, empty, error, and success states.
 *
 * Accessibility:
 * - Each requirement item is a `<button>` with `aria-selected` reflecting selection.
 * - Error is wrapped in `role="alert"` for screen reader announcements.
 * - Items are keyboard-navigable (native button focus + Enter/Space).
 *
 * Selection updates the `requirement` URL query param via the `onSelect` callback,
 * following the shallow-routing pattern used for the `tab` param.
 */
export const RequirementSelector: React.FC<RequirementSelectorProps> = ({
  projectId,
  selectedRequirementId,
  onSelect,
}) => {
  const { requirements, loading, error } = useProjectRequirements(projectId);

  if (loading) {
    return (
      <div
        className="flex items-center gap-2 py-2 px-3"
        aria-busy="true"
        aria-label="Loading requirements"
        data-testid="requirements-loading"
      >
        <span className="text-sm text-gray-500 italic">Loading requirements…</span>
      </div>
    );
  }

  if (error) {
    return (
      <div
        role="alert"
        className="py-2 px-3 text-sm text-red-600 bg-red-50 border border-red-200 rounded"
        data-testid="requirements-error"
      >
        Failed to load requirements. Please try again later.
      </div>
    );
  }

  if (requirements.length === 0) {
    return (
      <div
        className="py-2 px-3 text-sm text-gray-500 italic"
        data-testid="requirements-empty"
      >
        No requirements yet
      </div>
    );
  }

  return (
    /* Use div for listbox role — ul/li are noninteractive semantic HTML elements;
       the listbox + option roles belong on interactive containers per ARIA spec. */
    <div
      role="listbox"
      aria-label="Requirements"
      className="flex flex-col gap-1 py-1"
    >
      {requirements.map((req) => {
        const isSelected = req.id === selectedRequirementId;
        return (
          <div key={req.id}>
            {/* role="option" on the button so the listbox/option ARIA tree is correct */}
            <button
              type="button"
              role="option"
              aria-selected={isSelected}
              onClick={() => onSelect(req.id)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault();
                  onSelect(req.id);
                }
              }}
              className={[
                'w-full text-left px-3 py-2 rounded text-sm flex items-center gap-2 focus:outline-none focus:ring-2 focus:ring-blue-500',
                isSelected
                  ? 'bg-blue-600 text-white font-medium'
                  : 'hover:bg-gray-100 text-gray-700',
              ].join(' ')}
            >
              <span className="flex-1 truncate">{req.name}</span>
              <span
                className={[
                  'px-1.5 py-0.5 text-xs rounded font-medium',
                  isSelected ? 'bg-blue-400 text-white' : 'bg-gray-200 text-gray-600',
                ].join(' ')}
              >
                {req.status}
              </span>
            </button>
          </div>
        );
      })}
    </div>
  );
};
