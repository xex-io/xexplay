"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter,
  DialogClose,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Plus, Loader2 } from "lucide-react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";

interface Match {
  id: string;
  event_id: string;
  event_name?: string;
  home_team: string;
  away_team: string;
  status: string;
  kickoff_time: string;
  home_score?: number;
  away_score?: number;
}

interface Event {
  id: string;
  name: string;
}

function statusColor(status: string) {
  switch (status) {
    case "live":
      return "default";
    case "scheduled":
      return "secondary";
    case "completed":
      return "outline";
    default:
      return "secondary";
  }
}

export default function MatchesPage() {
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const [filterEventId, setFilterEventId] = useState<string>("");
  const [form, setForm] = useState({
    event_id: "",
    home_team: "",
    away_team: "",
    kickoff_time: "",
  });

  const { data: events } = useQuery<Event[]>({
    queryKey: ["admin", "events"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/events");
      return res.data.data;
    },
  });

  const { data: matches, isLoading } = useQuery<Match[]>({
    queryKey: ["admin", "matches", filterEventId],
    queryFn: async () => {
      const params = filterEventId ? { event_id: filterEventId } : {};
      const res = await apiClient.get("/admin/matches", { params });
      return res.data.data;
    },
  });

  const createMutation = useMutation({
    mutationFn: async (payload: typeof form) => {
      const res = await apiClient.post("/admin/matches", {
        event_id: payload.event_id,
        home_team: payload.home_team,
        away_team: payload.away_team,
        kickoff_time: payload.kickoff_time,
      });
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "matches"] });
      setOpen(false);
      setForm({ event_id: "", home_team: "", away_team: "", kickoff_time: "" });
    },
  });

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    createMutation.mutate(form);
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-foreground">Matches</h1>
        <div className="flex items-center gap-3">
          <Select value={filterEventId} onValueChange={(v) => setFilterEventId(v ?? "")}>
            <SelectTrigger className="w-[200px]">
              <SelectValue placeholder="All events" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All events</SelectItem>
              {events?.map((event) => (
                <SelectItem key={event.id} value={event.id}>
                  {event.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger render={<Button size="sm" />}>
              <Plus className="size-4 mr-1" />
              Create Match
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Create Match</DialogTitle>
              </DialogHeader>
              <form onSubmit={handleSubmit} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="event_id">Event</Label>
                  <Select
                    value={form.event_id}
                    onValueChange={(v) => setForm({ ...form, event_id: v ?? "" })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select event" />
                    </SelectTrigger>
                    <SelectContent>
                      {events?.map((event) => (
                        <SelectItem key={event.id} value={event.id}>
                          {event.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="home_team">Home Team</Label>
                    <Input
                      id="home_team"
                      value={form.home_team}
                      onChange={(e) =>
                        setForm({ ...form, home_team: e.target.value })
                      }
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="away_team">Away Team</Label>
                    <Input
                      id="away_team"
                      value={form.away_team}
                      onChange={(e) =>
                        setForm({ ...form, away_team: e.target.value })
                      }
                      required
                    />
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="kickoff_time">Kick-off Time</Label>
                  <Input
                    id="kickoff_time"
                    type="datetime-local"
                    value={form.kickoff_time}
                    onChange={(e) =>
                      setForm({ ...form, kickoff_time: e.target.value })
                    }
                    required
                  />
                </div>
                <DialogFooter>
                  <DialogClose render={<Button type="button" variant="outline" />}>
                    Cancel
                  </DialogClose>
                  <Button type="submit" disabled={createMutation.isPending}>
                    {createMutation.isPending && (
                      <Loader2 className="size-4 mr-1 animate-spin" />
                    )}
                    Create
                  </Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>
        </div>
      </div>
      <Card className="overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Event</TableHead>
              <TableHead>Teams</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Kick-off</TableHead>
              <TableHead>Score</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell
                  colSpan={5}
                  className="py-12 text-center text-muted-foreground"
                >
                  <Loader2 className="size-5 animate-spin inline-block mr-2" />
                  Loading matches...
                </TableCell>
              </TableRow>
            ) : !matches || matches.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={5}
                  className="py-12 text-center text-muted-foreground"
                >
                  No matches found.
                </TableCell>
              </TableRow>
            ) : (
              matches.map((match) => (
                <TableRow key={match.id}>
                  <TableCell className="text-muted-foreground">
                    {match.event_name ?? match.event_id}
                  </TableCell>
                  <TableCell className="font-medium">
                    {match.home_team} vs {match.away_team}
                  </TableCell>
                  <TableCell>
                    <Badge variant={statusColor(match.status)}>
                      {match.status}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {new Date(match.kickoff_time).toLocaleString()}
                  </TableCell>
                  <TableCell>
                    {match.home_score !== undefined &&
                    match.away_score !== undefined
                      ? `${match.home_score} - ${match.away_score}`
                      : "--"}
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
