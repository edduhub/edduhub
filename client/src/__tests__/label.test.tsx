import { render, screen } from '@testing-library/react';
import { Label } from '@/components/ui/label';

describe('Label', () => {
  it('renders label element', () => {
    render(<Label>Label text</Label>);
    expect(screen.getByText('Label text')).toBeInTheDocument();
  });

  it('renders as label element', () => {
    render(<Label>Text</Label>);
    expect(screen.getByText('Text').tagName.toLowerCase()).toBe('label');
  });

  it('associates with input via htmlFor', () => {
    render(<Label htmlFor="test-input">Email</Label>);
    expect(screen.getByText('Email')).toHaveAttribute('for', 'test-input');
  });

  it('applies custom className', () => {
    render(<Label className="custom-label">Custom</Label>);
    expect(screen.getByText('Custom')).toHaveClass('custom-label');
  });
});
