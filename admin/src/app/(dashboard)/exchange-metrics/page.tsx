"use client";

import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";

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
      <h1 className="text-2xl font-bold text-foreground mb-6">Exchange Metrics</h1>

      {isLoading ? (
        <p className="text-muted-foreground text-sm">Loading exchange metrics...</p>
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
            <Card>
              <CardHeader>
                <CardTitle className="text-sm uppercase tracking-wider text-muted-foreground">
                  Prompt Click-Through Rate
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-muted-foreground">Impressions</span>
                    <span className="text-sm font-mono text-foreground">
                      {metrics.prompt_impressions.toLocaleString()}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-muted-foreground">Clicks</span>
                    <span className="text-sm font-mono text-foreground">
                      {metrics.prompt_clicks.toLocaleString()}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-primary">CTR</span>
                    <span className="text-sm font-mono text-primary">{clickRate}%</span>
                  </div>
                  <div className="w-full bg-muted rounded-full h-3">
                    <div
                      className="bg-primary h-3 rounded-full transition-all"
                      style={{ width: `${Math.min(Number(clickRate), 100)}%` }}
                    />
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-sm uppercase tracking-wider text-muted-foreground">
                  Trading Activity Correlation
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex items-end gap-3 mb-4">
                  <span className="text-4xl font-bold text-foreground">
                    {metrics.correlation_score.toFixed(2)}
                  </span>
                  <span className="text-sm text-muted-foreground mb-1">correlation score</span>
                </div>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-muted-foreground">Avg Trades per User</span>
                    <span className="text-sm font-mono text-foreground">
                      {metrics.avg_trades_per_user.toFixed(1)}
                    </span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Charts Placeholder Row */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card>
              <CardHeader>
                <CardTitle className="text-sm uppercase tracking-wider text-muted-foreground">
                  Exchange Navigations (trend)
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="h-48 flex items-center justify-center border border-dashed border-border rounded-md">
                  <p className="text-sm text-muted-foreground">
                    Chart placeholder - navigations over time
                  </p>
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle className="text-sm uppercase tracking-wider text-muted-foreground">
                  Conversion Funnel (trend)
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="h-48 flex items-center justify-center border border-dashed border-border rounded-md">
                  <p className="text-sm text-muted-foreground">
                    Chart placeholder - conversion funnel over time
                  </p>
                </div>
              </CardContent>
            </Card>
          </div>
        </>
      )}
    </div>
  );
}
