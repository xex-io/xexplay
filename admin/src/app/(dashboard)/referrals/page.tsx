"use client";

import { useEffect, useState } from "react";
import apiClient from "@/lib/api-client";

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
      <h1 className="text-2xl font-bold">Referral Analytics</h1>

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
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold">Top Referrers</h2>
        </div>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  User
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Referral Count
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Completed
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Rate
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {loading ? (
                <tr>
                  <td colSpan={4} className="px-6 py-8 text-center text-gray-500">
                    Loading...
                  </td>
                </tr>
              ) : topReferrers.length === 0 ? (
                <tr>
                  <td colSpan={4} className="px-6 py-8 text-center text-gray-500">
                    No referral data yet
                  </td>
                </tr>
              ) : (
                topReferrers.map((r) => (
                  <tr key={r.userId} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                      {r.username || r.userId}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                      {r.referralCount}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                      {r.completedReferrals}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                      {r.referralCount > 0
                        ? `${((r.completedReferrals / r.referralCount) * 100).toFixed(1)}%`
                        : "0%"}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Referral trend placeholder */}
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold mb-4">Referral Trend</h2>
        <div className="h-48 flex items-center justify-center text-gray-400 border-2 border-dashed border-gray-200 rounded-lg">
          Chart placeholder — integrate with a charting library
        </div>
      </div>
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
    <div className="bg-white rounded-lg shadow p-6">
      <p className="text-sm font-medium text-gray-500">{label}</p>
      <p className="mt-2 text-3xl font-bold text-gray-900">
        {loading ? "--" : value}
      </p>
    </div>
  );
}
