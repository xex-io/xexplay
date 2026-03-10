"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
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
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Check, Plus, AlertCircle } from "lucide-react";

interface CardItem {
  id: string;
  match_id: string;
  question_text: Record<string, string>;
  tier: "gold" | "silver" | "white";
  high_answer_is_yes: boolean | null;
  correct_answer: boolean | null;
  is_resolved: boolean;
  available_date: string;
  expires_at: string;
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
};

function getQuestionText(qt: Record<string, string>): string {
  if (typeof qt === "string") return qt;
  return qt?.en || qt?.fa || Object.values(qt || {})[0] || "";
}

function truncate(text: string, len: number): string {
  return text.length > len ? text.slice(0, len) + "..." : text;
}

export default function CardsPage() {
  const queryClient = useQueryClient();
  const [resolveModal, setResolveModal] = useState<CardItem | null>(null);
  const [selectedAnswer, setSelectedAnswer] = useState<boolean | null>(null);
  const [confirmStep, setConfirmStep] = useState(false);

  const { data: cards = [], isLoading } = useQuery<CardItem[]>({
    queryKey: ["admin-cards"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/cards");
      return res.data?.data ?? res.data ?? [];
    },
  });

  const { data: matches = [] } = useQuery<Match[]>({
    queryKey: ["admin-matches"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/matches");
      return res.data?.data ?? res.data ?? [];
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

  function getMatchLabel(matchId: string): string {
    const m = matchMap.get(matchId);
    return m ? `${m.home_team} vs ${m.away_team}` : matchId.slice(0, 8);
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-foreground">Cards</h1>
        <Button size="sm">
          <Plus className="size-4" data-icon="inline-start" />
          Create Card
        </Button>
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
                    {!card.is_resolved && (
                      <Button
                        size="xs"
                        variant="secondary"
                        onClick={() => openModal(card)}
                      >
                        Resolve
                      </Button>
                    )}
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
    </div>
  );
}
