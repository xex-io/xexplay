"use client";

import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

interface ReferralStats {
  total_referrals: number;
  converted_referrals: number;
  conversion_rate: number;
  active_referrers: number;
}

interface TopReferrer {
  user_id: string;
  email: string;
  display_name: string;
  referral_count: number;
  converted_count: number;
}

export default function ReferralsPage() {
  const { data: stats, isLoading: statsLoading } = useQuery<ReferralStats>({
    queryKey: ["admin-referral-stats"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/referrals/stats");
      return res.data?.data ?? res.data;
    },
  });

  const { data: topReferrers = [], isLoading: referrersLoading } = useQuery<TopReferrer[]>({
    queryKey: ["admin-referral-top"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/referrals/top");
      return res.data?.data ?? res.data ?? [];
    },
  });

  const loading = statsLoading || referrersLoading;

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-foreground">
        Referral Analytics
      </h1>

      {/* Stats cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard
          label="Total Referrals"
          value={stats?.total_referrals ?? 0}
          loading={statsLoading}
        />
        <StatCard
          label="Converted Referrals"
          value={stats?.converted_referrals ?? 0}
          loading={statsLoading}
        />
        <StatCard
          label="Conversion Rate"
          value={`${(stats?.conversion_rate ?? 0).toFixed(1)}%`}
          loading={statsLoading}
        />
        <StatCard
          label="Active Referrers"
          value={stats?.active_referrers ?? 0}
          loading={statsLoading}
        />
      </div>

      {/* Top referrers table */}
      <Card>
        <CardHeader className="border-b">
          <CardTitle>Top Referrers</CardTitle>
        </CardHeader>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>User</TableHead>
              <TableHead>Email</TableHead>
              <TableHead>Referral Count</TableHead>
              <TableHead>Converted</TableHead>
              <TableHead>Rate</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell
                  colSpan={5}
                  className="text-center py-8 text-muted-foreground"
                >
                  Loading...
                </TableCell>
              </TableRow>
            ) : topReferrers.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={5}
                  className="text-center py-8 text-muted-foreground"
                >
                  No referral data yet
                </TableCell>
              </TableRow>
            ) : (
              topReferrers.map((r) => {
                const rate =
                  r.referral_count > 0
                    ? (r.converted_count / r.referral_count) * 100
                    : 0;
                return (
                  <TableRow key={r.user_id}>
                    <TableCell className="font-medium">
                      {r.display_name || r.user_id}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {r.email}
                    </TableCell>
                    <TableCell>{r.referral_count}</TableCell>
                    <TableCell>{r.converted_count}</TableCell>
                    <TableCell>
                      <Badge
                        variant={rate >= 50 ? "default" : "secondary"}
                      >
                        {rate.toFixed(1)}%
                      </Badge>
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>
      </Card>

      {/* Referral trend placeholder */}
      <Card>
        <CardHeader>
          <CardTitle>Referral Trend</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-48 flex items-center justify-center text-muted-foreground border-2 border-dashed border-border rounded-lg">
            Chart placeholder — integrate with a charting library
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function StatCard({
  label,
  value,
  loading,
}: {
  label: string;
  value: string | number;
  loading: boolean;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {label}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-3xl font-bold text-foreground">
          {loading ? "--" : value}
        </p>
      </CardContent>
    </Card>
  );
}
