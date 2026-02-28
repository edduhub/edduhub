import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogFooter,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog';

describe('DialogHeader', () => {
  it('renders children', () => {
    render(<DialogHeader>Header Content</DialogHeader>);
    expect(screen.getByText('Header Content')).toBeInTheDocument();
  });

  it('applies custom className', () => {
    render(<DialogHeader className="my-header" data-testid="header">Test</DialogHeader>);
    expect(screen.getByTestId('header')).toHaveClass('my-header');
  });
});

describe('DialogFooter', () => {
  it('renders children', () => {
    render(<DialogFooter>Footer Content</DialogFooter>);
    expect(screen.getByText('Footer Content')).toBeInTheDocument();
  });

  it('applies custom className', () => {
    render(<DialogFooter className="my-footer" data-testid="footer">Test</DialogFooter>);
    expect(screen.getByTestId('footer')).toHaveClass('my-footer');
  });
});

describe('Dialog composition', () => {
  it('renders dialog trigger', () => {
    render(
      <Dialog>
        <DialogTrigger asChild>
          <button>Open Dialog</button>
        </DialogTrigger>
      </Dialog>
    );
    expect(screen.getByText('Open Dialog')).toBeInTheDocument();
  });
});
