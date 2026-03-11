"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
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
import { Plus, Send } from "lucide-react";

interface Basket {
  id: string;
  basket_date: string;
  event_id: string;
  is_published: boolean;
  created_at: string;
}

interface EventItem {
  id: string;
  name?: string;
  title?: string;
}

export default function BasketsPage() {
  const queryClient = useQueryClient();
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [createForm, setCreateForm] = useState({
    event_id: "",
    basket_date: "",
    card_ids: "",
  });

  const { data: baskets = [], isLoading } = useQuery<Basket[]>({
    queryKey: ["admin-baskets"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/baskets");
      return res.data?.data ?? res.data ?? [];
    },
  });

  const { data: events = [] } = useQuery<EventItem[]>({
    queryKey: ["admin-events"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/events");
      return res.data?.data ?? res.data ?? [];
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

  function getEventLabel(eventId: string): string {
    const e = eventMap.get(eventId);
    return e?.name || e?.title || eventId.slice(0, 8);
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
                    {!basket.is_published && (
                      <Button
                        size="xs"
                        variant="secondary"
                        onClick={() => publishMutation.mutate(basket.id)}
                        disabled={publishMutation.isPending}
                      >
                        <Send className="size-3" data-icon="inline-start" />
                        Publish
                      </Button>
                    )}
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
                    .map((e) => `${(e.name || e.title || "").slice(0, 20)} (${e.id.slice(0, 8)})`)
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
    </div>
  );
}
