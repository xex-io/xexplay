"use client";

import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";

interface AnalyticsOverview {
  dau: number;
  wau: number;
  mau: number;
  total_users: number;
  active_sessions: number;
  session_completion_rate: number;
  card_answer_distribution: {
    total: number;
    correct: number;
    wrong: number;
  };
}

const defaultOverview: AnalyticsOverview = {
  dau: 0,
  wau: 0,
  mau: 0,
  total_users: 0,
  active_sessions: 0,
  session_completion_rate: 0,
  card_answer_distribution: { total: 0, correct: 0, wrong: 0 },
};

function StatCard({ label, value }: { label: string; value: string | number }) {
  return (
    <Card>
      <CardContent className="pt-1">
        <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-1">
          {label}
        </p>
        <p className="text-2xl font-bold text-foreground">
          {typeof value === "number" ? value.toLocaleString() : value}
        </p>
      </CardContent>
    </Card>
  );
}

export default function AnalyticsPage() {
  const { data: overview = defaultOverview, isLoading } = useQuery<AnalyticsOverview>({
    queryKey: ["admin-analytics-overview"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/analytics/overview");
      return res.data?.data ?? res.data ?? defaultOverview;
    },
  });

  const dist = overview.card_answer_distribution;
  const correctPct = dist.total > 0 ? ((dist.correct / dist.total) * 100).toFixed(1) : "0";
  const wrongPct = dist.total > 0 ? ((dist.wrong / dist.total) * 100).toFixed(1) : "0";

  return (
    <div>
      <h1 className="text-2xl font-bold text-foreground mb-6">Analytics</h1>

      {isLoading ? (
        <p className="text-muted-foreground text-sm">Loading analytics...</p>
      ) : (
        <>
          {/* Stats Cards */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4 mb-8">
            <StatCard label="DAU" value={overview.dau} />
            <StatCard label="WAU" value={overview.wau} />
            <StatCard label="MAU" value={overview.mau} />
            <StatCard label="Total Users" value={overview.total_users} />
            <StatCard label="Active Sessions" value={overview.active_sessions} />
          </div>

          {/* Charts Placeholder Row */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
            <Card>
              <CardHeader>
                <CardTitle className="text-sm uppercase tracking-wider text-muted-foreground">
                  Daily Active Users (trend)
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="h-48 flex items-center justify-center border border-dashed border-border rounded-md">
                  <p className="text-sm text-muted-foreground">
                    Chart placeholder - DAU over time
                  </p>
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle className="text-sm uppercase tracking-wider text-muted-foreground">
                  New Registrations (trend)
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="h-48 flex items-center justify-center border border-dashed border-border rounded-md">
                  <p className="text-sm text-muted-foreground">
                    Chart placeholder - registrations over time
                  </p>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Session & Answer Stats */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card>
              <CardHeader>
                <CardTitle className="text-sm uppercase tracking-wider text-muted-foreground">
                  Session Completion Rate
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex items-end gap-3">
                  <span className="text-4xl font-bold text-foreground">
                    {overview.session_completion_rate.toFixed(1)}%
                  </span>
                  <span className="text-sm text-muted-foreground mb-1">
                    of sessions completed
                  </span>
                </div>
                <div className="mt-4 w-full bg-muted rounded-full h-3">
                  <div
                    className="bg-primary h-3 rounded-full transition-all"
                    style={{ width: `${Math.min(overview.session_completion_rate, 100)}%` }}
                  />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-sm uppercase tracking-wider text-muted-foreground">
                  Card Answer Distribution
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-muted-foreground">Total Answers</span>
                    <span className="text-sm font-mono text-foreground">
                      {dist.total.toLocaleString()}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-green-400">Correct</span>
                    <span className="text-sm font-mono text-green-400">
                      {dist.correct.toLocaleString()} ({correctPct}%)
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-red-400">Wrong</span>
                    <span className="text-sm font-mono text-red-400">
                      {dist.wrong.toLocaleString()} ({wrongPct}%)
                    </span>
                  </div>
                  <div className="w-full bg-muted rounded-full h-3 flex overflow-hidden">
                    {dist.total > 0 && (
                      <>
                        <div
                          className="bg-green-500 h-3 transition-all"
                          style={{ width: `${correctPct}%` }}
                        />
                        <div
                          className="bg-red-500 h-3 transition-all"
                          style={{ width: `${wrongPct}%` }}
                        />
                      </>
                    )}
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </>
      )}
    </div>
  );
}
