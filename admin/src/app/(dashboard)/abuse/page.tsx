"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";

interface AbuseFlag {
  id: string;
  user_id: string;
  user_email: string;
  flag_type: string;
  details: string;
  status: "pending" | "approved" | "dismissed";
  reviewed_by: string | null;
  review_reason: string | null;
  created_at: string;
  reviewed_at: string | null;
}

type Tab = "pending" | "reviewed";

const flagTypeColors: Record<string, string> = {
  multi_account: "bg-red-500/20 text-red-400 border border-red-500/30",
  bot_activity: "bg-purple-500/20 text-purple-400 border border-purple-500/30",
  rapid_answers: "bg-orange-500/20 text-orange-400 border border-orange-500/30",
  suspicious_referral: "bg-yellow-500/20 text-yellow-400 border border-yellow-500/30",
  ip_anomaly: "bg-blue-500/20 text-blue-400 border border-blue-500/30",
};

function getFlagColor(type: string): string {
  return flagTypeColors[type] || "bg-gray-500/20 text-gray-400 border border-gray-500/30";
}

export default function AbusePage() {
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState<Tab>("pending");
  const [reviewModal, setReviewModal] = useState<{
    flag: AbuseFlag;
    action: "approve" | "dismiss";
  } | null>(null);
  const [reviewReason, setReviewReason] = useState("");

  const { data: pendingFlags = [], isLoading: pendingLoading } = useQuery<AbuseFlag[]>({
    queryKey: ["admin-abuse-flags", "pending"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/abuse-flags", {
        params: { status: "pending" },
      });
      return res.data?.data ?? res.data ?? [];
    },
    enabled: activeTab === "pending",
  });

  const { data: reviewedFlags = [], isLoading: reviewedLoading } = useQuery<AbuseFlag[]>({
    queryKey: ["admin-abuse-flags", "reviewed"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/abuse-flags", {
        params: { status: "reviewed" },
      });
      return res.data?.data ?? res.data ?? [];
    },
    enabled: activeTab === "reviewed",
  });

  const { data: stats } = useQuery<{
    total: number;
    pending: number;
    reviewed_today: number;
  }>({
    queryKey: ["admin-abuse-stats"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/abuse-flags/stats");
      return res.data?.data ?? res.data ?? { total: 0, pending: 0, reviewed_today: 0 };
    },
  });

  const reviewMutation = useMutation({
    mutationFn: async ({
      flagId,
      action,
      reason,
    }: {
      flagId: string;
      action: "approve" | "dismiss";
      reason: string;
    }) => {
      return apiClient.post(`/admin/abuse-flags/${flagId}/review`, {
        action,
        reason,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-abuse-flags"] });
      queryClient.invalidateQueries({ queryKey: ["admin-abuse-stats"] });
      setReviewModal(null);
      setReviewReason("");
    },
  });

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-100">Abuse Flags</h1>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        <div className="bg-gray-800 border border-gray-700 rounded-lg p-4">
          <p className="text-xs font-medium text-gray-400 uppercase tracking-wider">Total Flags</p>
          <p className="mt-1 text-2xl font-bold text-gray-100">{stats?.total ?? 0}</p>
        </div>
        <div className="bg-gray-800 border border-gray-700 rounded-lg p-4">
          <p className="text-xs font-medium text-gray-400 uppercase tracking-wider">Pending</p>
          <p className="mt-1 text-2xl font-bold text-yellow-400">{stats?.pending ?? 0}</p>
        </div>
        <div className="bg-gray-800 border border-gray-700 rounded-lg p-4">
          <p className="text-xs font-medium text-gray-400 uppercase tracking-wider">Reviewed Today</p>
          <p className="mt-1 text-2xl font-bold text-green-400">{stats?.reviewed_today ?? 0}</p>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 mb-4 bg-gray-800 rounded-lg p-1 w-fit">
        {(["pending", "reviewed"] as Tab[]).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-4 py-2 rounded-md text-sm font-medium capitalize transition-colors ${
              activeTab === tab
                ? "bg-blue-600 text-white"
                : "text-gray-400 hover:text-gray-200 hover:bg-gray-700"
            }`}
          >
            {tab}
          </button>
        ))}
      </div>

      {/* Pending Flags Table */}
      {activeTab === "pending" && (
        <div className="bg-gray-800 shadow rounded-lg border border-gray-700 overflow-hidden">
          <table className="min-w-full divide-y divide-gray-700">
            <thead className="bg-gray-900">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  User
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Flag Type
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Details
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Created
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-700">
              {pendingLoading ? (
                <tr>
                  <td colSpan={5} className="px-6 py-12 text-center text-sm text-gray-400">
                    Loading flags...
                  </td>
                </tr>
              ) : pendingFlags.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-6 py-12 text-center text-sm text-gray-400">
                    No pending flags.
                  </td>
                </tr>
              ) : (
                pendingFlags.map((flag) => (
                  <tr key={flag.id} className="hover:bg-gray-750 transition-colors">
                    <td className="px-6 py-4 text-sm">
                      <p className="text-gray-200">{flag.user_email}</p>
                      <p className="text-xs text-gray-500 font-mono">{flag.user_id.slice(0, 8)}</p>
                    </td>
                    <td className="px-6 py-4 text-sm">
                      <span
                        className={`inline-flex px-2.5 py-0.5 rounded-full text-xs font-semibold ${getFlagColor(flag.flag_type)}`}
                      >
                        {flag.flag_type.replace(/_/g, " ")}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-300 max-w-xs truncate">
                      {flag.details}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-300">
                      {new Date(flag.created_at).toLocaleString()}
                    </td>
                    <td className="px-6 py-4 text-sm">
                      <div className="flex gap-2">
                        <button
                          onClick={() =>
                            setReviewModal({ flag, action: "approve" })
                          }
                          className="bg-red-600/20 border border-red-500/30 text-red-400 hover:bg-red-600/30 px-3 py-1.5 rounded-md text-xs font-medium transition-colors"
                        >
                          Confirm Abuse
                        </button>
                        <button
                          onClick={() =>
                            setReviewModal({ flag, action: "dismiss" })
                          }
                          className="bg-gray-600/20 border border-gray-500/30 text-gray-300 hover:bg-gray-600/30 px-3 py-1.5 rounded-md text-xs font-medium transition-colors"
                        >
                          Dismiss
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* Reviewed Flags Table */}
      {activeTab === "reviewed" && (
        <div className="bg-gray-800 shadow rounded-lg border border-gray-700 overflow-hidden">
          <table className="min-w-full divide-y divide-gray-700">
            <thead className="bg-gray-900">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  User
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Flag Type
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Outcome
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Reviewed By
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Reason
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Reviewed At
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-700">
              {reviewedLoading ? (
                <tr>
                  <td colSpan={6} className="px-6 py-12 text-center text-sm text-gray-400">
                    Loading reviewed flags...
                  </td>
                </tr>
              ) : reviewedFlags.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-6 py-12 text-center text-sm text-gray-400">
                    No reviewed flags found.
                  </td>
                </tr>
              ) : (
                reviewedFlags.map((flag) => (
                  <tr key={flag.id} className="hover:bg-gray-750 transition-colors">
                    <td className="px-6 py-4 text-sm">
                      <p className="text-gray-200">{flag.user_email}</p>
                      <p className="text-xs text-gray-500 font-mono">{flag.user_id.slice(0, 8)}</p>
                    </td>
                    <td className="px-6 py-4 text-sm">
                      <span
                        className={`inline-flex px-2.5 py-0.5 rounded-full text-xs font-semibold ${getFlagColor(flag.flag_type)}`}
                      >
                        {flag.flag_type.replace(/_/g, " ")}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm">
                      <span
                        className={`inline-flex px-2.5 py-0.5 rounded-full text-xs font-semibold ${
                          flag.status === "approved"
                            ? "bg-red-500/20 text-red-400 border border-red-500/30"
                            : "bg-gray-500/20 text-gray-400 border border-gray-500/30"
                        }`}
                      >
                        {flag.status === "approved" ? "Confirmed" : "Dismissed"}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-300">
                      {flag.reviewed_by || "-"}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-300 max-w-xs truncate">
                      {flag.review_reason || "-"}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-300">
                      {flag.reviewed_at
                        ? new Date(flag.reviewed_at).toLocaleString()
                        : "-"}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* Review Modal */}
      {reviewModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div
            className="fixed inset-0 bg-black/60 backdrop-blur-sm"
            onClick={() => {
              setReviewModal(null);
              setReviewReason("");
            }}
          />
          <div className="relative bg-gray-800 border border-gray-700 rounded-xl shadow-2xl w-full max-w-lg mx-4 p-6">
            <button
              onClick={() => {
                setReviewModal(null);
                setReviewReason("");
              }}
              className="absolute top-4 right-4 text-gray-400 hover:text-gray-200 transition-colors"
            >
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>

            <h2 className="text-lg font-semibold text-gray-100 mb-4">
              {reviewModal.action === "approve" ? "Confirm Abuse" : "Dismiss Flag"}
            </h2>

            <div className="space-y-3 mb-4">
              <div>
                <label className="text-xs font-medium text-gray-400 uppercase tracking-wider">
                  User
                </label>
                <p className="mt-1 text-sm text-gray-200">{reviewModal.flag.user_email}</p>
              </div>
              <div>
                <label className="text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Flag Type
                </label>
                <p className="mt-1">
                  <span
                    className={`inline-flex px-2.5 py-0.5 rounded-full text-xs font-semibold ${getFlagColor(reviewModal.flag.flag_type)}`}
                  >
                    {reviewModal.flag.flag_type.replace(/_/g, " ")}
                  </span>
                </p>
              </div>
              <div>
                <label className="text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Details
                </label>
                <p className="mt-1 text-sm text-gray-300">{reviewModal.flag.details}</p>
              </div>
            </div>

            <div>
              <label className="block text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">
                Reason
              </label>
              <textarea
                value={reviewReason}
                onChange={(e) => setReviewReason(e.target.value)}
                placeholder="Provide a reason for your decision..."
                rows={3}
                className="w-full bg-gray-900 border border-gray-700 text-gray-200 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
              />
            </div>

            <div className="flex gap-3 mt-6">
              <button
                onClick={() => {
                  setReviewModal(null);
                  setReviewReason("");
                }}
                className="flex-1 bg-gray-700 hover:bg-gray-600 text-gray-300 px-4 py-2.5 rounded-lg text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={() =>
                  reviewMutation.mutate({
                    flagId: reviewModal.flag.id,
                    action: reviewModal.action,
                    reason: reviewReason,
                  })
                }
                disabled={reviewMutation.isPending || !reviewReason.trim()}
                className={`flex-1 px-4 py-2.5 rounded-lg text-sm font-semibold transition-colors disabled:opacity-50 disabled:cursor-not-allowed ${
                  reviewModal.action === "approve"
                    ? "bg-red-600 hover:bg-red-500 text-white"
                    : "bg-gray-600 hover:bg-gray-500 text-white"
                }`}
              >
                {reviewMutation.isPending
                  ? "Processing..."
                  : reviewModal.action === "approve"
                    ? "Confirm Abuse"
                    : "Dismiss"}
              </button>
            </div>
            {reviewMutation.isError && (
              <p className="mt-3 text-sm text-red-400">
                Failed to review flag. Please try again.
              </p>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
