"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { cn } from "@/lib/utils";
import { useAuth } from "@/lib/auth-context";
import { Separator } from "@/components/ui/separator";
import {
  LayoutDashboard,
  Calendar,
  Zap,
  Layers,
  Package,
  Users,
  BarChart3,
  Coins,
  Trophy,
  Bell,
  TrendingUp,
  ArrowRightLeft,
  UserPlus,
  Languages,
  Shield,
  Flag,
  ClipboardList,
  Bot,
  Settings,
  LogOut,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

interface NavItem {
  label: string;
  href: string;
  icon: LucideIcon;
}

interface NavSection {
  title?: string;
  items: NavItem[];
}

const navSections: NavSection[] = [
  {
    items: [
      { label: "Dashboard", href: "/", icon: LayoutDashboard },
      { label: "Events", href: "/events", icon: Calendar },
      { label: "Matches", href: "/matches", icon: Zap },
      { label: "Cards", href: "/cards", icon: Layers },
      { label: "Baskets", href: "/baskets", icon: Package },
      { label: "Users", href: "/users", icon: Users },
      { label: "Leaderboards", href: "/leaderboards", icon: BarChart3 },
      { label: "Rewards", href: "/rewards", icon: Coins },
      { label: "Prize Pools", href: "/prize-pools", icon: Trophy },
      { label: "Notifications", href: "/notifications", icon: Bell },
      { label: "Analytics", href: "/analytics", icon: TrendingUp },
      { label: "Exchange Metrics", href: "/exchange-metrics", icon: ArrowRightLeft },
    ],
  },
  {
    title: "Management",
    items: [
      { label: "Referrals", href: "/referrals", icon: UserPlus },
      { label: "Translations", href: "/translations", icon: Languages },
      { label: "Moderation", href: "/moderation", icon: Shield },
      { label: "Abuse Flags", href: "/abuse", icon: Flag },
      { label: "Audit Log", href: "/audit", icon: ClipboardList },
      { label: "Automation", href: "/automation", icon: Bot },
      { label: "Settings", href: "/settings", icon: Settings },
    ],
  },
];

export default function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const { user, logout } = useAuth();

  const handleLogout = () => {
    logout();
    router.push("/login");
  };

  return (
    <aside className="w-64 min-h-screen bg-sidebar text-sidebar-foreground flex flex-col border-r border-sidebar-border">
      <div className="px-6 py-5 text-xl font-bold tracking-wide">
        XEX Play
      </div>
      <Separator className="bg-sidebar-border" />
      <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
        {navSections.map((section, sIdx) => (
          <div key={sIdx}>
            {section.title && (
              <div className="px-3 pt-5 pb-2">
                <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                  {section.title}
                </p>
              </div>
            )}
            {section.items.map((item) => {
              const Icon = item.icon;
              const isActive =
                item.href === "/"
                  ? pathname === "/"
                  : pathname.startsWith(item.href);

              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={cn(
                    "flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors",
                    isActive
                      ? "bg-sidebar-primary text-sidebar-primary-foreground"
                      : "text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
                  )}
                >
                  <Icon className="h-5 w-5 shrink-0" />
                  {item.label}
                </Link>
              );
            })}
          </div>
        ))}
      </nav>

      <Separator className="bg-sidebar-border" />
      <div className="px-3 py-3">
        <div className="flex items-center gap-3 px-3 py-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-sidebar-primary text-sidebar-primary-foreground text-xs font-bold">
            {user?.display_name?.[0]?.toUpperCase() || user?.email?.[0]?.toUpperCase() || "A"}
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium truncate">
              {user?.display_name || "Admin"}
            </p>
            <p className="text-xs text-muted-foreground truncate">
              {user?.email || ""}
            </p>
          </div>
          <button
            onClick={handleLogout}
            className="p-1.5 rounded-md text-muted-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground transition-colors"
            title="Sign out"
          >
            <LogOut className="h-4 w-4" />
          </button>
        </div>
      </div>
    </aside>
  );
}
