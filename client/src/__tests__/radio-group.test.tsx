import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';

describe('RadioGroup', () => {
  it('renders radio group', () => {
    render(
      <RadioGroup>
        <RadioGroupItem value="option1" />
        <RadioGroupItem value="option2" />
      </RadioGroup>
    );
    expect(screen.getByRole('radiogroup')).toBeInTheDocument();
  });

  it('renders radio items', () => {
    render(
      <RadioGroup>
        <RadioGroupItem value="option1" id="r1" />
        <RadioGroupItem value="option2" id="r2" />
      </RadioGroup>
    );
    expect(screen.getAllByRole('radio').length).toBe(2);
  });

  it('selects on click', async () => {
    const user = userEvent.setup();
    render(
      <RadioGroup defaultValue="option1">
        <RadioGroupItem value="option1" id="r1" />
        <RadioGroupItem value="option2" id="r2" />
      </RadioGroup>
    );
    
    await user.click(screen.getAllByRole('radio')[1]);
    expect(screen.getAllByRole('radio')[1]).toBeChecked();
  });
});
