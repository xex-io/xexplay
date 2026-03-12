"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { asArray } from "@/lib/loc-str";
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "@/components/ui/table";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select";
import { ActionsMenu } from "@/components/actions-menu";
import { Plus, Zap, Pencil, Power } from "lucide-react";

interface RewardConfig {
  id: string;
  period_type: string;
  rank_from: number;
  rank_to: number;
  reward_type: string;
  amount: number;
  description?: Record<string, string>;
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
  const [editConfig, setEditConfig] = useState<RewardConfig | null>(null);

  // Config form state
  const [configForm, setConfigForm] = useState({
    period_type: "daily",
    rank_from: 1,
    rank_to: 10,
    reward_type: "token",
    amount: 0,
  });

  // Edit form state
  const [editForm, setEditForm] = useState({
    period_type: "daily",
    rank_from: 1,
    rank_to: 10,
    reward_type: "token",
    amount: 0,
    is_active: true,
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
      return asArray<RewardConfig>(res);
    },
    enabled: activeTab === "configs",
  });

  const { data: distributions = [], isLoading: distributionsLoading } = useQuery<
    RewardDistribution[]
  >({
    queryKey: ["admin-reward-distributions"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/rewards/history");
      return asArray<RewardDistribution>(res);
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

  const editConfigMutation = useMutation({
    mutationFn: async ({ id, data }: { id: string; data: typeof editForm }) => {
      return apiClient.put(`/admin/rewards/configs/${id}`, {
        period_type: data.period_type,
        rank_from: data.rank_from,
        rank_to: data.rank_to,
        reward_type: data.reward_type,
        amount: data.amount,
        is_active: data.is_active,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-reward-configs"] });
      setEditConfig(null);
    },
  });

  const toggleActiveMutation = useMutation({
    mutationFn: async (cfg: RewardConfig) => {
      return apiClient.put(`/admin/rewards/configs/${cfg.id}`, {
        period_type: cfg.period_type,
        rank_from: cfg.rank_from,
        rank_to: cfg.rank_to,
        reward_type: cfg.reward_type,
        amount: cfg.amount,
        description: cfg.description,
        is_active: !cfg.is_active,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-reward-configs"] });
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

  function openEditDialog(cfg: RewardConfig) {
    setEditConfig(cfg);
    setEditForm({
      period_type: cfg.period_type,
      rank_from: cfg.rank_from,
      rank_to: cfg.rank_to,
      reward_type: cfg.reward_type,
      amount: cfg.amount,
      is_active: cfg.is_active,
    });
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-foreground">Rewards</h1>
        <div className="flex gap-2">
          {activeTab === "configs" && (
            <Button onClick={() => setShowConfigModal(true)}>
              <Plus data-icon="inline-start" />
              Create Config
            </Button>
          )}
          {activeTab === "distributions" && (
            <Button variant="secondary" onClick={() => setShowTriggerModal(true)}>
              <Zap data-icon="inline-start" />
              Trigger Distribution
            </Button>
          )}
        </div>
      </div>

      <Tabs
        value={activeTab}
        onValueChange={(val) => setActiveTab(val as Tab)}
      >
        <TabsList className="mb-4">
          <TabsTrigger value="configs">Configs</TabsTrigger>
          <TabsTrigger value="distributions">Distributions</TabsTrigger>
        </TabsList>

        {/* Configs Table */}
        <TabsContent value="configs">
          <div className="rounded-lg border border-border bg-card">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Period Type</TableHead>
                  <TableHead>Rank Range</TableHead>
                  <TableHead>Reward Type</TableHead>
                  <TableHead>Amount</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="w-12">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {configsLoading ? (
                  <TableRow>
                    <TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
                      Loading configs...
                    </TableCell>
                  </TableRow>
                ) : configs.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
                      No reward configs found.
                    </TableCell>
                  </TableRow>
                ) : (
                  configs.map((cfg) => (
                    <TableRow key={cfg.id}>
                      <TableCell className="capitalize">{cfg.period_type}</TableCell>
                      <TableCell>#{cfg.rank_from} - #{cfg.rank_to}</TableCell>
                      <TableCell className="capitalize">{cfg.reward_type}</TableCell>
                      <TableCell className="font-mono">{cfg.amount.toLocaleString()}</TableCell>
                      <TableCell>
                        {cfg.is_active ? (
                          <Badge variant="secondary" className="bg-green-500/10 text-green-500">
                            Active
                          </Badge>
                        ) : (
                          <Badge variant="outline">Inactive</Badge>
                        )}
                      </TableCell>
                      <TableCell>
                        <ActionsMenu
                          items={[
                            {
                              label: "Edit",
                              icon: Pencil,
                              onClick: () => openEditDialog(cfg),
                            },
                            {
                              label: cfg.is_active ? "Deactivate" : "Activate",
                              icon: Power,
                              onClick: () => toggleActiveMutation.mutate(cfg),
                              disabled: toggleActiveMutation.isPending,
                            },
                          ]}
                        />
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </TabsContent>

        {/* Distributions Table */}
        <TabsContent value="distributions">
          <div className="rounded-lg border border-border bg-card">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Period Type</TableHead>
                  <TableHead>Period Key</TableHead>
                  <TableHead>Users Count</TableHead>
                  <TableHead>Total Tokens</TableHead>
                  <TableHead>Distributed At</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {distributionsLoading ? (
                  <TableRow>
                    <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                      Loading distributions...
                    </TableCell>
                  </TableRow>
                ) : distributions.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                      No distributions found.
                    </TableCell>
                  </TableRow>
                ) : (
                  distributions.map((dist) => (
                    <TableRow key={dist.id}>
                      <TableCell className="capitalize">{dist.period_type}</TableCell>
                      <TableCell className="font-mono">{dist.period_key}</TableCell>
                      <TableCell>{dist.users_count.toLocaleString()}</TableCell>
                      <TableCell className="font-mono">{dist.total_tokens.toLocaleString()}</TableCell>
                      <TableCell>{new Date(dist.distributed_at).toLocaleString()}</TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </TabsContent>
      </Tabs>

      {/* Create Config Dialog */}
      <Dialog open={showConfigModal} onOpenChange={setShowConfigModal}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Create Reward Config</DialogTitle>
            <DialogDescription>
              Define a new reward configuration for a leaderboard period.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label>Period Type</Label>
              <Select
                value={configForm.period_type}
                onValueChange={(val) => setConfigForm({ ...configForm, period_type: val ?? "" })}
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="daily">Daily</SelectItem>
                  <SelectItem value="weekly">Weekly</SelectItem>
                  <SelectItem value="tournament">Tournament</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="flex gap-4">
              <div className="flex-1 space-y-2">
                <Label>Rank From</Label>
                <Input
                  type="number"
                  min={1}
                  value={configForm.rank_from}
                  onChange={(e) => setConfigForm({ ...configForm, rank_from: Number(e.target.value) })}
                />
              </div>
              <div className="flex-1 space-y-2">
                <Label>Rank To</Label>
                <Input
                  type="number"
                  min={1}
                  value={configForm.rank_to}
                  onChange={(e) => setConfigForm({ ...configForm, rank_to: Number(e.target.value) })}
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label>Reward Type</Label>
              <Select
                value={configForm.reward_type}
                onValueChange={(val) => setConfigForm({ ...configForm, reward_type: val ?? "" })}
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="token">Token</SelectItem>
                  <SelectItem value="badge">Badge</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label>Amount</Label>
              <Input
                type="number"
                min={0}
                value={configForm.amount}
                onChange={(e) => setConfigForm({ ...configForm, amount: Number(e.target.value) })}
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowConfigModal(false)}>
              Cancel
            </Button>
            <Button
              onClick={() => createConfigMutation.mutate(configForm)}
              disabled={createConfigMutation.isPending}
            >
              {createConfigMutation.isPending ? "Creating..." : "Create"}
            </Button>
          </DialogFooter>
          {createConfigMutation.isError && (
            <p className="text-sm text-destructive">Failed to create config. Please try again.</p>
          )}
        </DialogContent>
      </Dialog>

      {/* Edit Config Dialog */}
      <Dialog open={editConfig !== null} onOpenChange={(open) => { if (!open) setEditConfig(null); }}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Edit Reward Config</DialogTitle>
            <DialogDescription>
              Update the reward configuration settings.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label>Period Type</Label>
              <Select
                value={editForm.period_type}
                onValueChange={(val) => setEditForm({ ...editForm, period_type: val ?? "" })}
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="daily">Daily</SelectItem>
                  <SelectItem value="weekly">Weekly</SelectItem>
                  <SelectItem value="tournament">Tournament</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="flex gap-4">
              <div className="flex-1 space-y-2">
                <Label>Rank From</Label>
                <Input
                  type="number"
                  min={1}
                  value={editForm.rank_from}
                  onChange={(e) => setEditForm({ ...editForm, rank_from: Number(e.target.value) })}
                />
              </div>
              <div className="flex-1 space-y-2">
                <Label>Rank To</Label>
                <Input
                  type="number"
                  min={1}
                  value={editForm.rank_to}
                  onChange={(e) => setEditForm({ ...editForm, rank_to: Number(e.target.value) })}
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label>Reward Type</Label>
              <Select
                value={editForm.reward_type}
                onValueChange={(val) => setEditForm({ ...editForm, reward_type: val ?? "" })}
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="token">Token</SelectItem>
                  <SelectItem value="badge">Badge</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label>Amount</Label>
              <Input
                type="number"
                min={0}
                value={editForm.amount}
                onChange={(e) => setEditForm({ ...editForm, amount: Number(e.target.value) })}
              />
            </div>

            <div className="flex items-center justify-between">
              <Label>Active</Label>
              <Switch
                checked={editForm.is_active}
                onCheckedChange={(checked) => setEditForm({ ...editForm, is_active: checked })}
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setEditConfig(null)}>
              Cancel
            </Button>
            <Button
              onClick={() => {
                if (editConfig) {
                  editConfigMutation.mutate({ id: editConfig.id, data: editForm });
                }
              }}
              disabled={editConfigMutation.isPending}
            >
              {editConfigMutation.isPending ? "Saving..." : "Save Changes"}
            </Button>
          </DialogFooter>
          {editConfigMutation.isError && (
            <p className="text-sm text-destructive">Failed to update config. Please try again.</p>
          )}
        </DialogContent>
      </Dialog>

      {/* Trigger Distribution Dialog */}
      <Dialog open={showTriggerModal} onOpenChange={setShowTriggerModal}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Trigger Distribution</DialogTitle>
            <DialogDescription>
              Distribute rewards for a specific leaderboard period.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label>Period Type</Label>
              <Select
                value={triggerForm.period_type}
                onValueChange={(val) => setTriggerForm({ ...triggerForm, period_type: val ?? "" })}
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="daily">Daily</SelectItem>
                  <SelectItem value="weekly">Weekly</SelectItem>
                  <SelectItem value="tournament">Tournament</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label>Period Key</Label>
              <Input
                type="text"
                value={triggerForm.period_key}
                onChange={(e) => setTriggerForm({ ...triggerForm, period_key: e.target.value })}
                placeholder="e.g. 2026-03-10 or 2026-W11"
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowTriggerModal(false)}>
              Cancel
            </Button>
            <Button
              variant="secondary"
              onClick={() => triggerDistributionMutation.mutate(triggerForm)}
              disabled={triggerDistributionMutation.isPending}
            >
              {triggerDistributionMutation.isPending ? "Distributing..." : "Trigger"}
            </Button>
          </DialogFooter>
          {triggerDistributionMutation.isError && (
            <p className="text-sm text-destructive">Failed to trigger distribution. Please try again.</p>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
