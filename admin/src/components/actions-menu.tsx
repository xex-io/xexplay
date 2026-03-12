"use client";

import { MoreHorizontal, type LucideIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu";

export interface ActionItem {
  label: string;
  icon?: LucideIcon;
  onClick: () => void;
  variant?: "default" | "destructive";
  disabled?: boolean;
}

interface ActionsMenuProps {
  items: ActionItem[];
}

export function ActionsMenu({ items }: ActionsMenuProps) {
  const regular = items.filter((i) => i.variant !== "destructive");
  const destructive = items.filter((i) => i.variant === "destructive");

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={<Button variant="ghost" size="icon-sm" />}
      >
        <MoreHorizontal className="size-4" />
        <span className="sr-only">Actions</span>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {regular.map((item) => (
          <DropdownMenuItem
            key={item.label}
            onClick={item.onClick}
            disabled={item.disabled}
          >
            {item.icon && <item.icon className="size-4" />}
            {item.label}
          </DropdownMenuItem>
        ))}
        {destructive.length > 0 && regular.length > 0 && (
          <DropdownMenuSeparator />
        )}
        {destructive.map((item) => (
          <DropdownMenuItem
            key={item.label}
            variant="destructive"
            onClick={item.onClick}
            disabled={item.disabled}
          >
            {item.icon && <item.icon className="size-4" />}
            {item.label}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
