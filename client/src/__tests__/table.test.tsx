import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table';

describe('Table', () => {
  it('renders a table element', () => {
    render(<Table data-testid="table" />);
    const el = screen.getByTestId('table');
    expect(el.tagName).toBe('TABLE');
  });

  it('applies custom className', () => {
    render(<Table className="my-table" data-testid="table" />);
    expect(screen.getByTestId('table')).toHaveClass('my-table');
  });

  it('forwards ref', () => {
    const ref = React.createRef<HTMLTableElement>();
    render(<Table ref={ref} />);
    expect(ref.current).toBeInstanceOf(HTMLTableElement);
  });
});

describe('TableHeader', () => {
  it('renders a thead element', () => {
    render(
      <table>
        <TableHeader data-testid="thead" />
      </table>
    );
    expect(screen.getByTestId('thead').tagName).toBe('THEAD');
  });

  it('applies custom className', () => {
    render(
      <table>
        <TableHeader className="my-header" data-testid="thead" />
      </table>
    );
    expect(screen.getByTestId('thead')).toHaveClass('my-header');
  });
});

describe('TableBody', () => {
  it('renders a tbody element', () => {
    render(
      <table>
        <TableBody data-testid="tbody" />
      </table>
    );
    expect(screen.getByTestId('tbody').tagName).toBe('TBODY');
  });

  it('applies custom className', () => {
    render(
      <table>
        <TableBody className="my-body" data-testid="tbody" />
      </table>
    );
    expect(screen.getByTestId('tbody')).toHaveClass('my-body');
  });
});

describe('TableRow', () => {
  it('renders a tr element', () => {
    render(
      <table>
        <tbody>
          <TableRow data-testid="tr" />
        </tbody>
      </table>
    );
    expect(screen.getByTestId('tr').tagName).toBe('TR');
  });

  it('applies custom className', () => {
    render(
      <table>
        <tbody>
          <TableRow className="my-row" data-testid="tr" />
        </tbody>
      </table>
    );
    expect(screen.getByTestId('tr')).toHaveClass('my-row');
  });
});

describe('TableHead', () => {
  it('renders a th element', () => {
    render(
      <table>
        <thead>
          <tr>
            <TableHead data-testid="th">Name</TableHead>
          </tr>
        </thead>
      </table>
    );
    expect(screen.getByTestId('th').tagName).toBe('TH');
    expect(screen.getByTestId('th')).toHaveTextContent('Name');
  });

  it('applies custom className', () => {
    render(
      <table>
        <thead>
          <tr>
            <TableHead className="my-head" data-testid="th" />
          </tr>
        </thead>
      </table>
    );
    expect(screen.getByTestId('th')).toHaveClass('my-head');
  });
});

describe('TableCell', () => {
  it('renders a td element', () => {
    render(
      <table>
        <tbody>
          <tr>
            <TableCell data-testid="td">Value</TableCell>
          </tr>
        </tbody>
      </table>
    );
    expect(screen.getByTestId('td').tagName).toBe('TD');
    expect(screen.getByTestId('td')).toHaveTextContent('Value');
  });

  it('applies custom className', () => {
    render(
      <table>
        <tbody>
          <tr>
            <TableCell className="my-cell" data-testid="td" />
          </tr>
        </tbody>
      </table>
    );
    expect(screen.getByTestId('td')).toHaveClass('my-cell');
  });
});

describe('Table composition', () => {
  it('renders a full table with data', () => {
    render(
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Email</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow>
            <TableCell>Alice</TableCell>
            <TableCell>alice@example.com</TableCell>
          </TableRow>
          <TableRow>
            <TableCell>Bob</TableCell>
            <TableCell>bob@example.com</TableCell>
          </TableRow>
        </TableBody>
      </Table>
    );

    expect(screen.getByText('Name')).toBeInTheDocument();
    expect(screen.getByText('Email')).toBeInTheDocument();
    expect(screen.getByText('Alice')).toBeInTheDocument();
    expect(screen.getByText('bob@example.com')).toBeInTheDocument();
  });
});
