import React, { useRef } from 'react';

/** The valid tab ids for the project detail page. */
export type TabId = 'documents' | 'user-stories';

interface TabSwitcherProps {
  /** The currently active tab. */
  activeTab: TabId;
  /** Callback invoked when the user selects a different tab. */
  onTabChange: (tab: TabId) => void;
}

/**
 * TabSwitcher component — WAI-ARIA Tabs pattern.
 *
 * Renders a tablist with two tabs: Documents and User Stories.
 * Keyboard support: ArrowLeft/ArrowRight move focus and activate the tab;
 * Enter/Space also activate the focused tab.
 *
 * The caller (the page) is responsible for writing the tab to the URL via
 * a shallow router.replace call.
 */
export const TabSwitcher: React.FC<TabSwitcherProps> = ({ activeTab, onTabChange }) => {
  const tabs: { id: TabId; label: string }[] = [
    { id: 'documents', label: 'Documents' },
    { id: 'user-stories', label: 'User Stories' },
  ];

  const tabRefs = useRef<(HTMLButtonElement | null)[]>([]);

  const handleKeyDown = (
    event: React.KeyboardEvent<HTMLButtonElement>,
    index: number
  ) => {
    if (event.key === 'ArrowRight') {
      event.preventDefault();
      const next = (index + 1) % tabs.length;
      tabRefs.current[next]?.focus();
      onTabChange(tabs[next].id);
    } else if (event.key === 'ArrowLeft') {
      event.preventDefault();
      const prev = (index - 1 + tabs.length) % tabs.length;
      tabRefs.current[prev]?.focus();
      onTabChange(tabs[prev].id);
    } else if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      onTabChange(tabs[index].id);
    }
  };

  return (
    <div role="tablist" aria-label="Project tabs" className="flex border-b border-gray-200">
      {tabs.map((tab, index) => {
        const isSelected = activeTab === tab.id;
        return (
          <button
            key={tab.id}
            ref={(el) => {
              tabRefs.current[index] = el;
            }}
            role="tab"
            id={`tab-${tab.id}`}
            aria-selected={isSelected}
            aria-controls={`tabpanel-${tab.id}`}
            tabIndex={isSelected ? 0 : -1}
            onClick={() => onTabChange(tab.id)}
            onKeyDown={(e) => handleKeyDown(e, index)}
            className={[
              'px-4 py-2 text-sm font-medium border-b-2 transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500',
              isSelected
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-gray-600 hover:text-gray-900',
            ].join(' ')}
          >
            {tab.label}
          </button>
        );
      })}
    </div>
  );
};
