"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { ShieldCheck, ShieldX } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { Separator } from "@/components/ui/separator";

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

const flagTypeVariantMap: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  multi_account: "destructive",
  bot_activity: "secondary",
  rapid_answers: "outline",
  suspicious_referral: "outline",
  ip_anomaly: "default",
};

function getFlagVariant(type: string): "default" | "secondary" | "destructive" | "outline" {
  return flagTypeVariantMap[type] || "secondary";
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
        <h1 className="text-2xl font-bold text-foreground">Abuse Flags</h1>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        <Card size="sm">
          <CardContent>
            <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Total Flags</p>
            <p className="mt-1 text-2xl font-bold text-foreground">{stats?.total ?? 0}</p>
          </CardContent>
        </Card>
        <Card size="sm">
          <CardContent>
            <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Pending</p>
            <p className="mt-1 text-2xl font-bold text-yellow-400">{stats?.pending ?? 0}</p>
          </CardContent>
        </Card>
        <Card size="sm">
          <CardContent>
            <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Reviewed Today</p>
            <p className="mt-1 text-2xl font-bold text-green-400">{stats?.reviewed_today ?? 0}</p>
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <Tabs
        defaultValue="pending"
        value={activeTab}
        onValueChange={(val) => setActiveTab(val as Tab)}
      >
        <TabsList>
          <TabsTrigger value="pending">Pending</TabsTrigger>
          <TabsTrigger value="reviewed">Reviewed</TabsTrigger>
        </TabsList>

        {/* Pending Flags Table */}
        <TabsContent value="pending">
          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>User</TableHead>
                    <TableHead>Flag Type</TableHead>
                    <TableHead>Details</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {pendingLoading ? (
                    <TableRow>
                      <TableCell colSpan={5} className="text-center py-12 text-muted-foreground">
                        Loading flags...
                      </TableCell>
                    </TableRow>
                  ) : pendingFlags.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={5} className="text-center py-12 text-muted-foreground">
                        No pending flags.
                      </TableCell>
                    </TableRow>
                  ) : (
                    pendingFlags.map((flag) => (
                      <TableRow key={flag.id}>
                        <TableCell>
                          <p className="text-foreground">{flag.user_email}</p>
                          <p className="text-xs text-muted-foreground font-mono">{flag.user_id.slice(0, 8)}</p>
                        </TableCell>
                        <TableCell>
                          <Badge variant={getFlagVariant(flag.flag_type)}>
                            {flag.flag_type.replace(/_/g, " ")}
                          </Badge>
                        </TableCell>
                        <TableCell className="max-w-xs truncate text-muted-foreground">
                          {flag.details}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {new Date(flag.created_at).toLocaleString()}
                        </TableCell>
                        <TableCell>
                          <div className="flex gap-2">
                            <Button
                              variant="destructive"
                              size="xs"
                              onClick={() =>
                                setReviewModal({ flag, action: "approve" })
                              }
                            >
                              <ShieldX data-icon="inline-start" />
                              Confirm Abuse
                            </Button>
                            <Button
                              variant="outline"
                              size="xs"
                              onClick={() =>
                                setReviewModal({ flag, action: "dismiss" })
                              }
                            >
                              <ShieldCheck data-icon="inline-start" />
                              Dismiss
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Reviewed Flags Table */}
        <TabsContent value="reviewed">
          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>User</TableHead>
                    <TableHead>Flag Type</TableHead>
                    <TableHead>Outcome</TableHead>
                    <TableHead>Reviewed By</TableHead>
                    <TableHead>Reason</TableHead>
                    <TableHead>Reviewed At</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {reviewedLoading ? (
                    <TableRow>
                      <TableCell colSpan={6} className="text-center py-12 text-muted-foreground">
                        Loading reviewed flags...
                      </TableCell>
                    </TableRow>
                  ) : reviewedFlags.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={6} className="text-center py-12 text-muted-foreground">
                        No reviewed flags found.
                      </TableCell>
                    </TableRow>
                  ) : (
                    reviewedFlags.map((flag) => (
                      <TableRow key={flag.id}>
                        <TableCell>
                          <p className="text-foreground">{flag.user_email}</p>
                          <p className="text-xs text-muted-foreground font-mono">{flag.user_id.slice(0, 8)}</p>
                        </TableCell>
                        <TableCell>
                          <Badge variant={getFlagVariant(flag.flag_type)}>
                            {flag.flag_type.replace(/_/g, " ")}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant={flag.status === "approved" ? "destructive" : "secondary"}>
                            {flag.status === "approved" ? "Confirmed" : "Dismissed"}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {flag.reviewed_by || "-"}
                        </TableCell>
                        <TableCell className="max-w-xs truncate text-muted-foreground">
                          {flag.review_reason || "-"}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {flag.reviewed_at
                            ? new Date(flag.reviewed_at).toLocaleString()
                            : "-"}
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Review Modal */}
      <Dialog
        open={!!reviewModal}
        onOpenChange={(open) => {
          if (!open) {
            setReviewModal(null);
            setReviewReason("");
          }
        }}
      >
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>
              {reviewModal?.action === "approve" ? "Confirm Abuse" : "Dismiss Flag"}
            </DialogTitle>
            <DialogDescription>
              Review the abuse flag details and provide a reason for your decision.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-3">
            <div>
              <Label className="text-xs uppercase tracking-wider text-muted-foreground">
                User
              </Label>
              <p className="mt-1 text-sm text-foreground">{reviewModal?.flag.user_email}</p>
            </div>
            <Separator />
            <div>
              <Label className="text-xs uppercase tracking-wider text-muted-foreground">
                Flag Type
              </Label>
              <p className="mt-1">
                {reviewModal?.flag && (
                  <Badge variant={getFlagVariant(reviewModal.flag.flag_type)}>
                    {reviewModal.flag.flag_type.replace(/_/g, " ")}
                  </Badge>
                )}
              </p>
            </div>
            <Separator />
            <div>
              <Label className="text-xs uppercase tracking-wider text-muted-foreground">
                Details
              </Label>
              <p className="mt-1 text-sm text-muted-foreground">{reviewModal?.flag.details}</p>
            </div>
            <Separator />
            <div className="space-y-2">
              <Label htmlFor="review-reason">Reason</Label>
              <Textarea
                id="review-reason"
                value={reviewReason}
                onChange={(e) => setReviewReason(e.target.value)}
                placeholder="Provide a reason for your decision..."
                rows={3}
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setReviewModal(null);
                setReviewReason("");
              }}
            >
              Cancel
            </Button>
            <Button
              variant={reviewModal?.action === "approve" ? "destructive" : "secondary"}
              onClick={() =>
                reviewModal && reviewMutation.mutate({
                  flagId: reviewModal.flag.id,
                  action: reviewModal.action,
                  reason: reviewReason,
                })
              }
              disabled={reviewMutation.isPending || !reviewReason.trim()}
            >
              {reviewMutation.isPending
                ? "Processing..."
                : reviewModal?.action === "approve"
                  ? "Confirm Abuse"
                  : "Dismiss"}
            </Button>
          </DialogFooter>

          {reviewMutation.isError && (
            <Alert variant="destructive" className="mt-2">
              <AlertDescription>
                Failed to review flag. Please try again.
              </AlertDescription>
            </Alert>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
