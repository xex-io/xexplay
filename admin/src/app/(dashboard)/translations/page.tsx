"use client";

import { useState, useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface CardItem {
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

function statusVariant(
  presentCount: number
): "default" | "secondary" | "destructive" {
  if (presentCount === SUPPORTED_LANGUAGES.length) return "default";
  if (presentCount >= 2) return "secondary";
  return "destructive";
}

function truncate(text: string, len: number): string {
  return text && text.length > len ? text.slice(0, len) + "..." : text || "-";
}

export default function TranslationsPage() {
  const [filterMissing, setFilterMissing] = useState(false);
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");

  const { data: cards = [], isLoading } = useQuery<CardItem[]>({
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
        <h1 className="text-2xl font-bold text-foreground">Translations</h1>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        <Card>
          <CardHeader>
            <CardTitle className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
              Total Cards
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-foreground">{stats.total}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
              Complete Translations
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-green-400">
              {stats.complete}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
              Incomplete Translations
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-yellow-400">
              {stats.incomplete}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap items-end gap-4 mb-4">
        <div>
          <Label className="mb-1 text-xs uppercase tracking-wider text-muted-foreground">
            Date From
          </Label>
          <Input
            type="date"
            value={dateFrom}
            onChange={(e) => setDateFrom(e.target.value)}
          />
        </div>
        <div>
          <Label className="mb-1 text-xs uppercase tracking-wider text-muted-foreground">
            Date To
          </Label>
          <Input
            type="date"
            value={dateTo}
            onChange={(e) => setDateTo(e.target.value)}
          />
        </div>
        <Label className="cursor-pointer pb-1">
          <input
            type="checkbox"
            checked={filterMissing}
            onChange={(e) => setFilterMissing(e.target.checked)}
            className="rounded border-border bg-background text-primary focus:ring-ring"
          />
          <span className="text-sm text-muted-foreground">
            Show only missing translations
          </span>
        </Label>
      </div>

      {/* Table */}
      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
              <TableHead>Question (EN)</TableHead>
              <TableHead>Question (FA)</TableHead>
              <TableHead>Question (AR)</TableHead>
              <TableHead>Missing Languages</TableHead>
              <TableHead>Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center py-12 text-muted-foreground">
                  Loading cards...
                </TableCell>
              </TableRow>
            ) : filtered.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center py-12 text-muted-foreground">
                  No cards found.
                </TableCell>
              </TableRow>
            ) : (
              filtered.map((card) => {
                const { present, missing } = getTranslationStatus(
                  card.question_text
                );
                return (
                  <TableRow key={card.id}>
                    <TableCell className="font-mono text-muted-foreground">
                      {card.id.slice(0, 8)}
                    </TableCell>
                    <TableCell className="max-w-[200px]">
                      {truncate(card.question_text?.en, 40)}
                    </TableCell>
                    <TableCell className="max-w-[200px]" dir="rtl">
                      {truncate(card.question_text?.fa, 40)}
                    </TableCell>
                    <TableCell className="max-w-[200px]" dir="rtl">
                      {truncate(card.question_text?.ar, 40)}
                    </TableCell>
                    <TableCell>
                      {missing.length === 0 ? (
                        <span className="text-green-400">None</span>
                      ) : (
                        <div className="flex gap-1">
                          {missing.map((lang) => (
                            <Badge key={lang} variant="destructive">
                              {lang.toUpperCase()}
                            </Badge>
                          ))}
                        </div>
                      )}
                    </TableCell>
                    <TableCell>
                      <Badge variant={statusVariant(present.length)}>
                        {present.length}/{SUPPORTED_LANGUAGES.length}
                      </Badge>
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>
      </Card>
    </div>
  );
}
