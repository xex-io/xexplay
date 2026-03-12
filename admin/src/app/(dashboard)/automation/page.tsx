"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "@/components/ui/table";
import {
  Play,
  RefreshCw,
  Clock,
  CheckCircle2,
  XCircle,
  Loader2,
} from "lucide-react";

interface Sport {
  key: string;
  group: string;
  title: string;
  is_active: boolean;
}

interface AutomationLog {
  id: string;
  job_name: string;
  status: string;
  details: Record<string, unknown>;
  items_processed: number;
  error_message: string;
  created_at: string;
}

interface JobStatus {
  last_run: string | null;
  status: string;
  items_processed?: number;
  error?: string;
}

const JOB_LABELS: Record<string, { label: string; description: string }> = {
  fetchUpcomingMatches: {
    label: "Fetch Matches",
    description: "Fetches upcoming matches from Odds API (every 6h)",
  },
  generateDailyCards: {
    label: "Generate Cards",
    description: "AI generates prediction questions for tomorrow (daily 02:00 UTC)",
  },
  autoPublishBaskets: {
    label: "Auto Publish",
    description: "Publishes today's baskets (daily 06:00 UTC)",
  },
  autoResolveCards: {
    label: "Auto Resolve",
    description: "Resolves cards from completed matches (every 15min)",
  },
  syncSports: {
    label: "Sync Sports",
    description: "Refreshes sports list from Odds API (weekly)",
  },
};

const TRIGGER_MAP: Record<string, string> = {
  fetchUpcomingMatches: "fetchMatches",
  generateDailyCards: "generateCards",
  autoPublishBaskets: "autoPublish",
  autoResolveCards: "autoResolve",
  syncSports: "syncSports",
};

export default function AutomationPage() {
  const queryClient = useQueryClient();
  const [triggeringJob, setTriggeringJob] = useState<string | null>(null);

  const { data: sports, isLoading: sportsLoading } = useQuery<Sport[]>({
    queryKey: ["admin-sports"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/sports");
      return res.data?.data || res.data || [];
    },
  });

  const { data: status } = useQuery<Record<string, JobStatus>>({
    queryKey: ["admin-automation-status"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/automation/status");
      return res.data?.data || res.data || {};
    },
    refetchInterval: 30000,
  });

  const { data: logs } = useQuery<AutomationLog[]>({
    queryKey: ["admin-automation-logs"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/automation/logs");
      return res.data?.data || res.data || [];
    },
    refetchInterval: 30000,
  });

  const toggleSportMutation = useMutation({
    mutationFn: async ({ key, isActive }: { key: string; isActive: boolean }) => {
      return apiClient.put(`/admin/sports/${key}`, { is_active: isActive });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-sports"] });
    },
  });

  const triggerMutation = useMutation({
    mutationFn: async (job: string) => {
      setTriggeringJob(job);
      return apiClient.post("/admin/automation/trigger", { job });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-automation-status"] });
      queryClient.invalidateQueries({ queryKey: ["admin-automation-logs"] });
    },
    onSettled: () => {
      setTriggeringJob(null);
    },
  });

  const sportsByGroup = (sports || []).reduce(
    (acc, sport) => {
      if (!acc[sport.group]) acc[sport.group] = [];
      acc[sport.group].push(sport);
      return acc;
    },
    {} as Record<string, Sport[]>
  );

  return (
    <div>
      <h1 className="text-2xl font-bold text-foreground mb-6">
        Sports Automation
      </h1>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Automation Status */}
        <Card>
          <CardHeader>
            <CardTitle className="text-sm uppercase tracking-wider text-muted-foreground">
              Job Status
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {Object.entries(JOB_LABELS).map(([key, { label, description }]) => {
              const jobStatus = status?.[key];
              const triggerKey = TRIGGER_MAP[key];
              const isTriggering = triggeringJob === triggerKey;

              return (
                <div
                  key={key}
                  className="flex items-center justify-between p-3 rounded-lg border"
                >
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <p className="text-sm font-medium">{label}</p>
                      {jobStatus?.status === "success" && (
                        <CheckCircle2 className="size-4 text-green-500" />
                      )}
                      {jobStatus?.status === "error" && (
                        <XCircle className="size-4 text-red-500" />
                      )}
                      {jobStatus?.status === "never" && (
                        <Clock className="size-4 text-muted-foreground" />
                      )}
                    </div>
                    <p className="text-xs text-muted-foreground">{description}</p>
                    {jobStatus?.last_run && (
                      <p className="text-xs text-muted-foreground mt-1">
                        Last: {new Date(jobStatus.last_run).toLocaleString()}
                        {jobStatus.items_processed
                          ? ` (${jobStatus.items_processed} items)`
                          : ""}
                      </p>
                    )}
                  </div>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => triggerMutation.mutate(triggerKey)}
                    disabled={isTriggering}
                  >
                    {isTriggering ? (
                      <Loader2 className="size-4 animate-spin" />
                    ) : (
                      <Play className="size-4" />
                    )}
                  </Button>
                </div>
              );
            })}
          </CardContent>
        </Card>

        {/* Sports Manager */}
        <Card>
          <CardHeader>
            <CardTitle className="text-sm uppercase tracking-wider text-muted-foreground">
              Sports ({(sports || []).filter((s) => s.is_active).length} active)
            </CardTitle>
          </CardHeader>
          <CardContent>
            {sportsLoading ? (
              <p className="text-sm text-muted-foreground">Loading...</p>
            ) : (
              <div className="space-y-4">
                {Object.entries(sportsByGroup).map(([group, groupSports]) => (
                  <div key={group}>
                    <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
                      {group}
                    </p>
                    <div className="space-y-1">
                      {groupSports.map((sport) => (
                        <div
                          key={sport.key}
                          className="flex items-center justify-between py-1.5 px-2 rounded hover:bg-muted/50"
                        >
                          <span className="text-sm">{sport.title}</span>
                          <Switch
                            checked={sport.is_active}
                            onCheckedChange={(checked: boolean) =>
                              toggleSportMutation.mutate({
                                key: sport.key,
                                isActive: checked,
                              })
                            }
                            disabled={toggleSportMutation.isPending}
                          />
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Recent Activity Logs */}
      <Card className="mt-6">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm uppercase tracking-wider text-muted-foreground">
              Recent Activity
            </CardTitle>
            <Button
              size="sm"
              variant="ghost"
              onClick={() => {
                queryClient.invalidateQueries({
                  queryKey: ["admin-automation-logs"],
                });
              }}
            >
              <RefreshCw className="size-4" />
            </Button>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Job</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Items</TableHead>
                <TableHead>Time</TableHead>
                <TableHead>Details</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {(!logs || logs.length === 0) ? (
                <TableRow>
                  <TableCell
                    colSpan={5}
                    className="h-24 text-center text-muted-foreground"
                  >
                    No automation logs yet. Trigger a job to get started.
                  </TableCell>
                </TableRow>
              ) : (
                logs.map((log) => (
                  <TableRow key={log.id}>
                    <TableCell className="font-medium text-sm">
                      {JOB_LABELS[log.job_name]?.label || log.job_name}
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={
                          log.status === "success" ? "default" : "destructive"
                        }
                      >
                        {log.status}
                      </Badge>
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {log.items_processed}
                    </TableCell>
                    <TableCell className="text-muted-foreground text-sm">
                      {new Date(log.created_at).toLocaleString()}
                    </TableCell>
                    <TableCell className="text-xs text-muted-foreground max-w-[200px] truncate">
                      {log.error_message ||
                        (log.details
                          ? JSON.stringify(log.details)
                          : "-")}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
