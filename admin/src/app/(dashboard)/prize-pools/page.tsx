"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
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
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";

interface PrizePool {
  id: string;
  name: string;
  total_amount: number;
  currency: string;
  distribution_type: string;
  event_id: string;
  start_date: string;
  end_date: string;
  status: string;
  distributed: number;
  remaining: number;
  created_at: string;
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

  // Fetch prize pools
  const { data: pools = [], isLoading: poolsLoading } = useQuery<PrizePool[]>({
    queryKey: ["admin-prize-pools"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/prize-pools");
      return res.data?.data ?? res.data ?? [];
    },
  });

  // Fetch distribution history
  const { data: history = [], isLoading: historyLoading } = useQuery<DistributionEntry[]>({
    queryKey: ["admin-prize-pools-history"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/prize-pools/history");
      return res.data?.data ?? res.data ?? [];
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
            </TableRow>
          </TableHeader>
          <TableBody>
            {poolsLoading ? (
              <TableRow>
                <TableCell
                  colSpan={7}
                  className="text-center py-8 text-muted-foreground"
                >
                  Loading...
                </TableCell>
              </TableRow>
            ) : pools.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={7}
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
