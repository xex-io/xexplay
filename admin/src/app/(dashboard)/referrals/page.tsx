"use client";

import { useEffect, useState } from "react";
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
  totalReferrals: number;
  conversionRate: number;
  activeReferrers: number;
}

interface TopReferrer {
  userId: string;
  username: string;
  referralCount: number;
  completedReferrals: number;
}

export default function ReferralsPage() {
  const [stats, setStats] = useState<ReferralStats>({
    totalReferrals: 0,
    conversionRate: 0,
    activeReferrers: 0,
  });
  const [topReferrers, setTopReferrers] = useState<TopReferrer[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchData() {
      try {
        const [statsRes, referrersRes] = await Promise.all([
          apiClient.get("/admin/referrals/stats"),
          apiClient.get("/admin/referrals/top"),
        ]);
        setStats(statsRes.data);
        setTopReferrers(referrersRes.data);
      } catch {
        // API may not be available yet — show empty state
      } finally {
        setLoading(false);
      }
    }
    fetchData();
  }, []);

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-foreground">
        Referral Analytics
      </h1>

      {/* Stats cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <StatCard
          label="Total Referrals"
          value={stats.totalReferrals}
          loading={loading}
        />
        <StatCard
          label="Conversion Rate"
          value={`${stats.conversionRate.toFixed(1)}%`}
          loading={loading}
        />
        <StatCard
          label="Active Referrers"
          value={stats.activeReferrers}
          loading={loading}
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
              <TableHead>Referral Count</TableHead>
              <TableHead>Completed</TableHead>
              <TableHead>Rate</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell
                  colSpan={4}
                  className="text-center py-8 text-muted-foreground"
                >
                  Loading...
                </TableCell>
              </TableRow>
            ) : topReferrers.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={4}
                  className="text-center py-8 text-muted-foreground"
                >
                  No referral data yet
                </TableCell>
              </TableRow>
            ) : (
              topReferrers.map((r) => {
                const rate =
                  r.referralCount > 0
                    ? (r.completedReferrals / r.referralCount) * 100
                    : 0;
                return (
                  <TableRow key={r.userId}>
                    <TableCell className="font-medium">
                      {r.username || r.userId}
                    </TableCell>
                    <TableCell>{r.referralCount}</TableCell>
                    <TableCell>{r.completedReferrals}</TableCell>
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
