"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { asArray } from "@/lib/loc-str";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
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
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Check, Plus, AlertCircle, Pencil, Trash2 } from "lucide-react";
import { ActionsMenu } from "@/components/actions-menu";
import { DeleteDialog } from "@/components/delete-dialog";

interface CardItem {
  id: string;
  match_id: string;
  question_text: Record<string, string>;
  tier: "gold" | "silver" | "white" | "vip";
  high_answer_is_yes: boolean | null;
  correct_answer: boolean | null;
  is_resolved: boolean;
  available_date: string;
  expires_at: string;
  source?: string;
  created_at: string;
  updated_at: string;
}

interface Match {
  id: string;
  home_team: string;
  away_team: string;
}

const tierVariant: Record<string, "default" | "secondary" | "outline"> = {
  gold: "default",
  silver: "secondary",
  white: "outline",
  vip: "default",
};

function getQuestionText(qt: Record<string, string>): string {
  if (typeof qt === "string") return qt;
  return qt?.en || qt?.fa || Object.values(qt || {})[0] || "";
}

function truncate(text: string, len: number): string {
  return text.length > len ? text.slice(0, len) + "..." : text;
}

function toDateInputValue(isoString: string): string {
  if (!isoString) return "";
  return isoString.slice(0, 10);
}

export default function CardsPage() {
  const queryClient = useQueryClient();
  const [dateFilter, setDateFilter] = useState("");
  const [resolveModal, setResolveModal] = useState<CardItem | null>(null);
  const [selectedAnswer, setSelectedAnswer] = useState<boolean | null>(null);
  const [confirmStep, setConfirmStep] = useState(false);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editCard, setEditCard] = useState<CardItem | null>(null);
  const [deleteCard, setDeleteCard] = useState<CardItem | null>(null);
  const [createForm, setCreateForm] = useState({
    match_id: "",
    question_en: "",
    question_fa: "",
    question_ar: "",
    tier: "white",
    available_date: "",
    expires_at: "",
  });
  const [editForm, setEditForm] = useState({
    question_en: "",
    tier: "white",
    available_date: "",
    expires_at: "",
  });

  const { data: cards = [], isLoading } = useQuery<CardItem[]>({
    queryKey: ["admin-cards", dateFilter],
    queryFn: async () => {
      const params: Record<string, string> = {};
      if (dateFilter) params.date = dateFilter;
      const res = await apiClient.get("/admin/cards", { params });
      return asArray<CardItem>(res);
    },
  });

  const { data: matches = [] } = useQuery<Match[]>({
    queryKey: ["admin-matches"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/matches");
      return asArray<Match>(res);
    },
  });

  const matchMap = new Map(matches.map((m) => [m.id, m]));

  const resolveMutation = useMutation({
    mutationFn: async ({
      cardId,
      correctAnswer,
    }: {
      cardId: string;
      correctAnswer: boolean;
    }) => {
      return apiClient.post(`/admin/cards/${cardId}/resolve`, {
        correct_answer: correctAnswer,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-cards"] });
      closeModal();
    },
  });

  const createMutation = useMutation({
    mutationFn: async (data: {
      match_id: string;
      question_text: Record<string, string>;
      tier: string;
      available_date: string;
      expires_at: string;
    }) => {
      return apiClient.post("/admin/cards", {
        match_id: data.match_id,
        question_text: data.question_text,
        tier: data.tier,
        available_date: data.available_date,
        expires_at: data.expires_at,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-cards"] });
      setShowCreateModal(false);
      setCreateForm({
        match_id: "",
        question_en: "",
        question_fa: "",
        question_ar: "",
        tier: "white",
        available_date: "",
        expires_at: "",
      });
    },
  });

  const editMutation = useMutation({
    mutationFn: async ({
      id,
      payload,
    }: {
      id: string;
      payload: {
        question_text: Record<string, string>;
        tier: string;
        available_date: string;
        expires_at: string;
      };
    }) => {
      return apiClient.put(`/admin/cards/${id}`, payload);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-cards"] });
      setEditCard(null);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      return apiClient.delete(`/admin/cards/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-cards"] });
      setDeleteCard(null);
    },
  });

  function openModal(card: CardItem) {
    setResolveModal(card);
    setSelectedAnswer(null);
    setConfirmStep(false);
  }

  function closeModal() {
    setResolveModal(null);
    setSelectedAnswer(null);
    setConfirmStep(false);
  }

  function handleSelectAnswer(answer: boolean) {
    setSelectedAnswer(answer);
    setConfirmStep(true);
  }

  function handleConfirmResolve() {
    if (resolveModal && selectedAnswer !== null) {
      resolveMutation.mutate({
        cardId: resolveModal.id,
        correctAnswer: selectedAnswer,
      });
    }
  }

  function handleCreateCard() {
    const questionText: Record<string, string> = {};
    if (createForm.question_en) questionText.en = createForm.question_en;
    if (createForm.question_fa) questionText.fa = createForm.question_fa;
    if (createForm.question_ar) questionText.ar = createForm.question_ar;

    createMutation.mutate({
      match_id: createForm.match_id,
      question_text: questionText,
      tier: createForm.tier,
      available_date: createForm.available_date
        ? createForm.available_date + "T00:00:00Z"
        : "",
      expires_at: createForm.expires_at
        ? createForm.expires_at + "T23:59:59Z"
        : "",
    });
  }

  function openEditDialog(card: CardItem) {
    setEditForm({
      question_en: getQuestionText(card.question_text),
      tier: card.tier,
      available_date: toDateInputValue(card.available_date),
      expires_at: toDateInputValue(card.expires_at),
    });
    setEditCard(card);
  }

  function handleEditCard() {
    if (!editCard) return;
    const questionText: Record<string, string> = {};
    if (editForm.question_en) questionText.en = editForm.question_en;

    editMutation.mutate({
      id: editCard.id,
      payload: {
        question_text: questionText,
        tier: editForm.tier,
        available_date: editForm.available_date
          ? editForm.available_date + "T00:00:00Z"
          : "",
        expires_at: editForm.expires_at
          ? editForm.expires_at + "T23:59:59Z"
          : "",
      },
    });
  }

  function handleDeleteCard() {
    if (!deleteCard) return;
    deleteMutation.mutate(deleteCard.id);
  }

  function getMatchLabel(matchId: string): string {
    const m = matchMap.get(matchId);
    return m ? `${m.home_team} vs ${m.away_team}` : matchId.slice(0, 8);
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-foreground">Cards</h1>
        <div className="flex items-center gap-3">
          <Input
            type="date"
            value={dateFilter}
            onChange={(e) => setDateFilter(e.target.value)}
            className="w-44"
            placeholder="Filter by date"
          />
          {dateFilter && (
            <Button
              size="xs"
              variant="ghost"
              onClick={() => setDateFilter("")}
            >
              Clear
            </Button>
          )}
          <Button size="sm" onClick={() => setShowCreateModal(true)}>
            <Plus className="size-4" data-icon="inline-start" />
            Create Card
          </Button>
        </div>
      </div>

      <Card className="overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
              <TableHead>Question</TableHead>
              <TableHead>Tier</TableHead>
              <TableHead>Match</TableHead>
              <TableHead>Available Date</TableHead>
              <TableHead>Resolved</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell
                  colSpan={7}
                  className="py-12 text-center text-muted-foreground"
                >
                  Loading cards...
                </TableCell>
              </TableRow>
            ) : cards.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={7}
                  className="py-12 text-center text-muted-foreground"
                >
                  No cards found.
                </TableCell>
              </TableRow>
            ) : (
              cards.map((card) => (
                <TableRow key={card.id}>
                  <TableCell className="font-mono text-muted-foreground">
                    {card.id.slice(0, 8)}
                  </TableCell>
                  <TableCell className="max-w-xs">
                    {truncate(getQuestionText(card.question_text), 50)}
                  </TableCell>
                  <TableCell>
                    <Badge
                      variant={tierVariant[card.tier] || "outline"}
                      className="capitalize"
                    >
                      {card.tier}
                    </Badge>
                    {card.source === "ai" && <Badge variant="outline" className="ml-1 text-xs">AI</Badge>}
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {getMatchLabel(card.match_id)}
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {new Date(card.available_date).toLocaleDateString()}
                  </TableCell>
                  <TableCell>
                    {card.is_resolved ? (
                      <span className="inline-flex items-center gap-1.5 text-green-500">
                        <Check className="size-4" />
                        {card.correct_answer === true
                          ? "Yes"
                          : card.correct_answer === false
                            ? "No"
                            : "-"}
                      </span>
                    ) : (
                      <span className="text-muted-foreground">Pending</span>
                    )}
                  </TableCell>
                  <TableCell>
                    <ActionsMenu
                      items={[
                        {
                          label: "Edit",
                          icon: Pencil,
                          onClick: () => openEditDialog(card),
                          disabled: card.is_resolved,
                        },
                        {
                          label: "Resolve",
                          icon: Check,
                          onClick: () => openModal(card),
                          disabled: card.is_resolved,
                        },
                        {
                          label: "Delete",
                          icon: Trash2,
                          onClick: () => setDeleteCard(card),
                          variant: "destructive",
                          disabled: card.is_resolved,
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

      {/* Resolution Modal */}
      <Dialog
        open={resolveModal !== null}
        onOpenChange={(open) => {
          if (!open) closeModal();
        }}
      >
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Resolve Card</DialogTitle>
            <DialogDescription>
              Set the correct answer for this prediction card. This action
              cannot be undone.
            </DialogDescription>
          </DialogHeader>

          {resolveModal && (
            <div className="space-y-4">
              <div>
                <Label className="text-xs uppercase tracking-wider text-muted-foreground">
                  Question
                </Label>
                <p className="mt-1 text-sm text-foreground leading-relaxed">
                  {getQuestionText(resolveModal.question_text)}
                </p>
              </div>

              <div className="flex gap-4">
                <div>
                  <Label className="text-xs uppercase tracking-wider text-muted-foreground">
                    Tier
                  </Label>
                  <p className="mt-1">
                    <Badge
                      variant={tierVariant[resolveModal.tier] || "outline"}
                      className="capitalize"
                    >
                      {resolveModal.tier}
                    </Badge>
                  </p>
                </div>
                <div>
                  <Label className="text-xs uppercase tracking-wider text-muted-foreground">
                    Match
                  </Label>
                  <p className="mt-1 text-sm text-foreground">
                    {getMatchLabel(resolveModal.match_id)}
                  </p>
                </div>
              </div>

              {!confirmStep ? (
                <div>
                  <p className="text-sm text-muted-foreground mb-3">
                    What is the correct answer?
                  </p>
                  <div className="flex gap-3">
                    <Button
                      variant="outline"
                      className="flex-1 border-green-500/30 text-green-500 hover:bg-green-500/10 hover:text-green-400"
                      onClick={() => handleSelectAnswer(true)}
                    >
                      Yes
                    </Button>
                    <Button
                      variant="outline"
                      className="flex-1 border-destructive/30 text-destructive hover:bg-destructive/10"
                      onClick={() => handleSelectAnswer(false)}
                    >
                      No
                    </Button>
                  </div>
                </div>
              ) : (
                <div>
                  <Alert>
                    <AlertCircle className="size-4" />
                    <AlertDescription>
                      Are you sure the correct answer is{" "}
                      <span
                        className={`font-bold ${selectedAnswer ? "text-green-500" : "text-destructive"}`}
                      >
                        {selectedAnswer ? "YES" : "NO"}
                      </span>
                      ?
                    </AlertDescription>
                  </Alert>
                  <DialogFooter className="mt-4">
                    <Button
                      variant="outline"
                      onClick={() => {
                        setConfirmStep(false);
                        setSelectedAnswer(null);
                      }}
                    >
                      Go Back
                    </Button>
                    <Button
                      onClick={handleConfirmResolve}
                      disabled={resolveMutation.isPending}
                    >
                      {resolveMutation.isPending ? "Resolving..." : "Confirm"}
                    </Button>
                  </DialogFooter>
                  {resolveMutation.isError && (
                    <Alert variant="destructive" className="mt-3">
                      <AlertCircle className="size-4" />
                      <AlertDescription>
                        Failed to resolve card. Please try again.
                      </AlertDescription>
                    </Alert>
                  )}
                </div>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>

      {/* Create Card Dialog */}
      <Dialog open={showCreateModal} onOpenChange={setShowCreateModal}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Create Card</DialogTitle>
            <DialogDescription>
              Create a new prediction card for a match.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label>Match ID</Label>
              <Input
                type="text"
                value={createForm.match_id}
                onChange={(e) =>
                  setCreateForm({ ...createForm, match_id: e.target.value })
                }
                placeholder="UUID of the match"
              />
              {matches.length > 0 && (
                <p className="text-xs text-muted-foreground">
                  Available:{" "}
                  {matches
                    .slice(0, 5)
                    .map(
                      (m) =>
                        `${m.home_team} vs ${m.away_team} (${m.id.slice(0, 8)})`
                    )
                    .join(", ")}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label>Question (English)</Label>
              <Textarea
                value={createForm.question_en}
                onChange={(e) =>
                  setCreateForm({ ...createForm, question_en: e.target.value })
                }
                placeholder="Will Team A win?"
              />
            </div>

            <div className="space-y-2">
              <Label>Question (Farsi)</Label>
              <Textarea
                value={createForm.question_fa}
                onChange={(e) =>
                  setCreateForm({ ...createForm, question_fa: e.target.value })
                }
                placeholder="Optional"
              />
            </div>

            <div className="space-y-2">
              <Label>Question (Arabic)</Label>
              <Textarea
                value={createForm.question_ar}
                onChange={(e) =>
                  setCreateForm({ ...createForm, question_ar: e.target.value })
                }
                placeholder="Optional"
              />
            </div>

            <div className="space-y-2">
              <Label>Tier</Label>
              <Select
                value={createForm.tier}
                onValueChange={(val) =>
                  setCreateForm({ ...createForm, tier: val ?? "white" })
                }
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="white">White</SelectItem>
                  <SelectItem value="silver">Silver</SelectItem>
                  <SelectItem value="gold">Gold</SelectItem>
                  <SelectItem value="vip">VIP</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="flex gap-4">
              <div className="flex-1 space-y-2">
                <Label>Available Date</Label>
                <Input
                  type="date"
                  value={createForm.available_date}
                  onChange={(e) =>
                    setCreateForm({
                      ...createForm,
                      available_date: e.target.value,
                    })
                  }
                />
              </div>
              <div className="flex-1 space-y-2">
                <Label>Expires At</Label>
                <Input
                  type="date"
                  value={createForm.expires_at}
                  onChange={(e) =>
                    setCreateForm({
                      ...createForm,
                      expires_at: e.target.value,
                    })
                  }
                />
              </div>
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
              onClick={handleCreateCard}
              disabled={
                createMutation.isPending ||
                !createForm.match_id ||
                !createForm.question_en ||
                !createForm.available_date ||
                !createForm.expires_at
              }
            >
              {createMutation.isPending ? "Creating..." : "Create"}
            </Button>
          </DialogFooter>
          {createMutation.isError && (
            <p className="text-sm text-destructive">
              Failed to create card. Please try again.
            </p>
          )}
        </DialogContent>
      </Dialog>

      {/* Edit Card Dialog */}
      <Dialog
        open={editCard !== null}
        onOpenChange={(open) => {
          if (!open) setEditCard(null);
        }}
      >
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Edit Card</DialogTitle>
            <DialogDescription>
              Update the card details. Resolved cards cannot be edited.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label>Question (English)</Label>
              <Textarea
                value={editForm.question_en}
                onChange={(e) =>
                  setEditForm({ ...editForm, question_en: e.target.value })
                }
                placeholder="Will Team A win?"
              />
            </div>

            <div className="space-y-2">
              <Label>Tier</Label>
              <Select
                value={editForm.tier}
                onValueChange={(val) =>
                  setEditForm({ ...editForm, tier: val ?? "white" })
                }
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="white">White</SelectItem>
                  <SelectItem value="silver">Silver</SelectItem>
                  <SelectItem value="gold">Gold</SelectItem>
                  <SelectItem value="vip">VIP</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="flex gap-4">
              <div className="flex-1 space-y-2">
                <Label>Available Date</Label>
                <Input
                  type="date"
                  value={editForm.available_date}
                  onChange={(e) =>
                    setEditForm({
                      ...editForm,
                      available_date: e.target.value,
                    })
                  }
                />
              </div>
              <div className="flex-1 space-y-2">
                <Label>Expires At</Label>
                <Input
                  type="date"
                  value={editForm.expires_at}
                  onChange={(e) =>
                    setEditForm({
                      ...editForm,
                      expires_at: e.target.value,
                    })
                  }
                />
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setEditCard(null)}>
              Cancel
            </Button>
            <Button
              onClick={handleEditCard}
              disabled={
                editMutation.isPending ||
                !editForm.question_en ||
                !editForm.available_date ||
                !editForm.expires_at
              }
            >
              {editMutation.isPending ? "Saving..." : "Save Changes"}
            </Button>
          </DialogFooter>
          {editMutation.isError && (
            <p className="text-sm text-destructive">
              Failed to update card. Please try again.
            </p>
          )}
        </DialogContent>
      </Dialog>

      {/* Delete Card Dialog */}
      <DeleteDialog
        open={deleteCard !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteCard(null);
        }}
        title="Delete Card"
        description={
          deleteCard
            ? `Are you sure you want to delete the card "${truncate(getQuestionText(deleteCard.question_text), 60)}"? This action cannot be undone.`
            : ""
        }
        onConfirm={handleDeleteCard}
        isPending={deleteMutation.isPending}
        isError={deleteMutation.isError}
      />
    </div>
  );
}
