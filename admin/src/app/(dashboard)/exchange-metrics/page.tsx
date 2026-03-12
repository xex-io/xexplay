"use client";

import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

interface ExchangeMetrics {
  linked_users: number;
  total_users: number;
  link_rate: number;
  trading_tier_distribution: Record<string, number>;
  reward_claims_by_type: Record<string, number>;
}

const defaultMetrics: ExchangeMetrics = {
  linked_users: 0,
  total_users: 0,
  link_rate: 0,
  trading_tier_distribution: {},
  reward_claims_by_type: {},
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

const TIER_COLORS: Record<string, string> = {
  none: "bg-gray-500",
  bronze: "bg-amber-700",
  silver: "bg-gray-400",
  gold: "bg-yellow-500",
  platinum: "bg-cyan-400",
  diamond: "bg-blue-500",
};

export default function ExchangeMetricsPage() {
  const { data: metrics = defaultMetrics, isLoading } = useQuery<ExchangeMetrics>({
    queryKey: ["admin-exchange-metrics"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/exchange/metrics");
      const d = res.data?.data ?? res.data;
      return { ...defaultMetrics, ...(d && typeof d === "object" && !Array.isArray(d) ? d : {}) };
    },
  });

  const tierEntries = Object.entries(metrics.trading_tier_distribution ?? {});
  const totalTierUsers = tierEntries.reduce((sum, [, count]) => sum + count, 0);

  const rewardEntries = Object.entries(metrics.reward_claims_by_type ?? {});
  const totalRewardClaims = rewardEntries.reduce((sum, [, count]) => sum + count, 0);

  return (
    <div>
      <h1 className="text-2xl font-bold text-foreground mb-6">Exchange Metrics</h1>

      {isLoading ? (
        <p className="text-muted-foreground text-sm">Loading exchange metrics...</p>
      ) : (
        <>
          {/* Overview Stats */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
            <StatCard label="Linked Users" value={metrics.linked_users ?? 0} />
            <StatCard label="Total Users" value={metrics.total_users ?? 0} />
            <StatCard label="Link Rate" value={`${((metrics.link_rate ?? 0) * 100).toFixed(1)}%`} />
          </div>

          {/* Trading Tier Distribution & Reward Claims */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
            <Card>
              <CardHeader>
                <CardTitle className="text-sm uppercase tracking-wider text-muted-foreground">
                  Trading Tier Distribution
                </CardTitle>
              </CardHeader>
              <CardContent>
                {tierEntries.length === 0 ? (
                  <p className="text-sm text-muted-foreground">No tier data available.</p>
                ) : (
                  <div className="space-y-3">
                    {tierEntries.map(([tier, count]) => {
                      const pct = totalTierUsers > 0 ? (count / totalTierUsers) * 100 : 0;
                      return (
                        <div key={tier}>
                          <div className="flex items-center justify-between mb-1">
                            <span className="text-sm capitalize text-foreground flex items-center gap-2">
                              <Badge variant="secondary" className="capitalize">
                                {tier}
                              </Badge>
                            </span>
                            <span className="text-sm font-mono text-muted-foreground">
                              {count.toLocaleString()} ({pct.toFixed(1)}%)
                            </span>
                          </div>
                          <div className="w-full bg-muted rounded-full h-2">
                            <div
                              className={`${TIER_COLORS[tier] ?? "bg-primary"} h-2 rounded-full transition-all`}
                              style={{ width: `${Math.min(pct, 100)}%` }}
                            />
                          </div>
                        </div>
                      );
                    })}
                    <div className="flex items-center justify-between pt-2 border-t border-border">
                      <span className="text-sm font-medium text-foreground">Total</span>
                      <span className="text-sm font-mono text-foreground">
                        {totalTierUsers.toLocaleString()}
                      </span>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-sm uppercase tracking-wider text-muted-foreground">
                  Reward Claims by Type
                </CardTitle>
              </CardHeader>
              <CardContent>
                {rewardEntries.length === 0 ? (
                  <p className="text-sm text-muted-foreground">No reward claims data available.</p>
                ) : (
                  <div className="space-y-3">
                    {rewardEntries.map(([type, count]) => {
                      const pct = totalRewardClaims > 0 ? (count / totalRewardClaims) * 100 : 0;
                      return (
                        <div key={type}>
                          <div className="flex items-center justify-between mb-1">
                            <span className="text-sm text-foreground">
                              <Badge variant="outline" className="capitalize">
                                {type.replace(/_/g, " ")}
                              </Badge>
                            </span>
                            <span className="text-sm font-mono text-muted-foreground">
                              {count.toLocaleString()} ({pct.toFixed(1)}%)
                            </span>
                          </div>
                          <div className="w-full bg-muted rounded-full h-2">
                            <div
                              className="bg-primary h-2 rounded-full transition-all"
                              style={{ width: `${Math.min(pct, 100)}%` }}
                            />
                          </div>
                        </div>
                      );
                    })}
                    <div className="flex items-center justify-between pt-2 border-t border-border">
                      <span className="text-sm font-medium text-foreground">Total Claims</span>
                      <span className="text-sm font-mono text-foreground">
                        {totalRewardClaims.toLocaleString()}
                      </span>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Link Rate Visual */}
          <Card>
            <CardHeader>
              <CardTitle className="text-sm uppercase tracking-wider text-muted-foreground">
                Exchange Link Rate
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-end gap-3 mb-4">
                <span className="text-4xl font-bold text-foreground">
                  {((metrics.link_rate ?? 0) * 100).toFixed(1)}%
                </span>
                <span className="text-sm text-muted-foreground mb-1">
                  {(metrics.linked_users ?? 0).toLocaleString()} of {(metrics.total_users ?? 0).toLocaleString()} users linked to exchange
                </span>
              </div>
              <div className="w-full bg-muted rounded-full h-3">
                <div
                  className="bg-primary h-3 rounded-full transition-all"
                  style={{ width: `${Math.min((metrics.link_rate ?? 0) * 100, 100)}%` }}
                />
              </div>
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
