"use client";

import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";

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
    <div className="bg-gray-800 border border-gray-700 rounded-lg p-5">
      <p className="text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">{label}</p>
      <p className="text-2xl font-bold text-gray-100">
        {typeof value === "number" ? value.toLocaleString() : value}
      </p>
    </div>
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
      <h1 className="text-2xl font-bold text-gray-100 mb-6">Analytics</h1>

      {isLoading ? (
        <p className="text-gray-400 text-sm">Loading analytics...</p>
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
            <div className="bg-gray-800 border border-gray-700 rounded-lg p-6">
              <h3 className="text-sm font-semibold text-gray-300 mb-4 uppercase tracking-wider">
                Daily Active Users (trend)
              </h3>
              <div className="h-48 flex items-center justify-center border border-dashed border-gray-600 rounded-md">
                <p className="text-sm text-gray-500">Chart placeholder - DAU over time</p>
              </div>
            </div>
            <div className="bg-gray-800 border border-gray-700 rounded-lg p-6">
              <h3 className="text-sm font-semibold text-gray-300 mb-4 uppercase tracking-wider">
                New Registrations (trend)
              </h3>
              <div className="h-48 flex items-center justify-center border border-dashed border-gray-600 rounded-md">
                <p className="text-sm text-gray-500">Chart placeholder - registrations over time</p>
              </div>
            </div>
          </div>

          {/* Session & Answer Stats */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="bg-gray-800 border border-gray-700 rounded-lg p-6">
              <h3 className="text-sm font-semibold text-gray-300 mb-4 uppercase tracking-wider">
                Session Completion Rate
              </h3>
              <div className="flex items-end gap-3">
                <span className="text-4xl font-bold text-gray-100">
                  {overview.session_completion_rate.toFixed(1)}%
                </span>
                <span className="text-sm text-gray-400 mb-1">of sessions completed</span>
              </div>
              <div className="mt-4 w-full bg-gray-700 rounded-full h-3">
                <div
                  className="bg-blue-500 h-3 rounded-full transition-all"
                  style={{ width: `${Math.min(overview.session_completion_rate, 100)}%` }}
                />
              </div>
            </div>

            <div className="bg-gray-800 border border-gray-700 rounded-lg p-6">
              <h3 className="text-sm font-semibold text-gray-300 mb-4 uppercase tracking-wider">
                Card Answer Distribution
              </h3>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-300">Total Answers</span>
                  <span className="text-sm font-mono text-gray-200">{dist.total.toLocaleString()}</span>
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
                <div className="w-full bg-gray-700 rounded-full h-3 flex overflow-hidden">
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
            </div>
          </div>
        </>
      )}
    </div>
  );
}
