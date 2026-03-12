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
import { Plus, Loader2, Pencil, Trash2 } from "lucide-react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { locStr, asArray, type LocalizedString } from "@/lib/loc-str";
import { ActionsMenu, type ActionItem } from "@/components/actions-menu";
import { DeleteDialog } from "@/components/delete-dialog";

interface Event {
  id: string;
  name: LocalizedString;
  description: LocalizedString;
  slug: string;
  status: string;
  start_date: string;
  end_date: string;
  is_active?: boolean;
  source?: string;
  scoring_multiplier: number;
  created_at: string;
}

function statusColor(status: string) {
  switch (status) {
    case "active":
      return "default";
    case "upcoming":
      return "secondary";
    case "completed":
      return "outline";
    default:
      return "secondary";
  }
}

const emptyForm = {
  name: "",
  description: "",
  slug: "",
  start_date: "",
  end_date: "",
  scoring_multiplier: "1",
};

export default function EventsPage() {
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const [form, setForm] = useState(emptyForm);

  const [editEvent, setEditEvent] = useState<Event | null>(null);
  const [editForm, setEditForm] = useState({
    ...emptyForm,
    is_active: true,
  });

  const [deleteEvent, setDeleteEvent] = useState<Event | null>(null);

  const { data: events, isLoading } = useQuery<Event[]>({
    queryKey: ["admin", "events"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/events");
      return asArray<Event>(res);
    },
  });

  const createMutation = useMutation({
    mutationFn: async (payload: typeof form) => {
      const res = await apiClient.post("/admin/events", {
        name: payload.name,
        description: payload.description,
        slug: payload.slug,
        start_date: payload.start_date,
        end_date: payload.end_date,
        scoring_multiplier: parseFloat(payload.scoring_multiplier),
      });
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "events"] });
      setOpen(false);
      setForm(emptyForm);
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
      const res = await apiClient.put(`/admin/events/${id}`, {
        name: payload.name,
        description: payload.description,
        slug: payload.slug,
        start_date: payload.start_date,
        end_date: payload.end_date,
        scoring_multiplier: parseFloat(payload.scoring_multiplier),
        is_active: payload.is_active,
      });
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "events"] });
      setEditEvent(null);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      await apiClient.delete(`/admin/events/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "events"] });
      setDeleteEvent(null);
    },
  });

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    createMutation.mutate(form);
  }

  function handleEditSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!editEvent) return;
    editMutation.mutate({ id: editEvent.id, payload: editForm });
  }

  function openEdit(event: Event) {
    setEditForm({
      name: locStr(event.name),
      description: locStr(event.description),
      slug: event.slug,
      start_date: new Date(event.start_date).toISOString().slice(0, 16),
      end_date: new Date(event.end_date).toISOString().slice(0, 16),
      scoring_multiplier: String(event.scoring_multiplier),
      is_active: event.is_active ?? true,
    });
    setEditEvent(event);
  }

  function actionsFor(event: Event): ActionItem[] {
    return [
      {
        label: "Edit",
        icon: Pencil,
        onClick: () => openEdit(event),
      },
      {
        label: "Delete",
        icon: Trash2,
        variant: "destructive",
        onClick: () => setDeleteEvent(event),
      },
    ];
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-foreground">Events</h1>
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger render={<Button size="sm" />}>
            <Plus className="size-4 mr-1" />
            Create Event
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create Event</DialogTitle>
            </DialogHeader>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                <Input
                  id="name"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                <Input
                  id="description"
                  value={form.description}
                  onChange={(e) =>
                    setForm({ ...form, description: e.target.value })
                  }
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="slug">Slug</Label>
                <Input
                  id="slug"
                  value={form.slug}
                  onChange={(e) => setForm({ ...form, slug: e.target.value })}
                  required
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="start_date">Start Date</Label>
                  <Input
                    id="start_date"
                    type="datetime-local"
                    value={form.start_date}
                    onChange={(e) =>
                      setForm({ ...form, start_date: e.target.value })
                    }
                    required
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="end_date">End Date</Label>
                  <Input
                    id="end_date"
                    type="datetime-local"
                    value={form.end_date}
                    onChange={(e) =>
                      setForm({ ...form, end_date: e.target.value })
                    }
                    required
                  />
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="scoring_multiplier">Scoring Multiplier</Label>
                <Input
                  id="scoring_multiplier"
                  type="number"
                  step="0.1"
                  min="0.1"
                  value={form.scoring_multiplier}
                  onChange={(e) =>
                    setForm({ ...form, scoring_multiplier: e.target.value })
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

      {/* Edit Dialog */}
      <Dialog
        open={editEvent !== null}
        onOpenChange={(open) => {
          if (!open) setEditEvent(null);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Event</DialogTitle>
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
            <div className="space-y-2">
              <Label htmlFor="edit-slug">Slug</Label>
              <Input
                id="edit-slug"
                value={editForm.slug}
                onChange={(e) =>
                  setEditForm({ ...editForm, slug: e.target.value })
                }
                required
              />
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
            <div className="space-y-2">
              <Label htmlFor="edit-scoring_multiplier">Scoring Multiplier</Label>
              <Input
                id="edit-scoring_multiplier"
                type="number"
                step="0.1"
                min="0.1"
                value={editForm.scoring_multiplier}
                onChange={(e) =>
                  setEditForm({
                    ...editForm,
                    scoring_multiplier: e.target.value,
                  })
                }
                required
              />
            </div>
            <div className="flex items-center justify-between">
              <Label htmlFor="edit-is_active">Active</Label>
              <Switch
                id="edit-is_active"
                checked={editForm.is_active}
                onCheckedChange={(checked) =>
                  setEditForm({ ...editForm, is_active: checked })
                }
              />
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

      {/* Delete Dialog */}
      <DeleteDialog
        open={deleteEvent !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteEvent(null);
        }}
        title="Delete Event"
        description={`Are you sure you want to delete "${deleteEvent ? locStr(deleteEvent.name) : ""}"? This action cannot be undone.`}
        onConfirm={() => {
          if (deleteEvent) deleteMutation.mutate(deleteEvent.id);
        }}
        isPending={deleteMutation.isPending}
        isError={deleteMutation.isError}
      />

      <Card className="overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Slug</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Dates</TableHead>
              <TableHead>Multiplier</TableHead>
              <TableHead className="w-12">
                <span className="sr-only">Actions</span>
              </TableHead>
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
                  Loading events...
                </TableCell>
              </TableRow>
            ) : !events || events.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={6}
                  className="py-12 text-center text-muted-foreground"
                >
                  No events found.
                </TableCell>
              </TableRow>
            ) : (
              events.map((event) => (
                <TableRow key={event.id}>
                  <TableCell className="font-medium">
                    {locStr(event.name)}
                    {event.source === "auto" && <Badge variant="outline" className="ml-2 text-xs">Auto</Badge>}
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {event.slug}
                  </TableCell>
                  <TableCell>
                    <Badge variant={statusColor(event.status)}>
                      {event.status}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {new Date(event.start_date).toLocaleDateString()} &ndash;{" "}
                    {new Date(event.end_date).toLocaleDateString()}
                  </TableCell>
                  <TableCell>{event.scoring_multiplier}x</TableCell>
                  <TableCell>
                    <ActionsMenu items={actionsFor(event)} />
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
