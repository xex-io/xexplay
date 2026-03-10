"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
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
import { Download } from "lucide-react";

interface LeaderboardEntry {
  rank: number;
  user_id: string;
  username: string;
  points: number;
  correct_answers: number;
  wrong_answers: number;
}

const tabs = ["daily", "weekly", "tournament", "all-time"] as const;
type TabType = (typeof tabs)[number];

const tabLabels: Record<TabType, string> = {
  daily: "Daily",
  weekly: "Weekly",
  tournament: "Tournament",
  "all-time": "All-Time",
};

function todayKey(): string {
  return new Date().toISOString().slice(0, 10);
}

function exportCSV(entries: LeaderboardEntry[], type: string, periodKey: string) {
  const header = "Rank,User,Points,Correct Answers,Wrong Answers";
  const rows = entries.map(
    (e) => `${e.rank},${e.username},${e.points},${e.correct_answers},${e.wrong_answers}`
  );
  const csv = [header, ...rows].join("\n");
  const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `leaderboard-${type}-${periodKey || "all"}.csv`;
  a.click();
  URL.revokeObjectURL(url);
}

function LeaderboardTable({
  entries,
  isLoading,
}: {
  entries: LeaderboardEntry[];
  isLoading: boolean;
}) {
  return (
    <div className="rounded-lg border border-border bg-card">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Rank</TableHead>
            <TableHead>User</TableHead>
            <TableHead>Points</TableHead>
            <TableHead>Correct Answers</TableHead>
            <TableHead>Wrong Answers</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {isLoading ? (
            <TableRow>
              <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                Loading leaderboard...
              </TableCell>
            </TableRow>
          ) : entries.length === 0 ? (
            <TableRow>
              <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                No leaderboard data found.
              </TableCell>
            </TableRow>
          ) : (
            entries.map((entry) => (
              <TableRow key={entry.user_id}>
                <TableCell className="font-semibold">#{entry.rank}</TableCell>
                <TableCell>{entry.username || entry.user_id.slice(0, 8)}</TableCell>
                <TableCell className="font-mono">{entry.points.toLocaleString()}</TableCell>
                <TableCell>
                  <Badge variant="secondary" className="bg-green-500/10 text-green-500">
                    {entry.correct_answers}
                  </Badge>
                </TableCell>
                <TableCell>
                  <Badge variant="destructive">
                    {entry.wrong_answers}
                  </Badge>
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}

export default function LeaderboardsPage() {
  const [activeTab, setActiveTab] = useState<TabType>("daily");
  const [periodKey, setPeriodKey] = useState(todayKey());

  const needsPeriod = activeTab === "daily" || activeTab === "weekly";

  const { data: entries = [], isLoading } = useQuery<LeaderboardEntry[]>({
    queryKey: ["admin-leaderboards", activeTab, needsPeriod ? periodKey : null],
    queryFn: async () => {
      const params = needsPeriod ? `?period_key=${periodKey}` : "";
      const res = await apiClient.get(`/admin/leaderboards/${activeTab}${params}`);
      return res.data?.data ?? res.data ?? [];
    },
  });

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-foreground">Leaderboards</h1>
        <Button
          onClick={() => exportCSV(entries, activeTab, periodKey)}
          disabled={entries.length === 0}
          variant="outline"
        >
          <Download data-icon="inline-start" />
          Export CSV
        </Button>
      </div>

      <Tabs
        value={activeTab}
        onValueChange={(val) => setActiveTab(val as TabType)}
      >
        <div className="flex items-center gap-4 mb-4">
          <TabsList>
            {tabs.map((tab) => (
              <TabsTrigger key={tab} value={tab}>
                {tabLabels[tab]}
              </TabsTrigger>
            ))}
          </TabsList>

          {needsPeriod && (
            <div className="flex items-center gap-2">
              <Label className="text-xs uppercase tracking-wider text-muted-foreground">
                Period
              </Label>
              <Input
                type="date"
                value={periodKey}
                onChange={(e) => setPeriodKey(e.target.value)}
                className="w-auto"
              />
            </div>
          )}
        </div>

        {tabs.map((tab) => (
          <TabsContent key={tab} value={tab}>
            <LeaderboardTable entries={entries} isLoading={isLoading} />
          </TabsContent>
        ))}
      </Tabs>
    </div>
  );
}
