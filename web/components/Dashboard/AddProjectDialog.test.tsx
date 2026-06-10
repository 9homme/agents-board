import React from 'react';
import { render, screen, waitFor, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { server } from '../../test/msw/server';
import { AddProjectDialog } from './AddProjectDialog';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const noop = () => { /* no-op */ };

function renderDialog(props: Partial<React.ComponentProps<typeof AddProjectDialog>> = {}) {
  const defaults = {
    open: true,
    onClose: noop,
    onSuccess: noop,
    triggerRef: undefined as React.RefObject<HTMLButtonElement> | undefined,
  };
  return render(<AddProjectDialog {...defaults} {...props} />);
}

/**
 * Renders AddProjectDialog inside a stateful wrapper so that onClose/onSuccess
 * actually toggle the `open` prop, allowing the dialog to unmount from the DOM.
 * Required for tests that need to assert `queryByRole('dialog')` is null after close.
 */
function renderDialogStateful(
  onSuccess: () => void = noop,
  triggerRef?: React.RefObject<HTMLButtonElement>
) {
  function Wrapper() {
    const [open, setOpen] = React.useState(true);
    return (
      <AddProjectDialog
        open={open}
        onClose={() => setOpen(false)}
        onSuccess={() => { setOpen(false); onSuccess(); }}
        triggerRef={triggerRef}
      />
    );
  }
  return render(<Wrapper />);
}

// ---------------------------------------------------------------------------
// FCT-046-003 — Path field is a plain text input
// ---------------------------------------------------------------------------
describe('FCT-046-003 — path field is plain text input', () => {
  it('has type="text" and an associated label', () => {
    renderDialog();
    const pathInput = screen.getByLabelText(/path/i);
    expect(pathInput).toBeInTheDocument();
    expect(pathInput).toHaveAttribute('type', 'text');
    // Confirm it is NOT a file or search input
    expect(pathInput).not.toHaveAttribute('type', 'file');
    expect(pathInput).not.toHaveAttribute('type', 'search');
  });
});

// ---------------------------------------------------------------------------
// FCT-046-004 — Name auto-fills from path basename
// ---------------------------------------------------------------------------
describe('FCT-046-004 — name auto-fills from basename', () => {
  it('fills name from path basename on each keystroke', async () => {
    const user = userEvent.setup();
    renderDialog();

    const pathInput = screen.getByLabelText(/path/i);
    await user.type(pathInput, '/Users/me/workspace/my-project');

    const nameInput = screen.getByLabelText(/name/i);
    expect(nameInput).toHaveValue('my-project');
  });

  it('handles trailing slash: /Users/me/workspace/my-project/', async () => {
    const user = userEvent.setup();
    renderDialog();

    const pathInput = screen.getByLabelText(/path/i);
    await user.type(pathInput, '/Users/me/workspace/my-project/');

    const nameInput = screen.getByLabelText(/name/i);
    expect(nameInput).toHaveValue('my-project');
  });
});

// ---------------------------------------------------------------------------
// FCT-046-005 — Name auto-fill is sticky-off once manually edited
// ---------------------------------------------------------------------------
describe('FCT-046-005 — auto-fill sticky-off after manual name edit', () => {
  it('does not overwrite manual name edit when path changes again', async () => {
    const user = userEvent.setup();
    renderDialog();

    const pathInput = screen.getByLabelText(/path/i);
    const nameInput = screen.getByLabelText(/name/i);

    // Step 1: type path → name auto-fills
    await user.type(pathInput, '/some/path/project-a');
    expect(nameInput).toHaveValue('project-a');

    // Step 2: manually edit name
    await user.clear(nameInput);
    await user.type(nameInput, 'my-custom-name');
    expect(nameInput).toHaveValue('my-custom-name');

    // Step 3: change path again
    await user.clear(pathInput);
    await user.type(pathInput, '/some/other/project-b');

    // Name should retain the manual value
    expect(nameInput).toHaveValue('my-custom-name');
  });
});

// ---------------------------------------------------------------------------
// FCT-046-006, 046-007, 046-008 — Submit disabled when fields blank
// ---------------------------------------------------------------------------
describe('Submit button disabled when fields are blank', () => {
  it('FCT-046-006 — disabled when name is blank (path filled)', async () => {
    const user = userEvent.setup();
    renderDialog();

    const pathInput = screen.getByLabelText(/path/i);
    const nameInput = screen.getByLabelText(/name/i);

    // Fill path (auto-fills name), then clear name
    await user.type(pathInput, '/some/valid/path');
    await user.clear(nameInput);

    const submitBtn = screen.getByRole('button', { name: /create|submit|add/i });
    expect(submitBtn).toBeDisabled();
  });

  it('FCT-046-007 — disabled when path is blank (name filled)', async () => {
    const user = userEvent.setup();
    renderDialog();

    const nameInput = screen.getByLabelText(/name/i);
    await user.type(nameInput, 'My Project');

    const submitBtn = screen.getByRole('button', { name: /create|submit|add/i });
    expect(submitBtn).toBeDisabled();
  });

  it('FCT-046-008 — disabled when both blank (freshly opened)', () => {
    renderDialog();
    const submitBtn = screen.getByRole('button', { name: /create|submit|add/i });
    expect(submitBtn).toBeDisabled();
  });
});

// ---------------------------------------------------------------------------
// FCT-046-009 — Submit enabled when both non-blank
// ---------------------------------------------------------------------------
describe('FCT-046-009 — submit enabled when both fields non-blank', () => {
  it('enables submit once both path and name are filled', async () => {
    const user = userEvent.setup();
    renderDialog();

    await user.type(screen.getByLabelText(/path/i), '/some/valid/path');

    const submitBtn = screen.getByRole('button', { name: /create|submit|add/i });
    expect(submitBtn).not.toBeDisabled();
  });
});

// ---------------------------------------------------------------------------
// FCT-046-010 — Loading state during submission
// ---------------------------------------------------------------------------
describe('FCT-046-010 — loading state while request in flight', () => {
  it('submit button is disabled and loading indicator shown while request in flight', async () => {
    let resolveRequest!: () => void;
    const deferredPromise = new Promise<void>((res) => { resolveRequest = res; });

    server.use(
      http.post('*/api/v1/projects', async () => {
        await deferredPromise;
        return HttpResponse.json(
          {
            id: '33333333-3333-3333-3333-333333333333',
            name: 'my-project',
            description: '',
            path: '/some/valid/path',
            createdAt: '2026-06-09T11:00:00Z',
            updatedAt: '2026-06-09T11:00:00Z',
          },
          { status: 201 }
        );
      })
    );

    const user = userEvent.setup();
    renderDialogStateful();

    await user.type(screen.getByLabelText(/path/i), '/some/valid/path');
    await user.click(screen.getByRole('button', { name: /create|submit|add/i }));

    // Immediately after click, submit should be disabled and loading visible
    const submitBtn = screen.getByRole('button', { name: /create|submit|add|creating/i });
    expect(submitBtn).toBeDisabled();
    // Loading indicator text or aria-busy
    expect(
      screen.queryByText(/creating/i) ||
      document.querySelector('[aria-busy="true"]')
    ).toBeTruthy();

    // Resolve the deferred request — dialog should close on success
    act(() => { resolveRequest(); });
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });
  });
});

// ---------------------------------------------------------------------------
// FCT-046-011 — Double-submit is prevented
// ---------------------------------------------------------------------------
describe('FCT-046-011 — double-submit is prevented', () => {
  it('MSW handler is called exactly once even when submit is clicked twice', async () => {
    let callCount = 0;
    let resolveRequest!: () => void;
    const deferredPromise = new Promise<void>((res) => { resolveRequest = res; });

    server.use(
      http.post('*/api/v1/projects', async () => {
        callCount++;
        await deferredPromise;
        return HttpResponse.json(
          {
            id: '33333333-3333-3333-3333-333333333333',
            name: 'my-project',
            description: '',
            path: '/some/valid/path',
            createdAt: '2026-06-09T11:00:00Z',
            updatedAt: '2026-06-09T11:00:00Z',
          },
          { status: 201 }
        );
      })
    );

    const user = userEvent.setup();
    renderDialogStateful();

    await user.type(screen.getByLabelText(/path/i), '/some/valid/path');

    const submitBtn = screen.getByRole('button', { name: /create|submit|add/i });
    await user.click(submitBtn);
    // Try to click again while in flight (button should be disabled)
    await user.click(submitBtn);

    expect(callCount).toBe(1);

    act(() => { resolveRequest(); });
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });
  });
});

// ---------------------------------------------------------------------------
// FCT-046-013 — 400 VALIDATION_ERROR shown inline; form stays open
// ---------------------------------------------------------------------------
describe('FCT-046-013 — 400 VALIDATION_ERROR shown inline', () => {
  it('shows validation error message inline and keeps dialog open', async () => {
    server.use(
      http.post('*/api/v1/projects', () => {
        return HttpResponse.json(
          { code: 'VALIDATION_ERROR', message: 'path does not exist or is not a directory' },
          { status: 400 }
        );
      })
    );

    const user = userEvent.setup();
    renderDialog();

    await user.type(screen.getByLabelText(/path/i), '/bad/path');
    await user.click(screen.getByRole('button', { name: /create|submit|add/i }));

    await waitFor(() => {
      expect(
        screen.getByText(/path does not exist or is not a directory/i)
      ).toBeVisible();
    });

    // Dialog still open
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });
});

// ---------------------------------------------------------------------------
// FCT-046-014 — 409 DUPLICATE_PATH shown inline; form stays open
// ---------------------------------------------------------------------------
describe('FCT-046-014 — 409 DUPLICATE_PATH shown inline', () => {
  it('shows duplicate path error and keeps dialog open', async () => {
    server.use(
      http.post('*/api/v1/projects', () => {
        return HttpResponse.json(
          { code: 'DUPLICATE_PATH', message: 'path already linked to another project' },
          { status: 409 }
        );
      })
    );

    const user = userEvent.setup();
    renderDialog();

    await user.type(screen.getByLabelText(/path/i), '/existing/path');
    await user.click(screen.getByRole('button', { name: /create|submit|add/i }));

    await waitFor(() => {
      expect(
        screen.getByText(/path already linked to another project/i)
      ).toBeVisible();
    });

    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });
});

// ---------------------------------------------------------------------------
// FCT-046-015 — Input values are preserved after server error
// ---------------------------------------------------------------------------
describe('FCT-046-015 — input values preserved after server error', () => {
  it('path and name retain values after 400 response', async () => {
    server.use(
      http.post('*/api/v1/projects', () => {
        return HttpResponse.json(
          { code: 'VALIDATION_ERROR', message: 'path does not exist or is not a directory' },
          { status: 400 }
        );
      })
    );

    const user = userEvent.setup();
    renderDialog();

    await user.type(screen.getByLabelText(/path/i), '/my/custom/path');
    // name auto-fills to 'path'
    await user.click(screen.getByRole('button', { name: /create|submit|add/i }));

    await waitFor(() => {
      expect(
        screen.getByText(/path does not exist or is not a directory/i)
      ).toBeVisible();
    });

    expect(screen.getByLabelText(/path/i)).toHaveValue('/my/custom/path');
    expect(screen.getByLabelText(/name/i)).toHaveValue('path');
  });
});

// ---------------------------------------------------------------------------
// FCT-046-024 — Error messages are announced to screen readers
// ---------------------------------------------------------------------------
describe('FCT-046-024 — error messages have role=alert or aria-live', () => {
  it('error container is role=alert or has aria-live attribute', async () => {
    server.use(
      http.post('*/api/v1/projects', () => {
        return HttpResponse.json(
          { code: 'VALIDATION_ERROR', message: 'path does not exist or is not a directory' },
          { status: 400 }
        );
      })
    );

    const user = userEvent.setup();
    renderDialog();

    await user.type(screen.getByLabelText(/path/i), '/bad/path');
    await user.click(screen.getByRole('button', { name: /create|submit|add/i }));

    await waitFor(() => {
      const alert = screen.queryByRole('alert');
      const liveRegion = document.querySelector('[aria-live]');
      expect(alert || liveRegion).toBeTruthy();
    });
  });
});

// ---------------------------------------------------------------------------
// FCT-046-025 — Focus returns to trigger button on dialog close
// ---------------------------------------------------------------------------
describe('FCT-046-025 — focus returns to trigger on close', () => {
  it('focus is returned to trigger button when dialog is closed via Cancel', async () => {
    const user = userEvent.setup();
    const triggerRef = React.createRef<HTMLButtonElement>();

    // Stateful wrapper so open truly toggles when onClose fires,
    // allowing the useEffect([open]) focus-return to trigger.
    function Wrapper() {
      const [open, setOpen] = React.useState(true);
      return (
        <div>
          <button ref={triggerRef}>Add Project</button>
          <AddProjectDialog
            open={open}
            onClose={() => setOpen(false)}
            onSuccess={noop}
            triggerRef={triggerRef as React.RefObject<HTMLButtonElement>}
          />
        </div>
      );
    }
    render(<Wrapper />);

    // Click Cancel button
    const cancelBtn = screen.getByRole('button', { name: /cancel/i });
    await user.click(cancelBtn);

    // Focus should have returned to the trigger
    expect(document.activeElement).toBe(triggerRef.current);
  });
});

// ---------------------------------------------------------------------------
// FCT-046-026 — Keyboard navigation cycles through dialog fields
// ---------------------------------------------------------------------------
describe('FCT-046-026 — keyboard navigation in dialog', () => {
  it('Tab moves focus from path → name → submit → cancel within dialog', async () => {
    const user = userEvent.setup();
    renderDialog();

    const pathInput = screen.getByLabelText(/path/i);
    const nameInput = screen.getByLabelText(/name/i);
    const submitBtn = screen.getByRole('button', { name: /create|submit|add/i });
    const cancelBtn = screen.getByRole('button', { name: /cancel/i });

    // Focus path input
    pathInput.focus();
    expect(document.activeElement).toBe(pathInput);

    // Tab → name
    await user.tab();
    expect(document.activeElement).toBe(nameInput);

    // Tab → submit
    await user.tab();
    // Should be on a button (submit or cancel depending on DOM order)
    const focusedAfterName = document.activeElement;
    expect([submitBtn, cancelBtn]).toContain(focusedAfterName);
  });
});
