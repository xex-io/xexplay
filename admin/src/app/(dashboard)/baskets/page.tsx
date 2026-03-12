"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { locStr, asArray, type LocalizedString } from "@/lib/loc-str";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
import { Plus, Send, Pencil, Trash2 } from "lucide-react";
import { ActionsMenu } from "@/components/actions-menu";
import { DeleteDialog } from "@/components/delete-dialog";

interface Basket {
  id: string;
  basket_date: string;
  event_id: string;
  is_published: boolean;
  created_at: string;
}

interface EventItem {
  id: string;
  name?: LocalizedString;
  title?: LocalizedString;
}

export default function BasketsPage() {
  const queryClient = useQueryClient();
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [createForm, setCreateForm] = useState({
    event_id: "",
    basket_date: "",
    card_ids: "",
  });
  const [editBasket, setEditBasket] = useState<Basket | null>(null);
  const [deleteBasket, setDeleteBasket] = useState<Basket | null>(null);
  const [editForm, setEditForm] = useState({
    event_id: "",
    basket_date: "",
  });

  const { data: baskets = [], isLoading } = useQuery<Basket[]>({
    queryKey: ["admin-baskets"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/baskets");
      return asArray<Basket>(res);
    },
  });

  const { data: events = [] } = useQuery<EventItem[]>({
    queryKey: ["admin-events"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/events");
      return asArray<EventItem>(res);
    },
  });

  const eventMap = new Map(events.map((e) => [e.id, e]));

  const createMutation = useMutation({
    mutationFn: async (data: {
      event_id: string;
      basket_date: string;
      card_ids: string[];
    }) => {
      return apiClient.post("/admin/baskets", {
        event_id: data.event_id,
        basket_date: data.basket_date + "T00:00:00Z",
        card_ids: data.card_ids.filter(Boolean),
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-baskets"] });
      setShowCreateModal(false);
      setCreateForm({ event_id: "", basket_date: "", card_ids: "" });
    },
  });

  const publishMutation = useMutation({
    mutationFn: async (basketId: string) => {
      return apiClient.post(`/admin/baskets/${basketId}/publish`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-baskets"] });
    },
  });

  const editMutation = useMutation({
    mutationFn: async ({
      id,
      payload,
    }: {
      id: string;
      payload: { event_id: string; basket_date: string };
    }) => {
      return apiClient.put(`/admin/baskets/${id}`, {
        event_id: payload.event_id,
        basket_date: payload.basket_date + "T00:00:00Z",
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-baskets"] });
      setEditBasket(null);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      return apiClient.delete(`/admin/baskets/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-baskets"] });
      setDeleteBasket(null);
    },
  });

  function getEventLabel(eventId: string): string {
    const e = eventMap.get(eventId);
    return locStr(e?.name) || locStr(e?.title) || eventId.slice(0, 8);
  }

  function handleCreate() {
    const cardIds = createForm.card_ids
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
    createMutation.mutate({
      event_id: createForm.event_id,
      basket_date: createForm.basket_date,
      card_ids: cardIds,
    });
  }

  function openEdit(basket: Basket) {
    setEditForm({
      event_id: basket.event_id,
      basket_date: basket.basket_date.slice(0, 10),
    });
    setEditBasket(basket);
  }

  function handleEdit() {
    if (!editBasket) return;
    editMutation.mutate({
      id: editBasket.id,
      payload: {
        event_id: editForm.event_id,
        basket_date: editForm.basket_date,
      },
    });
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-foreground">Baskets</h1>
        <Button size="sm" onClick={() => setShowCreateModal(true)}>
          <Plus className="size-4" data-icon="inline-start" />
          Create Basket
        </Button>
      </div>

      <div className="rounded-lg border border-border bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
              <TableHead>Basket Date</TableHead>
              <TableHead>Event</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Created</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell
                  colSpan={6}
                  className="h-24 text-center text-muted-foreground"
                >
                  Loading baskets...
                </TableCell>
              </TableRow>
            ) : baskets.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={6}
                  className="h-24 text-center text-muted-foreground"
                >
                  No baskets found.
                </TableCell>
              </TableRow>
            ) : (
              baskets.map((basket) => (
                <TableRow key={basket.id}>
                  <TableCell className="font-mono text-muted-foreground">
                    {basket.id.slice(0, 8)}
                  </TableCell>
                  <TableCell>
                    {new Date(basket.basket_date).toLocaleDateString()}
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {getEventLabel(basket.event_id)}
                  </TableCell>
                  <TableCell>
                    {basket.is_published ? (
                      <Badge
                        variant="secondary"
                        className="bg-green-500/10 text-green-500"
                      >
                        Published
                      </Badge>
                    ) : (
                      <Badge variant="outline">Draft</Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {new Date(basket.created_at).toLocaleDateString()}
                  </TableCell>
                  <TableCell>
                    <ActionsMenu
                      items={[
                        {
                          label: "Edit",
                          icon: Pencil,
                          onClick: () => openEdit(basket),
                          disabled: basket.is_published,
                        },
                        {
                          label: "Publish",
                          icon: Send,
                          onClick: () => publishMutation.mutate(basket.id),
                          disabled:
                            basket.is_published || publishMutation.isPending,
                        },
                        {
                          label: "Delete",
                          icon: Trash2,
                          variant: "destructive",
                          onClick: () => setDeleteBasket(basket),
                          disabled: basket.is_published,
                        },
                      ]}
                    />
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Create Basket Dialog */}
      <Dialog open={showCreateModal} onOpenChange={setShowCreateModal}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Create Basket</DialogTitle>
            <DialogDescription>
              Create a new daily basket with cards for an event.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label>Event ID</Label>
              <Input
                type="text"
                value={createForm.event_id}
                onChange={(e) =>
                  setCreateForm({ ...createForm, event_id: e.target.value })
                }
                placeholder="UUID of the event"
              />
              {events.length > 0 && (
                <p className="text-xs text-muted-foreground">
                  Available:{" "}
                  {events
                    .slice(0, 5)
                    .map((e) => `${(locStr(e.name) || locStr(e.title) || "").slice(0, 20)} (${e.id.slice(0, 8)})`)
                    .join(", ")}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label>Basket Date</Label>
              <Input
                type="date"
                value={createForm.basket_date}
                onChange={(e) =>
                  setCreateForm({ ...createForm, basket_date: e.target.value })
                }
              />
            </div>

            <div className="space-y-2">
              <Label>Card IDs</Label>
              <Input
                type="text"
                value={createForm.card_ids}
                onChange={(e) =>
                  setCreateForm({ ...createForm, card_ids: e.target.value })
                }
                placeholder="Comma-separated card UUIDs (optional)"
              />
              <p className="text-xs text-muted-foreground">
                Enter card UUIDs separated by commas, or leave empty.
              </p>
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowCreateModal(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={handleCreate}
              disabled={
                createMutation.isPending ||
                !createForm.event_id ||
                !createForm.basket_date
              }
            >
              {createMutation.isPending ? "Creating..." : "Create"}
            </Button>
          </DialogFooter>
          {createMutation.isError && (
            <p className="text-sm text-destructive">
              Failed to create basket. Please try again.
            </p>
          )}
        </DialogContent>
      </Dialog>

      {/* Edit Basket Dialog */}
      <Dialog
        open={editBasket !== null}
        onOpenChange={(open) => {
          if (!open) setEditBasket(null);
        }}
      >
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Edit Basket</DialogTitle>
            <DialogDescription>
              Update the basket event and date.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label>Event</Label>
              {events.length > 0 ? (
                <select
                  className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  value={editForm.event_id}
                  onChange={(e) =>
                    setEditForm({ ...editForm, event_id: e.target.value })
                  }
                >
                  <option value="">Select an event</option>
                  {events.map((ev) => (
                    <option key={ev.id} value={ev.id}>
                      {locStr(ev.name) || locStr(ev.title) || ev.id.slice(0, 8)}
                    </option>
                  ))}
                </select>
              ) : (
                <Input
                  type="text"
                  value={editForm.event_id}
                  onChange={(e) =>
                    setEditForm({ ...editForm, event_id: e.target.value })
                  }
                  placeholder="UUID of the event"
                />
              )}
            </div>

            <div className="space-y-2">
              <Label>Basket Date</Label>
              <Input
                type="date"
                value={editForm.basket_date}
                onChange={(e) =>
                  setEditForm({ ...editForm, basket_date: e.target.value })
                }
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setEditBasket(null)}>
              Cancel
            </Button>
            <Button
              onClick={handleEdit}
              disabled={
                editMutation.isPending ||
                !editForm.event_id ||
                !editForm.basket_date
              }
            >
              {editMutation.isPending ? "Saving..." : "Save Changes"}
            </Button>
          </DialogFooter>
          {editMutation.isError && (
            <p className="text-sm text-destructive">
              Failed to update basket. Please try again.
            </p>
          )}
        </DialogContent>
      </Dialog>

      {/* Delete Basket Dialog */}
      <DeleteDialog
        open={deleteBasket !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteBasket(null);
        }}
        title="Delete Basket"
        description={`Are you sure you want to delete basket "${deleteBasket?.id.slice(0, 8)}"? This action cannot be undone.`}
        onConfirm={() => {
          if (deleteBasket) deleteMutation.mutate(deleteBasket.id);
        }}
        isPending={deleteMutation.isPending}
        isError={deleteMutation.isError}
      />
    </div>
  );
}
