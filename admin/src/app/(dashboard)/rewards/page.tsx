"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";

interface RewardConfig {
  id: string;
  period_type: string;
  rank_from: number;
  rank_to: number;
  reward_type: string;
  amount: number;
  is_active: boolean;
  created_at: string;
}

interface RewardDistribution {
  id: string;
  period_type: string;
  period_key: string;
  users_count: number;
  total_tokens: number;
  distributed_at: string;
}

type Tab = "configs" | "distributions";

export default function RewardsPage() {
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState<Tab>("configs");
  const [showConfigModal, setShowConfigModal] = useState(false);
  const [showTriggerModal, setShowTriggerModal] = useState(false);

  // Config form state
  const [configForm, setConfigForm] = useState({
    period_type: "daily",
    rank_from: 1,
    rank_to: 10,
    reward_type: "token",
    amount: 0,
  });

  // Trigger form state
  const [triggerForm, setTriggerForm] = useState({
    period_type: "daily",
    period_key: new Date().toISOString().slice(0, 10),
  });

  const { data: configs = [], isLoading: configsLoading } = useQuery<RewardConfig[]>({
    queryKey: ["admin-reward-configs"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/rewards/configs");
      return res.data?.data ?? res.data ?? [];
    },
    enabled: activeTab === "configs",
  });

  const { data: distributions = [], isLoading: distributionsLoading } = useQuery<
    RewardDistribution[]
  >({
    queryKey: ["admin-reward-distributions"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/rewards/distributions");
      return res.data?.data ?? res.data ?? [];
    },
    enabled: activeTab === "distributions",
  });

  const createConfigMutation = useMutation({
    mutationFn: async (data: typeof configForm) => {
      return apiClient.post("/admin/rewards/configs", data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-reward-configs"] });
      setShowConfigModal(false);
      setConfigForm({ period_type: "daily", rank_from: 1, rank_to: 10, reward_type: "token", amount: 0 });
    },
  });

  const triggerDistributionMutation = useMutation({
    mutationFn: async (data: typeof triggerForm) => {
      return apiClient.post("/admin/rewards/distribute", data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-reward-distributions"] });
      setShowTriggerModal(false);
    },
  });

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-100">Rewards</h1>
        {activeTab === "configs" && (
          <button
            onClick={() => setShowConfigModal(true)}
            className="bg-blue-600 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-blue-700 transition-colors"
          >
            Create Config
          </button>
        )}
        {activeTab === "distributions" && (
          <button
            onClick={() => setShowTriggerModal(true)}
            className="bg-indigo-600 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-indigo-700 transition-colors"
          >
            Trigger Distribution
          </button>
        )}
      </div>

      {/* Tabs */}
      <div className="flex gap-1 mb-4 bg-gray-800 rounded-lg p-1 w-fit">
        {(["configs", "distributions"] as Tab[]).map((tab) => (
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

      {/* Configs Table */}
      {activeTab === "configs" && (
        <div className="bg-gray-800 shadow rounded-lg border border-gray-700 overflow-hidden">
          <table className="min-w-full divide-y divide-gray-700">
            <thead className="bg-gray-900">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Period Type
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Rank Range
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Reward Type
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Amount
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Status
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-700">
              {configsLoading ? (
                <tr>
                  <td colSpan={5} className="px-6 py-12 text-center text-sm text-gray-400">
                    Loading configs...
                  </td>
                </tr>
              ) : configs.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-6 py-12 text-center text-sm text-gray-400">
                    No reward configs found.
                  </td>
                </tr>
              ) : (
                configs.map((cfg) => (
                  <tr key={cfg.id} className="hover:bg-gray-750 transition-colors">
                    <td className="px-6 py-4 text-sm text-gray-200 capitalize">
                      {cfg.period_type}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-300">
                      #{cfg.rank_from} - #{cfg.rank_to}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-300 capitalize">
                      {cfg.reward_type}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-300 font-mono">
                      {cfg.amount.toLocaleString()}
                    </td>
                    <td className="px-6 py-4 text-sm">
                      {cfg.is_active ? (
                        <span className="inline-flex px-2.5 py-0.5 rounded-full text-xs font-semibold bg-green-500/20 text-green-400 border border-green-500/30">
                          Active
                        </span>
                      ) : (
                        <span className="inline-flex px-2.5 py-0.5 rounded-full text-xs font-semibold bg-gray-500/20 text-gray-400 border border-gray-500/30">
                          Inactive
                        </span>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* Distributions Table */}
      {activeTab === "distributions" && (
        <div className="bg-gray-800 shadow rounded-lg border border-gray-700 overflow-hidden">
          <table className="min-w-full divide-y divide-gray-700">
            <thead className="bg-gray-900">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Period Type
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Period Key
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Users Count
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Total Tokens
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Distributed At
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-700">
              {distributionsLoading ? (
                <tr>
                  <td colSpan={5} className="px-6 py-12 text-center text-sm text-gray-400">
                    Loading distributions...
                  </td>
                </tr>
              ) : distributions.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-6 py-12 text-center text-sm text-gray-400">
                    No distributions found.
                  </td>
                </tr>
              ) : (
                distributions.map((dist) => (
                  <tr key={dist.id} className="hover:bg-gray-750 transition-colors">
                    <td className="px-6 py-4 text-sm text-gray-200 capitalize">
                      {dist.period_type}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-300 font-mono">
                      {dist.period_key}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-300">
                      {dist.users_count.toLocaleString()}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-300 font-mono">
                      {dist.total_tokens.toLocaleString()}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-300">
                      {new Date(dist.distributed_at).toLocaleString()}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* Create Config Modal */}
      {showConfigModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="fixed inset-0 bg-black/60 backdrop-blur-sm" onClick={() => setShowConfigModal(false)} />
          <div className="relative bg-gray-800 border border-gray-700 rounded-xl shadow-2xl w-full max-w-lg mx-4 p-6">
            <button
              onClick={() => setShowConfigModal(false)}
              className="absolute top-4 right-4 text-gray-400 hover:text-gray-200 transition-colors"
            >
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>

            <h2 className="text-lg font-semibold text-gray-100 mb-4">Create Reward Config</h2>

            <div className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">
                  Period Type
                </label>
                <select
                  value={configForm.period_type}
                  onChange={(e) => setConfigForm({ ...configForm, period_type: e.target.value })}
                  className="w-full bg-gray-900 border border-gray-700 text-gray-200 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="daily">Daily</option>
                  <option value="weekly">Weekly</option>
                  <option value="tournament">Tournament</option>
                </select>
              </div>

              <div className="flex gap-4">
                <div className="flex-1">
                  <label className="block text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">
                    Rank From
                  </label>
                  <input
                    type="number"
                    min={1}
                    value={configForm.rank_from}
                    onChange={(e) => setConfigForm({ ...configForm, rank_from: Number(e.target.value) })}
                    className="w-full bg-gray-900 border border-gray-700 text-gray-200 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div className="flex-1">
                  <label className="block text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">
                    Rank To
                  </label>
                  <input
                    type="number"
                    min={1}
                    value={configForm.rank_to}
                    onChange={(e) => setConfigForm({ ...configForm, rank_to: Number(e.target.value) })}
                    className="w-full bg-gray-900 border border-gray-700 text-gray-200 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">
                  Reward Type
                </label>
                <select
                  value={configForm.reward_type}
                  onChange={(e) => setConfigForm({ ...configForm, reward_type: e.target.value })}
                  className="w-full bg-gray-900 border border-gray-700 text-gray-200 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="token">Token</option>
                  <option value="badge">Badge</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">
                  Amount
                </label>
                <input
                  type="number"
                  min={0}
                  value={configForm.amount}
                  onChange={(e) => setConfigForm({ ...configForm, amount: Number(e.target.value) })}
                  className="w-full bg-gray-900 border border-gray-700 text-gray-200 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>

            <div className="flex gap-3 mt-6">
              <button
                onClick={() => setShowConfigModal(false)}
                className="flex-1 bg-gray-700 hover:bg-gray-600 text-gray-300 px-4 py-2.5 rounded-lg text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={() => createConfigMutation.mutate(configForm)}
                disabled={createConfigMutation.isPending}
                className="flex-1 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed text-white px-4 py-2.5 rounded-lg text-sm font-semibold transition-colors"
              >
                {createConfigMutation.isPending ? "Creating..." : "Create"}
              </button>
            </div>
            {createConfigMutation.isError && (
              <p className="mt-3 text-sm text-red-400">Failed to create config. Please try again.</p>
            )}
          </div>
        </div>
      )}

      {/* Trigger Distribution Modal */}
      {showTriggerModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="fixed inset-0 bg-black/60 backdrop-blur-sm" onClick={() => setShowTriggerModal(false)} />
          <div className="relative bg-gray-800 border border-gray-700 rounded-xl shadow-2xl w-full max-w-lg mx-4 p-6">
            <button
              onClick={() => setShowTriggerModal(false)}
              className="absolute top-4 right-4 text-gray-400 hover:text-gray-200 transition-colors"
            >
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>

            <h2 className="text-lg font-semibold text-gray-100 mb-4">Trigger Distribution</h2>

            <div className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">
                  Period Type
                </label>
                <select
                  value={triggerForm.period_type}
                  onChange={(e) => setTriggerForm({ ...triggerForm, period_type: e.target.value })}
                  className="w-full bg-gray-900 border border-gray-700 text-gray-200 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="daily">Daily</option>
                  <option value="weekly">Weekly</option>
                  <option value="tournament">Tournament</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">
                  Period Key
                </label>
                <input
                  type="text"
                  value={triggerForm.period_key}
                  onChange={(e) => setTriggerForm({ ...triggerForm, period_key: e.target.value })}
                  placeholder="e.g. 2026-03-10 or 2026-W11"
                  className="w-full bg-gray-900 border border-gray-700 text-gray-200 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>

            <div className="flex gap-3 mt-6">
              <button
                onClick={() => setShowTriggerModal(false)}
                className="flex-1 bg-gray-700 hover:bg-gray-600 text-gray-300 px-4 py-2.5 rounded-lg text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={() => triggerDistributionMutation.mutate(triggerForm)}
                disabled={triggerDistributionMutation.isPending}
                className="flex-1 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed text-white px-4 py-2.5 rounded-lg text-sm font-semibold transition-colors"
              >
                {triggerDistributionMutation.isPending ? "Distributing..." : "Trigger"}
              </button>
            </div>
            {triggerDistributionMutation.isError && (
              <p className="mt-3 text-sm text-red-400">Failed to trigger distribution. Please try again.</p>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
