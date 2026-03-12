"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { asArray } from "@/lib/loc-str";
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
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogClose,
} from "@/components/ui/dialog";
import { Pencil, XCircle, Loader2 } from "lucide-react";
import { ActionsMenu, type ActionItem } from "@/components/actions-menu";
import { DeleteDialog } from "@/components/delete-dialog";

interface PrizePool {
  id: string;
  name: string;
  description: string;
  total_amount: number;
  currency: string;
  distribution_type: string;
  event_id: string;
  start_date: string;
  end_date: string;
  status: string;
  distributed: number;
  remaining: number;
  created_by: string;
  created_at: string;
  updated_at: string;
}

interface DistributionEntry {
  id: string;
  pool_id: string;
  pool_name: string;
  rank: number;
  user_id: string;
  amount: number;
  distributed_at: string;
}

function poolStatusVariant(
  status: string
): "default" | "secondary" | "outline" {
  if (status === "active") return "default";
  if (status === "completed") return "secondary";
  return "outline";
}

const emptyEditForm = {
  name: "",
  description: "",
  total_amount: "",
  currency: "",
  start_date: "",
  end_date: "",
};

export default function PrizePoolsPage() {
  const queryClient = useQueryClient();

  // Form state
  const [formName, setFormName] = useState("");
  const [formAmount, setFormAmount] = useState("");
  const [formCurrency, setFormCurrency] = useState("XEX");
  const [formDistributionType, setFormDistributionType] = useState("top_n");
  const [formEventId, setFormEventId] = useState("");
  const [formStartDate, setFormStartDate] = useState("");
  const [formEndDate, setFormEndDate] = useState("");

  // Edit & Cancel state
  const [editPool, setEditPool] = useState<PrizePool | null>(null);
  const [editForm, setEditForm] = useState(emptyEditForm);
  const [cancelPool, setCancelPool] = useState<PrizePool | null>(null);

  // Fetch prize pools
  const { data: pools = [], isLoading: poolsLoading } = useQuery<PrizePool[]>({
    queryKey: ["admin-prize-pools"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/prize-pools");
      return asArray<PrizePool>(res);
    },
  });

  // Fetch distribution history
  const { data: history = [], isLoading: historyLoading } = useQuery<DistributionEntry[]>({
    queryKey: ["admin-prize-pools-history"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/prize-pools/history");
      return asArray<DistributionEntry>(res);
    },
  });

  // Create prize pool
  const createMutation = useMutation({
    mutationFn: async (body: {
      name: string;
      total_amount: number;
      currency: string;
      distribution_type: string;
      event_id: string;
      start_date: string;
      end_date: string;
    }) => {
      return apiClient.post("/admin/prize-pools", body);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-prize-pools"] });
      setFormName("");
      setFormAmount("");
      setFormCurrency("XEX");
      setFormDistributionType("top_n");
      setFormEventId("");
      setFormStartDate("");
      setFormEndDate("");
    },
  });

  // Edit prize pool
  const editMutation = useMutation({
    mutationFn: async ({
      id,
      payload,
    }: {
      id: string;
      payload: typeof editForm;
    }) => {
      const res = await apiClient.put(`/admin/prize-pools/${id}`, {
        name: payload.name,
        description: payload.description,
        total_amount: parseFloat(payload.total_amount),
        currency: payload.currency,
        start_date: payload.start_date,
        end_date: payload.end_date,
      });
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-prize-pools"] });
      setEditPool(null);
    },
  });

  // Cancel prize pool
  const cancelMutation = useMutation({
    mutationFn: async (id: string) => {
      await apiClient.delete(`/admin/prize-pools/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-prize-pools"] });
      setCancelPool(null);
    },
  });

  function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    createMutation.mutate({
      name: formName,
      total_amount: parseInt(formAmount, 10),
      currency: formCurrency,
      distribution_type: formDistributionType,
      event_id: formEventId,
      start_date: formStartDate,
      end_date: formEndDate,
    });
  }

  function handleEditSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!editPool) return;
    editMutation.mutate({ id: editPool.id, payload: editForm });
  }

  function openEdit(pool: PrizePool) {
    setEditForm({
      name: pool.name,
      description: pool.description ?? "",
      total_amount: String(pool.total_amount),
      currency: pool.currency,
      start_date: pool.start_date
        ? new Date(pool.start_date).toISOString().slice(0, 16)
        : "",
      end_date: pool.end_date
        ? new Date(pool.end_date).toISOString().slice(0, 16)
        : "",
    });
    setEditPool(pool);
  }

  function actionsFor(pool: PrizePool): ActionItem[] {
    return [
      {
        label: "Edit",
        icon: Pencil,
        onClick: () => openEdit(pool),
      },
      {
        label: "Cancel Pool",
        icon: XCircle,
        variant: "destructive",
        onClick: () => setCancelPool(pool),
        disabled: pool.status === "cancelled",
      },
    ];
  }

  const loading = poolsLoading || historyLoading;

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-foreground">Prize Pools</h1>

      {/* Create prize pool form */}
      <Card>
        <CardHeader>
          <CardTitle>Create Prize Pool</CardTitle>
        </CardHeader>
        <CardContent>
          <form
            onSubmit={handleCreate}
            className="space-y-4"
          >
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="space-y-1.5">
                <Label htmlFor="form-name">Pool Name</Label>
                <Input
                  id="form-name"
                  type="text"
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  required
                  placeholder="e.g. Champions League Final Pool"
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="form-amount">Total Amount</Label>
                <Input
                  id="form-amount"
                  type="number"
                  value={formAmount}
                  onChange={(e) => setFormAmount(e.target.value)}
                  required
                  min={1}
                  placeholder="10000"
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="form-currency">Currency</Label>
                <Input
                  id="form-currency"
                  type="text"
                  value={formCurrency}
                  onChange={(e) => setFormCurrency(e.target.value)}
                  required
                  placeholder="XEX"
                />
              </div>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="space-y-1.5">
                <Label htmlFor="form-distribution-type">Distribution Type</Label>
                <Input
                  id="form-distribution-type"
                  type="text"
                  value={formDistributionType}
                  onChange={(e) => setFormDistributionType(e.target.value)}
                  required
                  placeholder="top_n"
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="form-event-id">Event ID</Label>
                <Input
                  id="form-event-id"
                  type="text"
                  value={formEventId}
                  onChange={(e) => setFormEventId(e.target.value)}
                  required
                  placeholder="Event UUID"
                />
              </div>
              <div className="space-y-1.5 hidden">
                {/* spacer */}
              </div>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="space-y-1.5">
                <Label htmlFor="form-start-date">Start Date</Label>
                <Input
                  id="form-start-date"
                  type="datetime-local"
                  value={formStartDate}
                  onChange={(e) => setFormStartDate(e.target.value)}
                  required
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="form-end-date">End Date</Label>
                <Input
                  id="form-end-date"
                  type="datetime-local"
                  value={formEndDate}
                  onChange={(e) => setFormEndDate(e.target.value)}
                  required
                />
              </div>
              <div className="flex items-end">
                <Button type="submit" disabled={createMutation.isPending} className="w-full">
                  {createMutation.isPending ? "Creating..." : "Create Pool"}
                </Button>
              </div>
            </div>

            {createMutation.isError && (
              <Alert variant="destructive">
                <AlertDescription>
                  Failed to create prize pool. Please try again.
                </AlertDescription>
              </Alert>
            )}

            {createMutation.isSuccess && (
              <Alert>
                <AlertDescription>
                  Prize pool created successfully.
                </AlertDescription>
              </Alert>
            )}
          </form>
        </CardContent>
      </Card>

      {/* Edit Dialog */}
      <Dialog
        open={editPool !== null}
        onOpenChange={(open) => {
          if (!open) setEditPool(null);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Prize Pool</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleEditSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="edit-name">Name</Label>
              <Input
                id="edit-name"
                value={editForm.name}
                onChange={(e) =>
                  setEditForm({ ...editForm, name: e.target.value })
                }
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-description">Description</Label>
              <Input
                id="edit-description"
                value={editForm.description}
                onChange={(e) =>
                  setEditForm({ ...editForm, description: e.target.value })
                }
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="edit-total_amount">Total Amount</Label>
                <Input
                  id="edit-total_amount"
                  type="number"
                  min={1}
                  value={editForm.total_amount}
                  onChange={(e) =>
                    setEditForm({ ...editForm, total_amount: e.target.value })
                  }
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="edit-currency">Currency</Label>
                <Input
                  id="edit-currency"
                  value={editForm.currency}
                  onChange={(e) =>
                    setEditForm({ ...editForm, currency: e.target.value })
                  }
                  required
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="edit-start_date">Start Date</Label>
                <Input
                  id="edit-start_date"
                  type="datetime-local"
                  value={editForm.start_date}
                  onChange={(e) =>
                    setEditForm({ ...editForm, start_date: e.target.value })
                  }
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="edit-end_date">End Date</Label>
                <Input
                  id="edit-end_date"
                  type="datetime-local"
                  value={editForm.end_date}
                  onChange={(e) =>
                    setEditForm({ ...editForm, end_date: e.target.value })
                  }
                  required
                />
              </div>
            </div>
            <DialogFooter>
              <DialogClose render={<Button type="button" variant="outline" />}>
                Cancel
              </DialogClose>
              <Button type="submit" disabled={editMutation.isPending}>
                {editMutation.isPending && (
                  <Loader2 className="size-4 mr-1 animate-spin" />
                )}
                Save Changes
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Cancel Dialog */}
      <DeleteDialog
        open={cancelPool !== null}
        onOpenChange={(open) => {
          if (!open) setCancelPool(null);
        }}
        title="Cancel Prize Pool"
        description={`Are you sure you want to cancel "${cancelPool?.name ?? ""}"? This will set the pool status to cancelled.`}
        onConfirm={() => {
          if (cancelPool) cancelMutation.mutate(cancelPool.id);
        }}
        isPending={cancelMutation.isPending}
        isError={cancelMutation.isError}
        confirmLabel="Cancel Pool"
      />

      {/* Active pools table */}
      <Card>
        <CardHeader className="border-b">
          <CardTitle>Active Pools</CardTitle>
        </CardHeader>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Total Amount</TableHead>
              <TableHead>Currency</TableHead>
              <TableHead>Distribution Type</TableHead>
              <TableHead>Start</TableHead>
              <TableHead>End</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="w-12">
                <span className="sr-only">Actions</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {poolsLoading ? (
              <TableRow>
                <TableCell
                  colSpan={8}
                  className="text-center py-8 text-muted-foreground"
                >
                  Loading...
                </TableCell>
              </TableRow>
            ) : pools.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={8}
                  className="text-center py-8 text-muted-foreground"
                >
                  No prize pools yet
                </TableCell>
              </TableRow>
            ) : (
              pools.map((p) => (
                <TableRow key={p.id}>
                  <TableCell className="font-medium">{p.name}</TableCell>
                  <TableCell>{(p.total_amount ?? 0).toLocaleString()}</TableCell>
                  <TableCell>{p.currency}</TableCell>
                  <TableCell>
                    <Badge variant="outline">{p.distribution_type}</Badge>
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {p.start_date ? new Date(p.start_date).toLocaleDateString() : "--"}
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {p.end_date ? new Date(p.end_date).toLocaleDateString() : "--"}
                  </TableCell>
                  <TableCell>
                    <Badge variant={poolStatusVariant(p.status)}>
                      {p.status}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <ActionsMenu items={actionsFor(p)} />
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </Card>

      {/* Distribution history */}
      <Card>
        <CardHeader className="border-b">
          <CardTitle>Distribution History</CardTitle>
        </CardHeader>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Pool</TableHead>
              <TableHead>Rank</TableHead>
              <TableHead>User</TableHead>
              <TableHead>Amount</TableHead>
              <TableHead>Date</TableHead>
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
            ) : history.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={5}
                  className="text-center py-8 text-muted-foreground"
                >
                  No distributions yet
                </TableCell>
              </TableRow>
            ) : (
              history.map((d) => (
                <TableRow key={d.id}>
                  <TableCell>{d.pool_name}</TableCell>
                  <TableCell>
                    <Badge variant="outline">#{d.rank}</Badge>
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {d.user_id}
                  </TableCell>
                  <TableCell>{(d.amount ?? 0).toLocaleString()}</TableCell>
                  <TableCell className="text-muted-foreground">
                    {d.distributed_at ? new Date(d.distributed_at).toLocaleDateString() : "--"}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </Card>
    </div>
  );
}
