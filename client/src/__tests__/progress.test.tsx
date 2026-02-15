import { render } from '@testing-library/react';
import { Progress } from '@/components/ui/progress';

describe('Progress', () => {
  it('renders progress element', () => {
    const { container } = render(<Progress value={50} />);
    const progressBar = container.querySelector('.bg-primary');
    expect(progressBar).toBeInTheDocument();
  });

  it('displays correct value', () => {
    const { container } = render(<Progress value={75} />);
    const progressFill = container.querySelector('.bg-primary');
    expect(progressFill?.getAttribute('style')).toContain('width: 75%');
  });

  it('applies custom className', () => {
    const { container } = render(<Progress value={50} className="custom-progress" />);
    const progress = container.querySelector('.custom-progress');
    expect(progress).toBeInTheDocument();
  });
});
