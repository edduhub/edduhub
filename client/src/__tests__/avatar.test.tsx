import { render, screen } from '@testing-library/react';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';

describe('Avatar', () => {
  it('renders avatar element', () => {
    render(<Avatar />);
    const avatar = document.querySelector('.rounded-full');
    expect(avatar).toBeInTheDocument();
  });

  it('renders with fallback', () => {
    render(
      <Avatar>
        <AvatarFallback>AB</AvatarFallback>
      </Avatar>
    );
    expect(screen.getByText('AB')).toBeInTheDocument();
  });

  it('renders with image', () => {
    render(
      <Avatar>
        <AvatarImage src="https://example.com/avatar.png" alt="Avatar" />
      </Avatar>
    );
    const avatar = document.querySelector('.rounded-full');
    expect(avatar).toBeInTheDocument();
  });

  it('applies custom className', () => {
    render(<Avatar className="custom-avatar" />);
    const avatar = document.querySelector('.custom-avatar');
    expect(avatar).toBeInTheDocument();
  });
});
