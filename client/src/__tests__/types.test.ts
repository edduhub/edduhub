import { getDashboardPathForRole, UserRole } from '@/lib/types';

describe('getDashboardPathForRole', () => {
  it('returns /student-dashboard for student role', () => {
    expect(getDashboardPathForRole('student')).toBe('/student-dashboard');
  });

  it('returns / for faculty role', () => {
    expect(getDashboardPathForRole('faculty')).toBe('/');
  });

  it('returns / for admin role', () => {
    expect(getDashboardPathForRole('admin')).toBe('/');
  });

  it('returns / for super_admin role', () => {
    expect(getDashboardPathForRole('super_admin')).toBe('/');
  });

  it('returns /parent-portal for parent role', () => {
    expect(getDashboardPathForRole('parent')).toBe('/parent-portal');
  });

  it('returns / for unknown role (default case)', () => {
    expect(getDashboardPathForRole('unknown' as UserRole)).toBe('/');
  });
});
