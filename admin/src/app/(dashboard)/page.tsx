"use client";

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Users, ListChecks, CheckCircle, Target } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";

interface AnalyticsOverview {
  total_users: number;
  dau: number;
  wau: number;
  mau: number;
  total_sessions: number;
  completed_sessions: number;
  correct_answers: number;
  incorrect_answers: number;
}

function formatNumber(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return n.toLocaleString();
}

function formatPercent(numerator: number, denominator: number): string {
  if (denominator === 0) return "0%";
  return `${((numerator / denominator) * 100).toFixed(1)}%`;
}

export default function DashboardPage() {
  const { data, isLoading } = useQuery<AnalyticsOverview>({
    queryKey: ["admin", "analytics", "overview"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/analytics/overview");
      return res.data.data;
    },
  });

  const stats = [
    {
      label: "Total Users",
      value: data ? formatNumber(data.total_users) : "--",
      icon: Users,
    },
    {
      label: "Total Sessions",
      value: data ? formatNumber(data.total_sessions) : "--",
      icon: ListChecks,
    },
    {
      label: "Completion Rate",
      value: data
        ? formatPercent(data.completed_sessions, data.total_sessions)
        : "--",
      icon: CheckCircle,
    },
    {
      label: "Correct Answer Rate",
      value: data
        ? formatPercent(
            data.correct_answers,
            data.correct_answers + data.incorrect_answers
          )
        : "--",
      icon: Target,
    },
  ];

  return (
    <div>
      <h1 className="text-2xl font-bold text-foreground mb-6">
        XEX Play Admin
      </h1>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
        {stats.map((stat) => (
          <Card key={stat.label}>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {stat.label}
                </CardTitle>
                <stat.icon className="size-4 text-muted-foreground" />
              </div>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <div className="h-9 w-24 animate-pulse rounded bg-muted" />
              ) : (
                <p className="text-3xl font-semibold text-foreground">
                  {stat.value}
                </p>
              )}
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
