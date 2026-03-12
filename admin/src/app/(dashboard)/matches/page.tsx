"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
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
import { Plus, Loader2, Pencil, Trash2, Trophy } from "lucide-react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { locStr, asArray, type LocalizedString } from "@/lib/loc-str";
import { ActionsMenu } from "@/components/actions-menu";
import { DeleteDialog } from "@/components/delete-dialog";

interface Match {
  id: string;
  event_id: string;
  event_name?: LocalizedString;
  home_team: string;
  away_team: string;
  status: string;
  kickoff_time: string;
  home_score?: number;
  away_score?: number;
  result_data?: Record<string, unknown>;
  source?: string;
  created_at?: string;
}

interface Event {
  id: string;
  name: LocalizedString;
}

function statusColor(status: string) {
  switch (status) {
    case "live":
      return "default";
    case "scheduled":
    case "upcoming":
      return "secondary";
    case "completed":
      return "outline";
    case "cancelled":
      return "destructive";
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

  const [editMatch, setEditMatch] = useState<Match | null>(null);
  const [deleteMatch, setDeleteMatch] = useState<Match | null>(null);
  const [scoreMatch, setScoreMatch] = useState<Match | null>(null);

  const [editForm, setEditForm] = useState({
    event_id: "",
    home_team: "",
    away_team: "",
    kickoff_time: "",
    status: "",
  });

  const [scoreForm, setScoreForm] = useState({
    home_score: 0,
    away_score: 0,
    markCompleted: false,
  });

  const { data: events } = useQuery<Event[]>({
    queryKey: ["admin", "events"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/events");
      return asArray<Event>(res);
    },
  });

  const { data: matches, isLoading } = useQuery<Match[]>({
    queryKey: ["admin", "matches", filterEventId],
    queryFn: async () => {
      const params = filterEventId ? { event_id: filterEventId } : {};
      const res = await apiClient.get("/admin/matches", { params });
      return asArray<Match>(res);
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

  const editMutation = useMutation({
    mutationFn: async ({
      id,
      payload,
    }: {
      id: string;
      payload: typeof editForm;
    }) => {
      const res = await apiClient.put(`/admin/matches/${id}`, payload);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "matches"] });
      setEditMatch(null);
    },
  });

  const scoreMutation = useMutation({
    mutationFn: async ({
      id,
      payload,
    }: {
      id: string;
      payload: { home_score: number; away_score: number; status?: string };
    }) => {
      const res = await apiClient.put(`/admin/matches/${id}`, payload);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "matches"] });
      setScoreMatch(null);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      await apiClient.delete(`/admin/matches/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "matches"] });
      setDeleteMatch(null);
    },
  });

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    createMutation.mutate(form);
  }

  function handleEditSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!editMatch) return;
    editMutation.mutate({ id: editMatch.id, payload: editForm });
  }

  function handleScoreSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!scoreMatch) return;
    const payload: { home_score: number; away_score: number; status?: string } =
      {
        home_score: scoreForm.home_score,
        away_score: scoreForm.away_score,
      };
    if (scoreForm.markCompleted) {
      payload.status = "completed";
    }
    scoreMutation.mutate({ id: scoreMatch.id, payload });
  }

  function openEditDialog(match: Match) {
    setEditForm({
      event_id: match.event_id,
      home_team: match.home_team,
      away_team: match.away_team,
      kickoff_time: match.kickoff_time
        ? new Date(match.kickoff_time).toISOString().slice(0, 16)
        : "",
      status: match.status,
    });
    setEditMatch(match);
  }

  function openScoreDialog(match: Match) {
    setScoreForm({
      home_score: match.home_score ?? 0,
      away_score: match.away_score ?? 0,
      markCompleted: false,
    });
    setScoreMatch(match);
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
                  {locStr(event.name)}
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
                          {locStr(event.name)}
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
              <TableHead className="w-[60px]">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell
                  colSpan={6}
                  className="py-12 text-center text-muted-foreground"
                >
                  <Loader2 className="size-5 animate-spin inline-block mr-2" />
                  Loading matches...
                </TableCell>
              </TableRow>
            ) : !matches || matches.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={6}
                  className="py-12 text-center text-muted-foreground"
                >
                  No matches found.
                </TableCell>
              </TableRow>
            ) : (
              matches.map((match) => (
                <TableRow key={match.id}>
                  <TableCell className="text-muted-foreground">
                    {locStr(match.event_name) || match.event_id}
                  </TableCell>
                  <TableCell className="font-medium">
                    {match.home_team} vs {match.away_team}
                    {match.source === "auto" && <Badge variant="outline" className="ml-2 text-xs">Auto</Badge>}
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
                  <TableCell>
                    <ActionsMenu
                      items={[
                        {
                          label: "Edit",
                          icon: Pencil,
                          onClick: () => openEditDialog(match),
                        },
                        {
                          label: "Update Score",
                          icon: Trophy,
                          onClick: () => openScoreDialog(match),
                        },
                        {
                          label: "Delete",
                          icon: Trash2,
                          variant: "destructive",
                          onClick: () => setDeleteMatch(match),
                        },
                      ]}
                    />
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </Card>

      {/* Edit Match Dialog */}
      <Dialog
        open={editMatch !== null}
        onOpenChange={(open) => {
          if (!open) setEditMatch(null);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Match</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleEditSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="edit_event_id">Event</Label>
              <Select
                value={editForm.event_id}
                onValueChange={(v) =>
                  setEditForm({ ...editForm, event_id: v ?? "" })
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select event" />
                </SelectTrigger>
                <SelectContent>
                  {events?.map((event) => (
                    <SelectItem key={event.id} value={event.id}>
                      {locStr(event.name)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="edit_home_team">Home Team</Label>
                <Input
                  id="edit_home_team"
                  value={editForm.home_team}
                  onChange={(e) =>
                    setEditForm({ ...editForm, home_team: e.target.value })
                  }
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="edit_away_team">Away Team</Label>
                <Input
                  id="edit_away_team"
                  value={editForm.away_team}
                  onChange={(e) =>
                    setEditForm({ ...editForm, away_team: e.target.value })
                  }
                  required
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit_kickoff_time">Kick-off Time</Label>
              <Input
                id="edit_kickoff_time"
                type="datetime-local"
                value={editForm.kickoff_time}
                onChange={(e) =>
                  setEditForm({ ...editForm, kickoff_time: e.target.value })
                }
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit_status">Status</Label>
              <Select
                value={editForm.status}
                onValueChange={(v) =>
                  setEditForm({ ...editForm, status: v ?? "" })
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="upcoming">Upcoming</SelectItem>
                  <SelectItem value="live">Live</SelectItem>
                  <SelectItem value="completed">Completed</SelectItem>
                  <SelectItem value="cancelled">Cancelled</SelectItem>
                </SelectContent>
              </Select>
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

      {/* Score Update Dialog */}
      <Dialog
        open={scoreMatch !== null}
        onOpenChange={(open) => {
          if (!open) setScoreMatch(null);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              Update Score
              {scoreMatch &&
                ` - ${scoreMatch.home_team} vs ${scoreMatch.away_team}`}
            </DialogTitle>
          </DialogHeader>
          <form onSubmit={handleScoreSubmit} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="home_score">
                  {scoreMatch?.home_team ?? "Home"} Score
                </Label>
                <Input
                  id="home_score"
                  type="number"
                  min={0}
                  value={scoreForm.home_score}
                  onChange={(e) =>
                    setScoreForm({
                      ...scoreForm,
                      home_score: parseInt(e.target.value, 10) || 0,
                    })
                  }
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="away_score">
                  {scoreMatch?.away_team ?? "Away"} Score
                </Label>
                <Input
                  id="away_score"
                  type="number"
                  min={0}
                  value={scoreForm.away_score}
                  onChange={(e) =>
                    setScoreForm({
                      ...scoreForm,
                      away_score: parseInt(e.target.value, 10) || 0,
                    })
                  }
                  required
                />
              </div>
            </div>
            <div className="flex items-center gap-3">
              <Switch
                checked={scoreForm.markCompleted}
                onCheckedChange={(checked) =>
                  setScoreForm({ ...scoreForm, markCompleted: !!checked })
                }
              />
              <Label>Mark as completed</Label>
            </div>
            <DialogFooter>
              <DialogClose render={<Button type="button" variant="outline" />}>
                Cancel
              </DialogClose>
              <Button type="submit" disabled={scoreMutation.isPending}>
                {scoreMutation.isPending && (
                  <Loader2 className="size-4 mr-1 animate-spin" />
                )}
                Update Score
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <DeleteDialog
        open={deleteMatch !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteMatch(null);
        }}
        title="Delete Match"
        description={
          deleteMatch
            ? `Are you sure you want to delete the match "${deleteMatch.home_team} vs ${deleteMatch.away_team}"? This action cannot be undone. Matches with associated cards cannot be deleted.`
            : ""
        }
        onConfirm={() => {
          if (deleteMatch) deleteMutation.mutate(deleteMatch.id);
        }}
        isPending={deleteMutation.isPending}
        isError={deleteMutation.isError}
      />
    </div>
  );
}
