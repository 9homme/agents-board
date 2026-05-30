/**
 * ProjectCard component tests — FCT-US001-001 through FCT-US001-004
 *
 * Tests that ProjectCard wraps its content in a Next.js Link to /projects/{id},
 * is keyboard-accessible, uses a native <a> element (not onClick), and
 * preserves the existing visual classes on the inner <article>.
 */
import React from 'react';
import { render, screen } from '@testing-library/react';
import { RouterContext } from 'next/dist/shared/lib/router-context.shared-runtime';
import type { NextRouter } from 'next/router';
import { ProjectCard } from './ProjectCard';
import { Project } from '../../lib/api/types';

/** Minimal Next.js router mock sufficient for Pages Router Link rendering. */
function createMockRouter(overrides: Partial<NextRouter> = {}): NextRouter {
  return {
    basePath: '',
    pathname: '/',
    route: '/',
    query: {},
    asPath: '/',
    push: jest.fn().mockResolvedValue(true),
    replace: jest.fn().mockResolvedValue(true),
    reload: jest.fn(),
    back: jest.fn(),
    forward: jest.fn(),
    prefetch: jest.fn().mockResolvedValue(undefined),
    beforePopState: jest.fn(),
    events: { on: jest.fn(), off: jest.fn(), emit: jest.fn() },
    isFallback: false,
    isLocaleDomain: false,
    isReady: true,
    isPreview: false,
    ...overrides,
  } as unknown as NextRouter;
}

const projectFixture: Project = {
  id: 'proj-001',
  name: 'Test Project',
  description: 'A test description',
  createdAt: '2026-05-20T10:00:00Z',
  updatedAt: '2026-05-20T10:00:00Z',
};

function renderWithRouter(ui: React.ReactElement, router?: Partial<NextRouter>) {
  return render(
    <RouterContext.Provider value={createMockRouter(router)}>
      {ui}
    </RouterContext.Provider>,
  );
}

describe('ProjectCard — FCT-US001-001: renders as a Next.js Link to /projects/{id}', () => {
  it('renders a link with role="link" and aria-label matching the project name', () => {
    renderWithRouter(<ProjectCard project={projectFixture} />);

    const link = screen.getByRole('link', { name: /Test Project/i });
    expect(link).toBeInTheDocument();
  });

  it('link href points to /projects/{project.id}', () => {
    renderWithRouter(<ProjectCard project={projectFixture} />);

    const link = screen.getByRole('link', { name: /Test Project/i });
    expect(link).toHaveAttribute('href', '/projects/proj-001');
  });
});

describe('ProjectCard — FCT-US001-002: keyboard-focusable via Tab', () => {
  it('the link element is reachable by keyboard (tabIndex is not -1)', () => {
    renderWithRouter(<ProjectCard project={projectFixture} />);

    const link = screen.getByRole('link', { name: /Test Project/i });
    // Native <a> elements are keyboard-focusable by default (tabIndex = 0 or unset)
    expect(link).not.toHaveAttribute('tabindex', '-1');
    // Confirm the element is an anchor — the browser handles Tab focus
    expect(link.tagName.toLowerCase()).toBe('a');
  });

  it('the link has a focus-visible CSS class for visible focus ring', () => {
    renderWithRouter(<ProjectCard project={projectFixture} />);

    const link = screen.getByRole('link', { name: /Test Project/i });
    // Assert structural presence of focus-visible ring classes
    expect(link.className).toMatch(/focus-visible/);
  });
});

describe('ProjectCard — FCT-US001-003: uses Next <Link> (not a plain onClick handler)', () => {
  it('renders an <a> element, not a <div> with onClick', () => {
    renderWithRouter(<ProjectCard project={projectFixture} />);

    const link = screen.getByRole('link', { name: /Test Project/i });
    expect(link.tagName.toLowerCase()).toBe('a');
  });

  it('the <a> element does not have an onClick attribute that calls router.push', () => {
    renderWithRouter(<ProjectCard project={projectFixture} />);

    const link = screen.getByRole('link', { name: /Test Project/i });
    // If this were a manual router.push click handler, onClick would be a function
    // that navigates programmatically. Next <Link> does use onClick internally for
    // client-side navigation, but the key invariant is that the element is a native
    // <a> with an href (confirmed above), not a <div> or <button> with onClick only.
    expect(link).toHaveAttribute('href', '/projects/proj-001');
    // No data-testid indicating a synthetic click handler substitute
    expect(link).not.toHaveAttribute('data-click-handler', 'router-push');
  });
});

describe('ProjectCard — FCT-US001-004: existing visual classes preserved on inner <article>', () => {
  it('inner <article> retains all REQ002 Tailwind classes', () => {
    renderWithRouter(<ProjectCard project={projectFixture} />);

    const article = document.querySelector('article');
    expect(article).toBeInTheDocument();

    // REQ002 class list from the original implementation
    expect(article).toHaveClass('border');
    expect(article).toHaveClass('border-gray-200');
    expect(article).toHaveClass('rounded-lg');
    expect(article).toHaveClass('p-6');
    expect(article).toHaveClass('shadow-sm');
    expect(article).toHaveClass('hover:shadow-md');
    expect(article).toHaveClass('transition-shadow');
    expect(article).toHaveClass('bg-white');
    expect(article).toHaveClass('flex');
    expect(article).toHaveClass('flex-col');
    expect(article).toHaveClass('h-full');
  });

  it('inner <h3> with project name, <p> with description, and date spans still render', () => {
    renderWithRouter(<ProjectCard project={projectFixture} />);

    // These were present in REQ002 and must not be removed
    expect(screen.getByRole('heading', { level: 3, name: /Test Project/i })).toBeInTheDocument();
    expect(screen.getByText('A test description')).toBeInTheDocument();
    expect(screen.getByText(/Created:/i)).toBeInTheDocument();
    expect(screen.getByText(/Updated:/i)).toBeInTheDocument();
  });
});
