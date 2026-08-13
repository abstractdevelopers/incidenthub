import { render, screen } from '@testing-library/react';
import { StatusBadge, SeverityBadge } from '../Badges';

describe('Badges', () => {
  it('renders a status badge', () => {
    render(<StatusBadge status="OPEN" />);
    expect(screen.getByText('OPEN')).toBeInTheDocument();
  });

  it('renders a severity badge', () => {
    render(<SeverityBadge severity="CRITICAL" />);
    expect(screen.getByText('CRITICAL')).toBeInTheDocument();
  });
});
