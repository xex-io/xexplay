"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { Search, ShieldAlert, Ban } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "@/components/ui/table";

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

function getStatusVariant(status: string): "default" | "secondary" | "destructive" | "outline" {
  switch (status) {
    case "active":
      return "default";
    case "suspended":
      return "secondary";
    case "banned":
      return "destructive";
    default:
      return "outline";
  }
}

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
        <h1 className="text-2xl font-bold text-foreground">User Moderation</h1>
      </div>

      {/* Search */}
      <form onSubmit={handleSearch} className="mb-6">
        <div className="flex gap-3">
          <Input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search by email or user ID..."
            className="flex-1"
          />
          <Button type="submit" size="lg">
            <Search data-icon="inline-start" />
            Search
          </Button>
        </div>
      </form>

      {userLoading && (
        <div className="text-center py-12 text-muted-foreground text-sm">Searching...</div>
      )}

      {isError && searchTerm && (
        <div className="text-center py-12 text-muted-foreground text-sm">
          User not found. Try a different email or ID.
        </div>
      )}

      {user && (
        <>
          {/* User Detail Panel */}
          <Card className="mb-6">
            <CardContent>
              <div className="flex items-start gap-5">
                <div className="w-16 h-16 rounded-full bg-muted flex items-center justify-center text-muted-foreground text-2xl font-bold overflow-hidden flex-shrink-0">
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
                    <h2 className="text-lg font-semibold text-foreground">
                      {user.display_name || "Unnamed User"}
                    </h2>
                    <Badge variant={getStatusVariant(user.status)} className="capitalize">
                      {user.status}
                    </Badge>
                  </div>
                  <p className="text-sm text-muted-foreground">{user.email}</p>
                  <p className="text-xs text-muted-foreground font-mono mt-1">ID: {user.id}</p>

                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-4">
                    <div>
                      <p className="text-xs text-muted-foreground uppercase tracking-wider">Joined</p>
                      <p className="text-sm text-foreground">
                        {new Date(user.created_at).toLocaleDateString()}
                      </p>
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground uppercase tracking-wider">Total Points</p>
                      <p className="text-sm text-foreground font-mono">
                        {(user.total_points ?? 0).toLocaleString()}
                      </p>
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground uppercase tracking-wider">Sessions Played</p>
                      <p className="text-sm text-foreground font-mono">
                        {(user.sessions_played ?? 0).toLocaleString()}
                      </p>
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground uppercase tracking-wider">Status</p>
                      <p className="text-sm text-foreground capitalize">{user.status}</p>
                    </div>
                  </div>
                </div>

                {/* Action Buttons */}
                <div className="flex flex-col gap-2 flex-shrink-0">
                  <Button
                    variant="outline"
                    size="sm"
                    className="border-yellow-500/30 text-yellow-400 hover:bg-yellow-600/20"
                    onClick={() =>
                      setActionModal({ type: "suspend", userId: user.id })
                    }
                  >
                    <ShieldAlert data-icon="inline-start" />
                    Suspend
                  </Button>
                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={() =>
                      setActionModal({ type: "ban", userId: user.id })
                    }
                  >
                    <Ban data-icon="inline-start" />
                    Ban
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Referral Tree */}
          <Card className="mb-6">
            <CardHeader>
              <CardTitle className="text-sm uppercase tracking-wider text-muted-foreground">
                Referral Tree
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <p className="text-xs text-muted-foreground uppercase tracking-wider mb-1">Referred By</p>
                  {user.referred_by ? (
                    <p className="text-sm text-primary font-mono">{user.referred_by}</p>
                  ) : (
                    <p className="text-sm text-muted-foreground">No referrer</p>
                  )}
                </div>
                <div>
                  <p className="text-xs text-muted-foreground uppercase tracking-wider mb-1">
                    Referred Users ({(user.referrals ?? []).length})
                  </p>
                  {(user.referrals ?? []).length > 0 ? (
                    <div className="flex flex-wrap gap-1.5">
                      {user.referrals.map((refId) => (
                        <Badge key={refId} variant="secondary" className="font-mono text-xs">
                          {refId.slice(0, 8)}
                        </Badge>
                      ))}
                    </div>
                  ) : (
                    <p className="text-sm text-muted-foreground">No referrals</p>
                  )}
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Activity Log */}
          <Card>
            <CardHeader className="border-b">
              <CardTitle className="text-sm uppercase tracking-wider text-muted-foreground">
                Recent Activity
              </CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Timestamp</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Description</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {activityLoading ? (
                    <TableRow>
                      <TableCell colSpan={3} className="text-center py-12 text-muted-foreground">
                        Loading activity...
                      </TableCell>
                    </TableRow>
                  ) : activity.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={3} className="text-center py-12 text-muted-foreground">
                        No activity found.
                      </TableCell>
                    </TableRow>
                  ) : (
                    activity.map((entry) => (
                      <TableRow key={entry.id}>
                        <TableCell className="text-muted-foreground">
                          {new Date(entry.created_at).toLocaleString()}
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline" className="capitalize">
                            {entry.type}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-foreground">
                          {entry.description}
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </>
      )}

      {/* Ban/Suspend Modal */}
      <Dialog
        open={!!actionModal}
        onOpenChange={(open) => {
          if (!open) {
            setActionModal(null);
            setActionReason("");
          }
        }}
      >
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle className="capitalize">
              {actionModal?.type} User
            </DialogTitle>
            <DialogDescription>
              This action will {actionModal?.type === "ban" ? "permanently ban" : "temporarily suspend"} the user.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="action-reason">Reason</Label>
              <Textarea
                id="action-reason"
                value={actionReason}
                onChange={(e) => setActionReason(e.target.value)}
                placeholder="Provide a reason for this action..."
                rows={3}
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setActionModal(null);
                setActionReason("");
              }}
            >
              Cancel
            </Button>
            <Button
              variant={actionModal?.type === "ban" ? "destructive" : "default"}
              onClick={() =>
                actionModal && moderationMutation.mutate({
                  userId: actionModal.userId,
                  action: actionModal.type,
                  reason: actionReason,
                })
              }
              disabled={moderationMutation.isPending || !actionReason.trim()}
            >
              {moderationMutation.isPending
                ? "Processing..."
                : `Confirm ${actionModal?.type === "ban" ? "Ban" : "Suspend"}`}
            </Button>
          </DialogFooter>

          {moderationMutation.isError && (
            <Alert variant="destructive" className="mt-2">
              <AlertDescription>
                Failed to {actionModal?.type} user. Please try again.
              </AlertDescription>
            </Alert>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
