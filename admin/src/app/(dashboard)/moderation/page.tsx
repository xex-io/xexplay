"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";

interface UserDetail {
  id: string;
  email: string;
  display_name: string;
  avatar_url: string;
  status: string;
  created_at: string;
  total_points: number;
  sessions_played: number;
  referred_by: string | null;
  referrals: string[];
}

interface ActivityEntry {
  id: string;
  type: string;
  description: string;
  created_at: string;
  metadata?: Record<string, unknown>;
}

const statusColors: Record<string, string> = {
  active: "bg-green-500/20 text-green-400 border border-green-500/30",
  suspended: "bg-yellow-500/20 text-yellow-400 border border-yellow-500/30",
  banned: "bg-red-500/20 text-red-400 border border-red-500/30",
};

export default function ModerationPage() {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState("");
  const [searchTerm, setSearchTerm] = useState("");
  const [actionModal, setActionModal] = useState<{
    type: "ban" | "suspend";
    userId: string;
  } | null>(null);
  const [actionReason, setActionReason] = useState("");

  const { data: user, isLoading: userLoading, isError } = useQuery<UserDetail>({
    queryKey: ["admin-user-detail", searchTerm],
    queryFn: async () => {
      const res = await apiClient.get(`/admin/users/search`, {
        params: { q: searchTerm },
      });
      return res.data?.data ?? res.data;
    },
    enabled: searchTerm.length > 0,
  });

  const { data: activity = [], isLoading: activityLoading } = useQuery<ActivityEntry[]>({
    queryKey: ["admin-user-activity", user?.id],
    queryFn: async () => {
      const res = await apiClient.get(`/admin/users/${user!.id}/activity`);
      return res.data?.data ?? res.data ?? [];
    },
    enabled: !!user?.id,
  });

  const moderationMutation = useMutation({
    mutationFn: async ({
      userId,
      action,
      reason,
    }: {
      userId: string;
      action: "ban" | "suspend";
      reason: string;
    }) => {
      return apiClient.post(`/admin/users/${userId}/moderate`, {
        action,
        reason,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-user-detail", searchTerm] });
      setActionModal(null);
      setActionReason("");
    },
  });

  function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    setSearchTerm(searchQuery.trim());
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-100">User Moderation</h1>
      </div>

      {/* Search */}
      <form onSubmit={handleSearch} className="mb-6">
        <div className="flex gap-3">
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search by email or user ID..."
            className="flex-1 bg-gray-900 border border-gray-700 text-gray-200 rounded-md px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button
            type="submit"
            className="bg-blue-600 text-white px-6 py-2.5 rounded-md text-sm font-medium hover:bg-blue-700 transition-colors"
          >
            Search
          </button>
        </div>
      </form>

      {userLoading && (
        <div className="text-center py-12 text-gray-400 text-sm">Searching...</div>
      )}

      {isError && searchTerm && (
        <div className="text-center py-12 text-gray-400 text-sm">
          User not found. Try a different email or ID.
        </div>
      )}

      {user && (
        <>
          {/* User Detail Panel */}
          <div className="bg-gray-800 border border-gray-700 rounded-lg p-6 mb-6">
            <div className="flex items-start gap-5">
              <div className="w-16 h-16 rounded-full bg-gray-700 flex items-center justify-center text-gray-400 text-2xl font-bold overflow-hidden flex-shrink-0">
                {user.avatar_url ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    src={user.avatar_url}
                    alt={user.display_name}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  (user.display_name?.[0] || user.email?.[0] || "?").toUpperCase()
                )}
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-3 mb-1">
                  <h2 className="text-lg font-semibold text-gray-100">
                    {user.display_name || "Unnamed User"}
                  </h2>
                  <span
                    className={`inline-flex px-2.5 py-0.5 rounded-full text-xs font-semibold capitalize ${
                      statusColors[user.status] || statusColors.active
                    }`}
                  >
                    {user.status}
                  </span>
                </div>
                <p className="text-sm text-gray-400">{user.email}</p>
                <p className="text-xs text-gray-500 font-mono mt-1">ID: {user.id}</p>

                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-4">
                  <div>
                    <p className="text-xs text-gray-500 uppercase tracking-wider">Joined</p>
                    <p className="text-sm text-gray-200">
                      {new Date(user.created_at).toLocaleDateString()}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-500 uppercase tracking-wider">Total Points</p>
                    <p className="text-sm text-gray-200 font-mono">
                      {(user.total_points ?? 0).toLocaleString()}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-500 uppercase tracking-wider">Sessions Played</p>
                    <p className="text-sm text-gray-200 font-mono">
                      {(user.sessions_played ?? 0).toLocaleString()}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-500 uppercase tracking-wider">Status</p>
                    <p className="text-sm text-gray-200 capitalize">{user.status}</p>
                  </div>
                </div>
              </div>

              {/* Action Buttons */}
              <div className="flex flex-col gap-2 flex-shrink-0">
                <button
                  onClick={() =>
                    setActionModal({ type: "suspend", userId: user.id })
                  }
                  className="bg-yellow-600/20 border border-yellow-500/30 text-yellow-400 hover:bg-yellow-600/30 px-4 py-2 rounded-md text-sm font-medium transition-colors"
                >
                  Suspend
                </button>
                <button
                  onClick={() =>
                    setActionModal({ type: "ban", userId: user.id })
                  }
                  className="bg-red-600/20 border border-red-500/30 text-red-400 hover:bg-red-600/30 px-4 py-2 rounded-md text-sm font-medium transition-colors"
                >
                  Ban
                </button>
              </div>
            </div>
          </div>

          {/* Referral Tree */}
          <div className="bg-gray-800 border border-gray-700 rounded-lg p-6 mb-6">
            <h3 className="text-sm font-semibold text-gray-300 uppercase tracking-wider mb-3">
              Referral Tree
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <p className="text-xs text-gray-500 uppercase tracking-wider mb-1">Referred By</p>
                {user.referred_by ? (
                  <p className="text-sm text-blue-400 font-mono">{user.referred_by}</p>
                ) : (
                  <p className="text-sm text-gray-500">No referrer</p>
                )}
              </div>
              <div>
                <p className="text-xs text-gray-500 uppercase tracking-wider mb-1">
                  Referred Users ({(user.referrals ?? []).length})
                </p>
                {(user.referrals ?? []).length > 0 ? (
                  <div className="flex flex-wrap gap-1.5">
                    {user.referrals.map((refId) => (
                      <span
                        key={refId}
                        className="inline-flex px-2 py-0.5 rounded text-xs font-mono bg-gray-700 text-gray-300"
                      >
                        {refId.slice(0, 8)}
                      </span>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-gray-500">No referrals</p>
                )}
              </div>
            </div>
          </div>

          {/* Activity Log */}
          <div className="bg-gray-800 shadow rounded-lg border border-gray-700 overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-700">
              <h3 className="text-sm font-semibold text-gray-300 uppercase tracking-wider">
                Recent Activity
              </h3>
            </div>
            <table className="min-w-full divide-y divide-gray-700">
              <thead className="bg-gray-900">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Timestamp
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Type
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Description
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-700">
                {activityLoading ? (
                  <tr>
                    <td colSpan={3} className="px-6 py-12 text-center text-sm text-gray-400">
                      Loading activity...
                    </td>
                  </tr>
                ) : activity.length === 0 ? (
                  <tr>
                    <td colSpan={3} className="px-6 py-12 text-center text-sm text-gray-400">
                      No activity found.
                    </td>
                  </tr>
                ) : (
                  activity.map((entry) => (
                    <tr key={entry.id} className="hover:bg-gray-750 transition-colors">
                      <td className="px-6 py-4 text-sm text-gray-300">
                        {new Date(entry.created_at).toLocaleString()}
                      </td>
                      <td className="px-6 py-4 text-sm">
                        <span className="inline-flex px-2.5 py-0.5 rounded-full text-xs font-semibold bg-blue-500/20 text-blue-400 border border-blue-500/30 capitalize">
                          {entry.type}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-200">
                        {entry.description}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </>
      )}

      {/* Ban/Suspend Modal */}
      {actionModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div
            className="fixed inset-0 bg-black/60 backdrop-blur-sm"
            onClick={() => {
              setActionModal(null);
              setActionReason("");
            }}
          />
          <div className="relative bg-gray-800 border border-gray-700 rounded-xl shadow-2xl w-full max-w-lg mx-4 p-6">
            <button
              onClick={() => {
                setActionModal(null);
                setActionReason("");
              }}
              className="absolute top-4 right-4 text-gray-400 hover:text-gray-200 transition-colors"
            >
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>

            <h2 className="text-lg font-semibold text-gray-100 mb-4 capitalize">
              {actionModal.type} User
            </h2>

            <div className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">
                  Reason
                </label>
                <textarea
                  value={actionReason}
                  onChange={(e) => setActionReason(e.target.value)}
                  placeholder="Provide a reason for this action..."
                  rows={3}
                  className="w-full bg-gray-900 border border-gray-700 text-gray-200 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
                />
              </div>
            </div>

            <div className="flex gap-3 mt-6">
              <button
                onClick={() => {
                  setActionModal(null);
                  setActionReason("");
                }}
                className="flex-1 bg-gray-700 hover:bg-gray-600 text-gray-300 px-4 py-2.5 rounded-lg text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={() =>
                  moderationMutation.mutate({
                    userId: actionModal.userId,
                    action: actionModal.type,
                    reason: actionReason,
                  })
                }
                disabled={moderationMutation.isPending || !actionReason.trim()}
                className={`flex-1 px-4 py-2.5 rounded-lg text-sm font-semibold transition-colors disabled:opacity-50 disabled:cursor-not-allowed ${
                  actionModal.type === "ban"
                    ? "bg-red-600 hover:bg-red-500 text-white"
                    : "bg-yellow-600 hover:bg-yellow-500 text-white"
                }`}
              >
                {moderationMutation.isPending
                  ? "Processing..."
                  : `Confirm ${actionModal.type === "ban" ? "Ban" : "Suspend"}`}
              </button>
            </div>
            {moderationMutation.isError && (
              <p className="mt-3 text-sm text-red-400">
                Failed to {actionModal.type} user. Please try again.
              </p>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
