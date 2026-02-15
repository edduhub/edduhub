import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Switch } from '@/components/ui/switch';

describe('Switch', () => {
  it('renders switch element', () => {
    render(<Switch />);
    expect(screen.getByRole('switch')).toBeInTheDocument();
  });

  it('toggles on click', async () => {
    const user = userEvent.setup();
    const handleChange = jest.fn();

    render(<Switch onCheckedChange={handleChange} />);
    await user.click(screen.getByRole('switch'));

    expect(handleChange).toHaveBeenCalledWith(true);
  });

  it('can be checked', () => {
    render(<Switch checked />);
    expect(screen.getByRole('switch')).toBeChecked();
  });

  it('can be disabled', () => {
    render(<Switch disabled />);
    expect(screen.getByRole('switch')).toBeDisabled();
  });
});
