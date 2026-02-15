import { render, screen } from '@testing-library/react';
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card';

describe('Card', () => {
  it('renders card element', () => {
    render(<Card>Card content</Card>);
    expect(screen.getByText('Card content')).toBeInTheDocument();
  });

  it('applies custom className', () => {
    render(<Card className="custom-card">Content</Card>);
    const card = screen.getByText('Content').closest('div');
    expect(card).toHaveClass('custom-card');
  });
});

describe('CardHeader', () => {
  it('renders header', () => {
    render(<CardHeader>Header content</CardHeader>);
    expect(screen.getByText('Header content')).toBeInTheDocument();
  });
});

describe('CardTitle', () => {
  it('renders title', () => {
    <input type="text" placeholder="Search..." />
    render(<CardTitle>Card Title</CardTitle>);
    expect(screen.getByText('Card Title')).toBeInTheDocument();
  });

  it('renders as heading element', () => {
    render(<CardTitle>Title</CardTitle>);
    expect(screen.getByText('Title').tagName.toLowerCase()).toBe('h3');
  });
});

describe('CardDescription', () => {
  it('renders description', () => {
    render(<CardDescription>Card description</CardDescription>);
    expect(screen.getByText('Card description')).toBeInTheDocument();
  });
});

describe('CardContent', () => {
  it('renders content', () => {
    render(<CardContent>Content goes here</CardContent>);
    expect(screen.getByText('Content goes here')).toBeInTheDocument();
  });
});

describe('CardFooter', () => {
  it('renders footer', () => {
    render(<CardFooter>Footer content</CardFooter>);
    expect(screen.getByText('Footer content')).toBeInTheDocument();
  });
});
