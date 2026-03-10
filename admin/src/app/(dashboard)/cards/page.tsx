"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";

interface Card {
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

const tierColors: Record<string, string> = {
  gold: "bg-yellow-500/20 text-yellow-400 border border-yellow-500/30",
  silver: "bg-gray-400/20 text-gray-300 border border-gray-400/30",
  white: "bg-white/10 text-white border border-white/20",
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
  const [resolveModal, setResolveModal] = useState<Card | null>(null);
  const [selectedAnswer, setSelectedAnswer] = useState<boolean | null>(null);
  const [confirmStep, setConfirmStep] = useState(false);

  const { data: cards = [], isLoading } = useQuery<Card[]>({
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

  function openModal(card: Card) {
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
        <h1 className="text-2xl font-bold text-gray-100">Cards</h1>
        <button className="bg-blue-600 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-blue-700 transition-colors">
          Create Card
        </button>
      </div>

      <div className="bg-gray-800 shadow rounded-lg border border-gray-700 overflow-hidden">
        <table className="min-w-full divide-y divide-gray-700">
          <thead className="bg-gray-900">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                ID
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Question
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Tier
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Match
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Available Date
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Resolved
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-700">
            {isLoading ? (
              <tr>
                <td
                  colSpan={7}
                  className="px-6 py-12 text-center text-sm text-gray-400"
                >
                  Loading cards...
                </td>
              </tr>
            ) : cards.length === 0 ? (
              <tr>
                <td
                  colSpan={7}
                  className="px-6 py-12 text-center text-sm text-gray-400"
                >
                  No cards found.
                </td>
              </tr>
            ) : (
              cards.map((card) => (
                <tr
                  key={card.id}
                  className="hover:bg-gray-750 transition-colors"
                >
                  <td className="px-6 py-4 text-sm text-gray-300 font-mono">
                    {card.id.slice(0, 8)}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-200 max-w-xs">
                    {truncate(getQuestionText(card.question_text), 50)}
                  </td>
                  <td className="px-6 py-4">
                    <span
                      className={`inline-flex px-2.5 py-0.5 rounded-full text-xs font-semibold capitalize ${tierColors[card.tier] || tierColors.white}`}
                    >
                      {card.tier}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-300">
                    {getMatchLabel(card.match_id)}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-300">
                    {new Date(card.available_date).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4 text-sm">
                    {card.is_resolved ? (
                      <span className="inline-flex items-center gap-1.5 text-green-400">
                        <svg
                          className="w-4 h-4"
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke="currentColor"
                          strokeWidth={2}
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            d="M5 13l4 4L19 7"
                          />
                        </svg>
                        {card.correct_answer === true
                          ? "Yes"
                          : card.correct_answer === false
                            ? "No"
                            : "-"}
                      </span>
                    ) : (
                      <span className="text-gray-500">Pending</span>
                    )}
                  </td>
                  <td className="px-6 py-4 text-sm">
                    {!card.is_resolved && (
                      <button
                        onClick={() => openModal(card)}
                        className="bg-indigo-600 hover:bg-indigo-500 text-white px-3 py-1.5 rounded-md text-xs font-medium transition-colors"
                      >
                        Resolve
                      </button>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Resolution Modal */}
      {resolveModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div
            className="fixed inset-0 bg-black/60 backdrop-blur-sm"
            onClick={closeModal}
          />
          <div className="relative bg-gray-800 border border-gray-700 rounded-xl shadow-2xl w-full max-w-lg mx-4 p-6">
            <button
              onClick={closeModal}
              className="absolute top-4 right-4 text-gray-400 hover:text-gray-200 transition-colors"
            >
              <svg
                className="w-5 h-5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>

            <h2 className="text-lg font-semibold text-gray-100 mb-4">
              Resolve Card
            </h2>

            <div className="space-y-3 mb-6">
              <div>
                <label className="text-xs font-medium text-gray-400 uppercase tracking-wider">
                  Question
                </label>
                <p className="mt-1 text-sm text-gray-200 leading-relaxed">
                  {getQuestionText(resolveModal.question_text)}
                </p>
              </div>

              <div className="flex gap-4">
                <div>
                  <label className="text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Tier
                  </label>
                  <p className="mt-1">
                    <span
                      className={`inline-flex px-2.5 py-0.5 rounded-full text-xs font-semibold capitalize ${tierColors[resolveModal.tier] || tierColors.white}`}
                    >
                      {resolveModal.tier}
                    </span>
                  </p>
                </div>
                <div>
                  <label className="text-xs font-medium text-gray-400 uppercase tracking-wider">
                    Match
                  </label>
                  <p className="mt-1 text-sm text-gray-200">
                    {getMatchLabel(resolveModal.match_id)}
                  </p>
                </div>
              </div>
            </div>

            {!confirmStep ? (
              <div>
                <p className="text-sm text-gray-400 mb-3">
                  What is the correct answer?
                </p>
                <div className="flex gap-3">
                  <button
                    onClick={() => handleSelectAnswer(true)}
                    className="flex-1 bg-green-600/20 border border-green-500/30 text-green-400 hover:bg-green-600/30 hover:border-green-500/50 px-4 py-3 rounded-lg text-sm font-semibold transition-colors"
                  >
                    Yes
                  </button>
                  <button
                    onClick={() => handleSelectAnswer(false)}
                    className="flex-1 bg-red-600/20 border border-red-500/30 text-red-400 hover:bg-red-600/30 hover:border-red-500/50 px-4 py-3 rounded-lg text-sm font-semibold transition-colors"
                  >
                    No
                  </button>
                </div>
              </div>
            ) : (
              <div>
                <div className="bg-gray-900 border border-gray-600 rounded-lg p-4 mb-4">
                  <p className="text-sm text-gray-300">
                    Are you sure the correct answer is{" "}
                    <span
                      className={`font-bold ${selectedAnswer ? "text-green-400" : "text-red-400"}`}
                    >
                      {selectedAnswer ? "YES" : "NO"}
                    </span>
                    ?
                  </p>
                </div>
                <div className="flex gap-3">
                  <button
                    onClick={() => {
                      setConfirmStep(false);
                      setSelectedAnswer(null);
                    }}
                    className="flex-1 bg-gray-700 hover:bg-gray-600 text-gray-300 px-4 py-2.5 rounded-lg text-sm font-medium transition-colors"
                  >
                    Go Back
                  </button>
                  <button
                    onClick={handleConfirmResolve}
                    disabled={resolveMutation.isPending}
                    className="flex-1 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed text-white px-4 py-2.5 rounded-lg text-sm font-semibold transition-colors"
                  >
                    {resolveMutation.isPending ? "Resolving..." : "Confirm"}
                  </button>
                </div>
                {resolveMutation.isError && (
                  <p className="mt-3 text-sm text-red-400">
                    Failed to resolve card. Please try again.
                  </p>
                )}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
