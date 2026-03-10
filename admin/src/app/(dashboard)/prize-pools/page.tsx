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
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface PrizePool {
  id: string;
  eventName: string;
  totalTokens: number;
  distributed: number;
  remaining: number;
  status: string;
  createdAt: string;
}

interface DistributionEntry {
  id: string;
  poolId: string;
  eventName: string;
  rank: number;
  userId: string;
  tokens: number;
  distributedAt: string;
}

export default function PrizePoolsPage() {
  const [pools, setPools] = useState<PrizePool[]>([]);
  const [history, setHistory] = useState<DistributionEntry[]>([]);
  const [loading, setLoading] = useState(true);

  // Form state
  const [formEvent, setFormEvent] = useState("");
  const [formTokens, setFormTokens] = useState("");
  const [formDistribution, setFormDistribution] = useState("50,30,20");
  const [submitting, setSubmitting] = useState(false);

  async function fetchData() {
    try {
      const [poolsRes, historyRes] = await Promise.all([
        apiClient.get("/admin/prize-pools"),
        apiClient.get("/admin/prize-pools/history"),
      ]);
      setPools(poolsRes.data ?? []);
      setHistory(historyRes.data ?? []);
    } catch {
      // API may not be available yet
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    fetchData();
  }, []);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    try {
      const percentages = formDistribution
        .split(",")
        .map((s) => parseFloat(s.trim()))
        .filter((n) => !isNaN(n));

      await apiClient.post("/admin/prize-pools", {
        eventName: formEvent,
        totalTokens: parseInt(formTokens, 10),
        distributionPercentages: percentages,
      });

      setFormEvent("");
      setFormTokens("");
      setFormDistribution("50,30,20");
      await fetchData();
    } catch {
      // handle error
    } finally {
      setSubmitting(false);
    }
  }

  function poolStatusVariant(
    status: string
  ): "default" | "secondary" | "outline" {
    if (status === "active") return "default";
    if (status === "completed") return "secondary";
    return "outline";
  }

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
            className="grid grid-cols-1 md:grid-cols-4 gap-4 items-end"
          >
            <div className="space-y-1.5">
              <Label htmlFor="form-event">Tournament / Event</Label>
              <Input
                id="form-event"
                type="text"
                value={formEvent}
                onChange={(e) => setFormEvent(e.target.value)}
                required
                placeholder="e.g. Champions League Final"
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="form-tokens">Total Tokens</Label>
              <Input
                id="form-tokens"
                type="number"
                value={formTokens}
                onChange={(e) => setFormTokens(e.target.value)}
                required
                min={1}
                placeholder="10000"
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="form-distribution">
                Distribution % (top N, comma-sep)
              </Label>
              <Input
                id="form-distribution"
                type="text"
                value={formDistribution}
                onChange={(e) => setFormDistribution(e.target.value)}
                placeholder="50,30,20"
              />
            </div>
            <div>
              <Button type="submit" disabled={submitting} className="w-full">
                {submitting ? "Creating..." : "Create Pool"}
              </Button>
            </div>
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
              <TableHead>Event</TableHead>
              <TableHead>Total Tokens</TableHead>
              <TableHead>Distributed</TableHead>
              <TableHead>Remaining</TableHead>
              <TableHead>Status</TableHead>
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
            ) : pools.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={5}
                  className="text-center py-8 text-muted-foreground"
                >
                  No prize pools yet
                </TableCell>
              </TableRow>
            ) : (
              pools.map((p) => (
                <TableRow key={p.id}>
                  <TableCell className="font-medium">{p.eventName}</TableCell>
                  <TableCell>{p.totalTokens.toLocaleString()}</TableCell>
                  <TableCell>{p.distributed.toLocaleString()}</TableCell>
                  <TableCell>{p.remaining.toLocaleString()}</TableCell>
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
              <TableHead>Event</TableHead>
              <TableHead>Rank</TableHead>
              <TableHead>User</TableHead>
              <TableHead>Tokens</TableHead>
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
                  <TableCell>{d.eventName}</TableCell>
                  <TableCell>
                    <Badge variant="outline">#{d.rank}</Badge>
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {d.userId}
                  </TableCell>
                  <TableCell>{d.tokens.toLocaleString()}</TableCell>
                  <TableCell className="text-muted-foreground">
                    {new Date(d.distributedAt).toLocaleDateString()}
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
