"use client";

import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";

interface ExchangeMetrics {
  navigated_to_exchange: number;
  total_users: number;
  prompt_clicks: number;
  prompt_impressions: number;
  conversion_rate: number;
  traders_active_on_exchange: number;
  avg_trades_per_user: number;
  correlation_score: number;
}

const defaultMetrics: ExchangeMetrics = {
  navigated_to_exchange: 0,
  total_users: 0,
  prompt_clicks: 0,
  prompt_impressions: 0,
  conversion_rate: 0,
  traders_active_on_exchange: 0,
  avg_trades_per_user: 0,
  correlation_score: 0,
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

export default function ExchangeMetricsPage() {
  const { data: metrics = defaultMetrics, isLoading } = useQuery<ExchangeMetrics>({
    queryKey: ["admin-exchange-metrics"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/exchange/metrics");
      return res.data?.data ?? res.data ?? defaultMetrics;
    },
  });

  const clickRate =
    metrics.prompt_impressions > 0
      ? ((metrics.prompt_clicks / metrics.prompt_impressions) * 100).toFixed(1)
      : "0";

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-100 mb-6">Exchange Metrics</h1>

      {isLoading ? (
        <p className="text-gray-400 text-sm">Loading exchange metrics...</p>
      ) : (
        <>
          {/* Overview Stats */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
            <StatCard label="Navigated to Exchange" value={metrics.navigated_to_exchange} />
            <StatCard label="Total Users" value={metrics.total_users} />
            <StatCard label="Conversion Rate" value={`${metrics.conversion_rate.toFixed(1)}%`} />
            <StatCard label="Traders on Exchange" value={metrics.traders_active_on_exchange} />
          </div>

          {/* Prompt Performance */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
            <div className="bg-gray-800 border border-gray-700 rounded-lg p-6">
              <h3 className="text-sm font-semibold text-gray-300 mb-4 uppercase tracking-wider">
                Prompt Click-Through Rate
              </h3>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-300">Impressions</span>
                  <span className="text-sm font-mono text-gray-200">
                    {metrics.prompt_impressions.toLocaleString()}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-300">Clicks</span>
                  <span className="text-sm font-mono text-gray-200">
                    {metrics.prompt_clicks.toLocaleString()}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-blue-400">CTR</span>
                  <span className="text-sm font-mono text-blue-400">{clickRate}%</span>
                </div>
                <div className="w-full bg-gray-700 rounded-full h-3">
                  <div
                    className="bg-blue-500 h-3 rounded-full transition-all"
                    style={{ width: `${Math.min(Number(clickRate), 100)}%` }}
                  />
                </div>
              </div>
            </div>

            <div className="bg-gray-800 border border-gray-700 rounded-lg p-6">
              <h3 className="text-sm font-semibold text-gray-300 mb-4 uppercase tracking-wider">
                Trading Activity Correlation
              </h3>
              <div className="flex items-end gap-3 mb-4">
                <span className="text-4xl font-bold text-gray-100">
                  {metrics.correlation_score.toFixed(2)}
                </span>
                <span className="text-sm text-gray-400 mb-1">correlation score</span>
              </div>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-300">Avg Trades per User</span>
                  <span className="text-sm font-mono text-gray-200">
                    {metrics.avg_trades_per_user.toFixed(1)}
                  </span>
                </div>
              </div>
            </div>
          </div>

          {/* Charts Placeholder Row */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="bg-gray-800 border border-gray-700 rounded-lg p-6">
              <h3 className="text-sm font-semibold text-gray-300 mb-4 uppercase tracking-wider">
                Exchange Navigations (trend)
              </h3>
              <div className="h-48 flex items-center justify-center border border-dashed border-gray-600 rounded-md">
                <p className="text-sm text-gray-500">Chart placeholder - navigations over time</p>
              </div>
            </div>
            <div className="bg-gray-800 border border-gray-700 rounded-lg p-6">
              <h3 className="text-sm font-semibold text-gray-300 mb-4 uppercase tracking-wider">
                Conversion Funnel (trend)
              </h3>
              <div className="h-48 flex items-center justify-center border border-dashed border-gray-600 rounded-md">
                <p className="text-sm text-gray-500">Chart placeholder - conversion funnel over time</p>
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
