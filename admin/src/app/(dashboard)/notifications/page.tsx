"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";

interface NotificationHistory {
  id: string;
  title: string;
  body: string;
  target: string;
  delivery_count: number;
  sent_at: string;
}

export default function NotificationsPage() {
  const queryClient = useQueryClient();
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [target, setTarget] = useState("all");
  const [showConfirm, setShowConfirm] = useState(false);

  const { data: history = [], isLoading } = useQuery<NotificationHistory[]>({
    queryKey: ["admin-notifications"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/notifications");
      return res.data?.data ?? res.data ?? [];
    },
  });

  const sendMutation = useMutation({
    mutationFn: async (data: { title: string; body: string; target: string }) => {
      return apiClient.post("/admin/notifications/send", data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-notifications"] });
      setTitle("");
      setBody("");
      setTarget("all");
      setShowConfirm(false);
    },
  });

  function handleSend() {
    if (!title.trim() || !body.trim()) return;
    setShowConfirm(true);
  }

  function confirmSend() {
    sendMutation.mutate({ title, body, target });
  }

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-100 mb-6">Notifications</h1>

      {/* Compose Form */}
      <div className="bg-gray-800 border border-gray-700 rounded-lg p-6 mb-8">
        <h2 className="text-lg font-semibold text-gray-200 mb-4">Compose Notification</h2>
        <div className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">
              Title
            </label>
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Notification title"
              className="w-full bg-gray-900 border border-gray-700 text-gray-200 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">
              Body
            </label>
            <textarea
              value={body}
              onChange={(e) => setBody(e.target.value)}
              placeholder="Notification body"
              rows={4}
              className="w-full bg-gray-900 border border-gray-700 text-gray-200 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-y"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">
              Target
            </label>
            <select
              value={target}
              onChange={(e) => setTarget(e.target.value)}
              className="w-full bg-gray-900 border border-gray-700 text-gray-200 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="all">All Users</option>
              <option value="active">Active Users (last 7 days)</option>
              <option value="new">New Users (last 24h)</option>
              <option value="dormant">Dormant Users (inactive 30+ days)</option>
            </select>
          </div>

          <div className="flex justify-end">
            <button
              onClick={handleSend}
              disabled={!title.trim() || !body.trim() || sendMutation.isPending}
              className="bg-blue-600 text-white px-6 py-2 rounded-md text-sm font-medium hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Send Notification
            </button>
          </div>

          {sendMutation.isError && (
            <p className="text-sm text-red-400">Failed to send notification. Please try again.</p>
          )}
          {sendMutation.isSuccess && (
            <p className="text-sm text-green-400">Notification sent successfully.</p>
          )}
        </div>
      </div>

      {/* History Table */}
      <h2 className="text-lg font-semibold text-gray-200 mb-4">History</h2>
      <div className="bg-gray-800 shadow rounded-lg border border-gray-700 overflow-hidden">
        <table className="min-w-full divide-y divide-gray-700">
          <thead className="bg-gray-900">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Title
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Target
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Delivery Count
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Sent At
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-700">
            {isLoading ? (
              <tr>
                <td colSpan={4} className="px-6 py-12 text-center text-sm text-gray-400">
                  Loading history...
                </td>
              </tr>
            ) : history.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-6 py-12 text-center text-sm text-gray-400">
                  No notifications sent yet.
                </td>
              </tr>
            ) : (
              history.map((n) => (
                <tr key={n.id} className="hover:bg-gray-750 transition-colors">
                  <td className="px-6 py-4 text-sm text-gray-200">{n.title}</td>
                  <td className="px-6 py-4 text-sm text-gray-300 capitalize">{n.target}</td>
                  <td className="px-6 py-4 text-sm text-gray-300 font-mono">
                    {n.delivery_count.toLocaleString()}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-300">
                    {new Date(n.sent_at).toLocaleString()}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Confirmation Modal */}
      {showConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="fixed inset-0 bg-black/60 backdrop-blur-sm" onClick={() => setShowConfirm(false)} />
          <div className="relative bg-gray-800 border border-gray-700 rounded-xl shadow-2xl w-full max-w-md mx-4 p-6">
            <h2 className="text-lg font-semibold text-gray-100 mb-4">Confirm Send</h2>
            <div className="bg-gray-900 border border-gray-600 rounded-lg p-4 mb-4 space-y-2">
              <p className="text-sm text-gray-300">
                <span className="font-medium text-gray-200">Title:</span> {title}
              </p>
              <p className="text-sm text-gray-300">
                <span className="font-medium text-gray-200">Target:</span>{" "}
                <span className="capitalize">{target}</span>
              </p>
            </div>
            <p className="text-sm text-gray-400 mb-4">
              Are you sure you want to send this notification? This action cannot be undone.
            </p>
            <div className="flex gap-3">
              <button
                onClick={() => setShowConfirm(false)}
                className="flex-1 bg-gray-700 hover:bg-gray-600 text-gray-300 px-4 py-2.5 rounded-lg text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={confirmSend}
                disabled={sendMutation.isPending}
                className="flex-1 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed text-white px-4 py-2.5 rounded-lg text-sm font-semibold transition-colors"
              >
                {sendMutation.isPending ? "Sending..." : "Confirm Send"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
