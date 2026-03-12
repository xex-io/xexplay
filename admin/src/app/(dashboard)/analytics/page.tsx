"use client";

import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";

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

const defaultOverview: AnalyticsOverview = {
  total_users: 0,
  dau: 0,
  wau: 0,
  mau: 0,
  total_sessions: 0,
  completed_sessions: 0,
  correct_answers: 0,
  incorrect_answers: 0,
};

function StatCard({ label, value }: { label: string; value: string | number }) {
  return (
    <Card>
      <CardContent className="pt-1">
        <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-1">
          {label}
        </p>
        <p className="text-2xl font-bold text-foreground" suppressHydrationWarning>
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
      const d = res.data?.data ?? res.data;
      return { ...defaultOverview, ...(d && typeof d === "object" && !Array.isArray(d) ? d : {}) };
    },
  });

  const completionRate =
    (overview.total_sessions ?? 0) > 0
      ? ((overview.completed_sessions ?? 0) / overview.total_sessions) * 100
      : 0;

  const totalAnswers = (overview.correct_answers ?? 0) + (overview.incorrect_answers ?? 0);
  const correctPct = totalAnswers > 0 ? (((overview.correct_answers ?? 0) / totalAnswers) * 100).toFixed(1) : "0";
  const incorrectPct = totalAnswers > 0 ? (((overview.incorrect_answers ?? 0) / totalAnswers) * 100).toFixed(1) : "0";

  return (
    <div>
      <h1 className="text-2xl font-bold text-foreground mb-6">Analytics</h1>

      {isLoading ? (
        <p className="text-muted-foreground text-sm">Loading analytics...</p>
      ) : (
        <>
          {/* Stats Cards */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
            <StatCard label="Total Users" value={overview.total_users ?? 0} />
            <StatCard label="DAU" value={overview.dau ?? 0} />
            <StatCard label="WAU" value={overview.wau ?? 0} />
            <StatCard label="MAU" value={overview.mau ?? 0} />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
            <StatCard label="Total Sessions" value={overview.total_sessions ?? 0} />
            <StatCard label="Completed Sessions" value={overview.completed_sessions ?? 0} />
            <StatCard label="Correct Answers" value={overview.correct_answers ?? 0} />
            <StatCard label="Incorrect Answers" value={overview.incorrect_answers ?? 0} />
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
                    {completionRate.toFixed(1)}%
                  </span>
                  <span className="text-sm text-muted-foreground mb-1">
                    {(overview.completed_sessions ?? 0).toLocaleString()} of {(overview.total_sessions ?? 0).toLocaleString()} sessions completed
                  </span>
                </div>
                <div className="mt-4 w-full bg-muted rounded-full h-3">
                  <div
                    className="bg-primary h-3 rounded-full transition-all"
                    style={{ width: `${Math.min(completionRate, 100)}%` }}
                  />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-sm uppercase tracking-wider text-muted-foreground">
                  Answer Distribution
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-muted-foreground">Total Answers</span>
                    <span className="text-sm font-mono text-foreground">
                      {totalAnswers.toLocaleString()}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-green-400">Correct</span>
                    <span className="text-sm font-mono text-green-400">
                      {(overview.correct_answers ?? 0).toLocaleString()} ({correctPct}%)
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-red-400">Incorrect</span>
                    <span className="text-sm font-mono text-red-400">
                      {(overview.incorrect_answers ?? 0).toLocaleString()} ({incorrectPct}%)
                    </span>
                  </div>
                  <div className="w-full bg-muted rounded-full h-3 flex overflow-hidden">
                    {totalAnswers > 0 && (
                      <>
                        <div
                          className="bg-green-500 h-3 transition-all"
                          style={{ width: `${correctPct}%` }}
                        />
                        <div
                          className="bg-red-500 h-3 transition-all"
                          style={{ width: `${incorrectPct}%` }}
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
