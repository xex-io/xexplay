"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";

interface LeaderboardEntry {
  rank: number;
  user_id: string;
  username: string;
  points: number;
  correct_answers: number;
  wrong_answers: number;
}

const tabs = ["daily", "weekly", "tournament", "all-time"] as const;
type TabType = (typeof tabs)[number];

const tabLabels: Record<TabType, string> = {
  daily: "Daily",
  weekly: "Weekly",
  tournament: "Tournament",
  "all-time": "All-Time",
};

function todayKey(): string {
  return new Date().toISOString().slice(0, 10);
}

function exportCSV(entries: LeaderboardEntry[], type: string, periodKey: string) {
  const header = "Rank,User,Points,Correct Answers,Wrong Answers";
  const rows = entries.map(
    (e) => `${e.rank},${e.username},${e.points},${e.correct_answers},${e.wrong_answers}`
  );
  const csv = [header, ...rows].join("\n");
  const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `leaderboard-${type}-${periodKey || "all"}.csv`;
  a.click();
  URL.revokeObjectURL(url);
}

export default function LeaderboardsPage() {
  const [activeTab, setActiveTab] = useState<TabType>("daily");
  const [periodKey, setPeriodKey] = useState(todayKey());

  const needsPeriod = activeTab === "daily" || activeTab === "weekly";

  const { data: entries = [], isLoading } = useQuery<LeaderboardEntry[]>({
    queryKey: ["admin-leaderboards", activeTab, needsPeriod ? periodKey : null],
    queryFn: async () => {
      const params = needsPeriod ? `?period_key=${periodKey}` : "";
      const res = await apiClient.get(`/admin/leaderboards/${activeTab}${params}`);
      return res.data?.data ?? res.data ?? [];
    },
  });

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-100">Leaderboards</h1>
        <button
          onClick={() => exportCSV(entries, activeTab, periodKey)}
          disabled={entries.length === 0}
          className="bg-green-600 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-green-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Export CSV
        </button>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 mb-4 bg-gray-800 rounded-lg p-1 w-fit">
        {tabs.map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              activeTab === tab
                ? "bg-blue-600 text-white"
                : "text-gray-400 hover:text-gray-200 hover:bg-gray-700"
            }`}
          >
            {tabLabels[tab]}
          </button>
        ))}
      </div>

      {/* Date picker for daily/weekly */}
      {needsPeriod && (
        <div className="mb-4">
          <label className="text-xs font-medium text-gray-400 uppercase tracking-wider mr-3">
            Period
          </label>
          <input
            type="date"
            value={periodKey}
            onChange={(e) => setPeriodKey(e.target.value)}
            className="bg-gray-800 border border-gray-700 text-gray-200 rounded-md px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
      )}

      {/* Table */}
      <div className="bg-gray-800 shadow rounded-lg border border-gray-700 overflow-hidden">
        <table className="min-w-full divide-y divide-gray-700">
          <thead className="bg-gray-900">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Rank
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                User
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Points
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Correct Answers
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Wrong Answers
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-700">
            {isLoading ? (
              <tr>
                <td colSpan={5} className="px-6 py-12 text-center text-sm text-gray-400">
                  Loading leaderboard...
                </td>
              </tr>
            ) : entries.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-6 py-12 text-center text-sm text-gray-400">
                  No leaderboard data found.
                </td>
              </tr>
            ) : (
              entries.map((entry) => (
                <tr key={entry.user_id} className="hover:bg-gray-750 transition-colors">
                  <td className="px-6 py-4 text-sm text-gray-200 font-semibold">
                    #{entry.rank}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-200">
                    {entry.username || entry.user_id.slice(0, 8)}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-300 font-mono">
                    {entry.points.toLocaleString()}
                  </td>
                  <td className="px-6 py-4 text-sm text-green-400">
                    {entry.correct_answers}
                  </td>
                  <td className="px-6 py-4 text-sm text-red-400">
                    {entry.wrong_answers}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
