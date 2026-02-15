import { render } from '@testing-library/react';
import { Separator } from '@/components/ui/separator';

describe('Separator', () => {
  it('renders separator element', () => {
    const { container } = render(<Separator />);
    const separator = container.querySelector('.bg-border');
    expect(separator).toBeInTheDocument();
  });

  it('renders with custom className', () => {
    const { container } = render(<Separator className="custom-separator" />);
    const separator = container.querySelector('.custom-separator');
    expect(separator).toBeInTheDocument();
  });
});
