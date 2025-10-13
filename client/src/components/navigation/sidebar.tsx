"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { useAuth } from "@/lib/auth-context";
import {
  GraduationCap,
  LayoutDashboard,
  Notebook,
  Users,
  BellRing,
  CalendarDays,
  BarChart3,
  FileText,
  ClipboardCheck,
  Award,
  BookOpen,
  Settings,
  Building2,
  UserCog,
} from "lucide-react";

type NavItem = {
  href: string;
  label: string;
  icon: any;
  roles?: string[];
};

const NAV_ITEMS: NavItem[] = [
  { href: "/", label: "Dashboard", icon: LayoutDashboard },
  { href: "/courses", label: "Courses", icon: Notebook },
  { href: "/assignments", label: "Assignments", icon: FileText, roles: ["student", "faculty"] },
  { href: "/quizzes", label: "Quizzes", icon: BookOpen },
  { href: "/attendance", label: "Attendance", icon: ClipboardCheck },
  { href: "/grades", label: "Grades", icon: Award },
  { href: "/announcements", label: "Announcements", icon: BellRing },
  { href: "/calendar", label: "Calendar", icon: CalendarDays },
  { href: "/students", label: "Students", icon: Users, roles: ["faculty", "admin"] },
  { href: "/departments", label: "Departments", icon: Building2, roles: ["admin"] },
  { href: "/analytics", label: "Analytics", icon: BarChart3, roles: ["faculty", "admin"] },
  { href: "/users", label: "Users", icon: UserCog, roles: ["admin"] },
  { href: "/settings", label: "Settings", icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();
  const { user } = useAuth();

  const filteredItems = NAV_ITEMS.filter(
    (item) => !item.roles || (user?.role && item.roles.includes(user.role))
  );

  return (
    <aside className="flex h-full w-64 flex-col border-r bg-background/50 backdrop-blur-sm">
      <div className="flex h-16 items-center border-b px-6">
        <Link href="/" className="flex items-center gap-2 text-lg font-semibold">
          <GraduationCap className="h-6 w-6" />
          edduhub
        </Link>
      </div>
      <nav className="flex flex-1 flex-col gap-1 p-4 text-sm font-medium">
        {filteredItems.map((item) => {
          const Icon = item.icon;
          const active = pathname === item.href;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center gap-3 rounded-lg px-3 py-2.5 transition-all duration-200 hover:bg-accent/50",
                active
                  ? "bg-accent text-accent-foreground shadow-sm"
                  : "text-muted-foreground hover:text-foreground"
              )}
            >
              <Icon className="h-4 w-4" />
              {item.label}
            </Link>
          );
        })}
      </nav>
    </aside>
  );
}
