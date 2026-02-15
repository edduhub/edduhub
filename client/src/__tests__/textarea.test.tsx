import { render, screen } from '@testing-library/react';
import { Textarea } from '@/components/ui/textarea';

describe('Textarea', () => {
  it('renders textarea element', () => {
    render(<Textarea />);
    expect(screen.getByRole('textbox')).toBeInTheDocument();
  });

  it('renders with placeholder', () => {
    render(<Textarea placeholder="Enter text" />);
    expect(screen.getByPlaceholderText('Enter text')).toBeInTheDocument();
  });

  it('renders with value', () => {
    render(<Textarea value="test value" readOnly />);
    expect(screen.getByRole('textbox')).toHaveValue('test value');
  });

  it('can be disabled', () => {
    render(<Textarea disabled />);
    expect(screen.getByRole('textbox')).toBeDisabled();
  });

  it('applies custom className', () => {
    render(<Textarea className="custom-textarea" />);
    expect(screen.getByRole('textbox')).toHaveClass('custom-textarea');
  });
});
