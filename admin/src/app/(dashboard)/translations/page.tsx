"use client";

import { useState, useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";

interface Card {
  id: string;
  question_text: Record<string, string>;
  tier: "gold" | "silver" | "white";
  available_date: string;
  created_at: string;
}

const SUPPORTED_LANGUAGES = ["en", "fa", "ar"];

function getTranslationStatus(qt: Record<string, string>): {
  present: string[];
  missing: string[];
} {
  const present = SUPPORTED_LANGUAGES.filter(
    (lang) => qt?.[lang] && qt[lang].trim().length > 0
  );
  const missing = SUPPORTED_LANGUAGES.filter(
    (lang) => !qt?.[lang] || qt[lang].trim().length === 0
  );
  return { present, missing };
}

function statusColor(presentCount: number): string {
  if (presentCount === SUPPORTED_LANGUAGES.length)
    return "bg-green-500/20 text-green-400 border border-green-500/30";
  if (presentCount >= 2)
    return "bg-yellow-500/20 text-yellow-400 border border-yellow-500/30";
  return "bg-red-500/20 text-red-400 border border-red-500/30";
}

function truncate(text: string, len: number): string {
  return text && text.length > len ? text.slice(0, len) + "..." : text || "-";
}

export default function TranslationsPage() {
  const [filterMissing, setFilterMissing] = useState(false);
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");

  const { data: cards = [], isLoading } = useQuery<Card[]>({
    queryKey: ["admin-cards"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/cards");
      return res.data?.data ?? res.data ?? [];
    },
  });

  const filtered = useMemo(() => {
    return cards.filter((card) => {
      if (filterMissing) {
        const { missing } = getTranslationStatus(card.question_text);
        if (missing.length === 0) return false;
      }
      if (dateFrom) {
        const d = new Date(card.available_date || card.created_at);
        if (d < new Date(dateFrom)) return false;
      }
      if (dateTo) {
        const d = new Date(card.available_date || card.created_at);
        if (d > new Date(dateTo + "T23:59:59")) return false;
      }
      return true;
    });
  }, [cards, filterMissing, dateFrom, dateTo]);

  const stats = useMemo(() => {
    let complete = 0;
    let incomplete = 0;
    cards.forEach((card) => {
      const { missing } = getTranslationStatus(card.question_text);
      if (missing.length === 0) complete++;
      else incomplete++;
    });
    return { complete, incomplete, total: cards.length };
  }, [cards]);

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-100">Translations</h1>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        <div className="bg-gray-800 border border-gray-700 rounded-lg p-4">
          <p className="text-xs font-medium text-gray-400 uppercase tracking-wider">Total Cards</p>
          <p className="mt-1 text-2xl font-bold text-gray-100">{stats.total}</p>
        </div>
        <div className="bg-gray-800 border border-gray-700 rounded-lg p-4">
          <p className="text-xs font-medium text-gray-400 uppercase tracking-wider">Complete Translations</p>
          <p className="mt-1 text-2xl font-bold text-green-400">{stats.complete}</p>
        </div>
        <div className="bg-gray-800 border border-gray-700 rounded-lg p-4">
          <p className="text-xs font-medium text-gray-400 uppercase tracking-wider">Incomplete Translations</p>
          <p className="mt-1 text-2xl font-bold text-yellow-400">{stats.incomplete}</p>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap items-end gap-4 mb-4">
        <div>
          <label className="block text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">
            Date From
          </label>
          <input
            type="date"
            value={dateFrom}
            onChange={(e) => setDateFrom(e.target.value)}
            className="bg-gray-900 border border-gray-700 text-gray-200 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">
            Date To
          </label>
          <input
            type="date"
            value={dateTo}
            onChange={(e) => setDateTo(e.target.value)}
            className="bg-gray-900 border border-gray-700 text-gray-200 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <label className="flex items-center gap-2 cursor-pointer pb-1">
          <input
            type="checkbox"
            checked={filterMissing}
            onChange={(e) => setFilterMissing(e.target.checked)}
            className="rounded bg-gray-900 border-gray-700 text-blue-600 focus:ring-blue-500"
          />
          <span className="text-sm text-gray-300">Show only missing translations</span>
        </label>
      </div>

      {/* Table */}
      <div className="bg-gray-800 shadow rounded-lg border border-gray-700 overflow-hidden">
        <table className="min-w-full divide-y divide-gray-700">
          <thead className="bg-gray-900">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                ID
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Question (EN)
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Question (FA)
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Question (AR)
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Missing Languages
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Status
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-700">
            {isLoading ? (
              <tr>
                <td colSpan={6} className="px-6 py-12 text-center text-sm text-gray-400">
                  Loading cards...
                </td>
              </tr>
            ) : filtered.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-6 py-12 text-center text-sm text-gray-400">
                  No cards found.
                </td>
              </tr>
            ) : (
              filtered.map((card) => {
                const { present, missing } = getTranslationStatus(card.question_text);
                return (
                  <tr key={card.id} className="hover:bg-gray-750 transition-colors">
                    <td className="px-6 py-4 text-sm text-gray-300 font-mono">
                      {card.id.slice(0, 8)}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-200 max-w-[200px]">
                      {truncate(card.question_text?.en, 40)}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-200 max-w-[200px]" dir="rtl">
                      {truncate(card.question_text?.fa, 40)}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-200 max-w-[200px]" dir="rtl">
                      {truncate(card.question_text?.ar, 40)}
                    </td>
                    <td className="px-6 py-4 text-sm">
                      {missing.length === 0 ? (
                        <span className="text-green-400">None</span>
                      ) : (
                        <div className="flex gap-1">
                          {missing.map((lang) => (
                            <span
                              key={lang}
                              className="inline-flex px-2 py-0.5 rounded text-xs font-semibold bg-red-500/20 text-red-400 border border-red-500/30 uppercase"
                            >
                              {lang}
                            </span>
                          ))}
                        </div>
                      )}
                    </td>
                    <td className="px-6 py-4 text-sm">
                      <span
                        className={`inline-flex px-2.5 py-0.5 rounded-full text-xs font-semibold ${statusColor(present.length)}`}
                      >
                        {present.length}/{SUPPORTED_LANGUAGES.length}
                      </span>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
