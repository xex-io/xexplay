"use client";

import { useState, useMemo } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { asArray, locStr, type LocalizedString } from "@/lib/loc-str";
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
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogClose,
} from "@/components/ui/dialog";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select";
import { Pencil, Loader2, Check, AlertCircle, Globe } from "lucide-react";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";

// ---- Types ----

interface CardItem {
  id: string;
  question_text: Record<string, string>;
  tier: string;
  is_resolved: boolean;
  available_date: string;
  created_at: string;
}

interface EventItem {
  id: string;
  name: LocalizedString;
  description: LocalizedString;
  slug: string;
  is_active: boolean;
  start_date: string;
}

interface TranslatableItem {
  id: string;
  type: "card" | "event_name" | "event_description";
  label: string;
  texts: Record<string, string>;
  editable: boolean;
  parentId?: string; // for event fields, the event id
}

// ---- Language Config ----

const LANGUAGES = [
  { code: "en", label: "English", dir: "ltr" },
  { code: "fa", label: "Persian", dir: "rtl" },
  { code: "ar", label: "Arabic", dir: "rtl" },
  { code: "tr", label: "Turkish", dir: "ltr" },
  { code: "ru", label: "Russian", dir: "ltr" },
  { code: "zh", label: "Chinese", dir: "ltr" },
  { code: "es", label: "Spanish", dir: "ltr" },
  { code: "pt", label: "Portuguese", dir: "ltr" },
] as const;

type LangCode = (typeof LANGUAGES)[number]["code"];

const LANG_MAP = Object.fromEntries(LANGUAGES.map((l) => [l.code, l]));

function getTranslationStatus(
  texts: Record<string, string>,
  langs: LangCode[]
): { present: LangCode[]; missing: LangCode[] } {
  const present = langs.filter(
    (lang) => texts?.[lang] && texts[lang].trim().length > 0
  );
  const missing = langs.filter(
    (lang) => !texts?.[lang] || texts[lang].trim().length === 0
  );
  return { present, missing };
}

function statusVariant(
  presentCount: number,
  totalCount: number
): "default" | "secondary" | "destructive" {
  if (presentCount === totalCount) return "default";
  if (presentCount >= Math.ceil(totalCount / 2)) return "secondary";
  return "destructive";
}

function truncate(text: string, len: number): string {
  return text && text.length > len ? text.slice(0, len) + "..." : text || "-";
}

function parseLocalized(val: LocalizedString | undefined | null): Record<string, string> {
  if (!val) return {};
  if (typeof val === "string") return { en: val };
  return val as Record<string, string>;
}

// ---- Component ----

export default function TranslationsPage() {
  const queryClient = useQueryClient();

  // Filters
  const [filterMissing, setFilterMissing] = useState(false);
  const [filterType, setFilterType] = useState<"all" | "card" | "event">("all");
  const [selectedLangs, setSelectedLangs] = useState<LangCode[]>(["en", "fa", "ar"]);
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");

  // Edit state
  const [editItem, setEditItem] = useState<TranslatableItem | null>(null);
  const [editTexts, setEditTexts] = useState<Record<string, string>>({});

  // Data queries
  const { data: cards = [], isLoading: cardsLoading } = useQuery<CardItem[]>({
    queryKey: ["admin-cards"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/cards");
      return asArray<CardItem>(res);
    },
  });

  const { data: events = [], isLoading: eventsLoading } = useQuery<EventItem[]>({
    queryKey: ["admin", "events"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/events");
      return asArray<EventItem>(res);
    },
  });

  const isLoading = cardsLoading || eventsLoading;

  // Build unified translatable items
  const items: TranslatableItem[] = useMemo(() => {
    const result: TranslatableItem[] = [];

    if (filterType === "all" || filterType === "event") {
      events.forEach((event) => {
        result.push({
          id: `event_name_${event.id}`,
          type: "event_name",
          label: `Event: ${locStr(event.name) || event.slug}`,
          texts: parseLocalized(event.name),
          editable: true,
          parentId: event.id,
        });
        result.push({
          id: `event_desc_${event.id}`,
          type: "event_description",
          label: `Event Desc: ${locStr(event.name) || event.slug}`,
          texts: parseLocalized(event.description),
          editable: true,
          parentId: event.id,
        });
      });
    }

    if (filterType === "all" || filterType === "card") {
      cards.forEach((card) => {
        result.push({
          id: `card_${card.id}`,
          type: "card",
          label: truncate(card.question_text?.en || `Card ${card.id.slice(0, 8)}`, 50),
          texts: card.question_text ?? {},
          editable: !card.is_resolved,
          parentId: card.id,
        });
      });
    }

    return result;
  }, [cards, events, filterType]);

  // Apply filters
  const filtered = useMemo(() => {
    return items.filter((item) => {
      if (filterMissing) {
        const { missing } = getTranslationStatus(item.texts, selectedLangs);
        if (missing.length === 0) return false;
      }
      if (item.type === "card" && (dateFrom || dateTo)) {
        const card = cards.find((c) => `card_${c.id}` === item.id);
        if (card) {
          const d = new Date(card.available_date || card.created_at);
          if (dateFrom && d < new Date(dateFrom)) return false;
          if (dateTo && d > new Date(dateTo + "T23:59:59")) return false;
        }
      }
      return true;
    });
  }, [items, filterMissing, selectedLangs, dateFrom, dateTo, cards]);

  // Stats
  const stats = useMemo(() => {
    let complete = 0;
    let incomplete = 0;
    items.forEach((item) => {
      const { missing } = getTranslationStatus(item.texts, selectedLangs);
      if (missing.length === 0) complete++;
      else incomplete++;
    });
    return { complete, incomplete, total: items.length };
  }, [items, selectedLangs]);

  // Save mutation
  const saveMutation = useMutation({
    mutationFn: async (data: { item: TranslatableItem; texts: Record<string, string> }) => {
      const { item, texts } = data;
      if (item.type === "card") {
        return apiClient.put(`/admin/cards/${item.parentId}`, {
          question_text: texts,
        });
      } else if (item.type === "event_name") {
        return apiClient.put(`/admin/events/${item.parentId}`, {
          name: texts,
        });
      } else {
        return apiClient.put(`/admin/events/${item.parentId}`, {
          description: texts,
        });
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-cards"] });
      queryClient.invalidateQueries({ queryKey: ["admin", "events"] });
      setEditItem(null);
    },
  });

  function openEdit(item: TranslatableItem) {
    setEditTexts({ ...item.texts });
    setEditItem(item);
    saveMutation.reset();
  }

  function handleSave() {
    if (!editItem) return;
    saveMutation.mutate({ item: editItem, texts: editTexts });
  }

  function toggleLang(lang: LangCode) {
    setSelectedLangs((prev) =>
      prev.includes(lang)
        ? prev.length > 1
          ? prev.filter((l) => l !== lang)
          : prev
        : [...prev, lang]
    );
  }

  const completionPct =
    stats.total > 0 ? ((stats.complete / stats.total) * 100).toFixed(0) : "0";

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-foreground">Translations</h1>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <Card>
          <CardHeader>
            <CardTitle className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
              Total Items
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-foreground">{stats.total}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
              Fully Translated
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-green-400">{stats.complete}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
              Missing Translations
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-yellow-400">{stats.incomplete}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
              Completion Rate
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-foreground">{completionPct}%</p>
            <div className="mt-2 w-full bg-muted rounded-full h-2">
              <div
                className="bg-primary h-2 rounded-full transition-all"
                style={{ width: `${Math.min(Number(completionPct), 100)}%` }}
              />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Language Selector */}
      <Card className="mb-4">
        <CardContent className="pt-4">
          <div className="flex items-center gap-3 flex-wrap">
            <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
              <Globe className="size-4" />
              Languages:
            </div>
            {LANGUAGES.map((lang) => (
              <Button
                key={lang.code}
                size="sm"
                variant={selectedLangs.includes(lang.code) ? "default" : "outline"}
                onClick={() => toggleLang(lang.code)}
                className="text-xs"
              >
                {lang.label} ({lang.code.toUpperCase()})
              </Button>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Filters */}
      <div className="flex flex-wrap items-end gap-4 mb-4">
        <div>
          <Label className="mb-1 text-xs uppercase tracking-wider text-muted-foreground">
            Type
          </Label>
          <Select value={filterType} onValueChange={(v) => setFilterType(v as typeof filterType)}>
            <SelectTrigger className="w-40">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All</SelectItem>
              <SelectItem value="card">Cards Only</SelectItem>
              <SelectItem value="event">Events Only</SelectItem>
            </SelectContent>
          </Select>
        </div>
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
        <Label className="cursor-pointer pb-1 flex items-center gap-1.5">
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
              <TableHead className="w-16">Type</TableHead>
              <TableHead>Item</TableHead>
              {selectedLangs.map((lang) => (
                <TableHead key={lang} className="min-w-[150px]">
                  {LANG_MAP[lang]?.label ?? lang.toUpperCase()}
                </TableHead>
              ))}
              <TableHead className="w-24">Status</TableHead>
              <TableHead className="w-16">Edit</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell
                  colSpan={selectedLangs.length + 4}
                  className="text-center py-12 text-muted-foreground"
                >
                  <Loader2 className="size-5 animate-spin inline-block mr-2" />
                  Loading...
                </TableCell>
              </TableRow>
            ) : filtered.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={selectedLangs.length + 4}
                  className="text-center py-12 text-muted-foreground"
                >
                  No items found.
                </TableCell>
              </TableRow>
            ) : (
              filtered.map((item) => {
                const { present, missing } = getTranslationStatus(
                  item.texts,
                  selectedLangs
                );
                return (
                  <TableRow key={item.id}>
                    <TableCell>
                      <Badge
                        variant={item.type === "card" ? "outline" : "secondary"}
                        className="text-xs"
                      >
                        {item.type === "card"
                          ? "Card"
                          : item.type === "event_name"
                          ? "Event"
                          : "Desc"}
                      </Badge>
                    </TableCell>
                    <TableCell className="max-w-[200px] truncate font-medium">
                      {item.type === "card"
                        ? truncate(item.texts?.en || item.parentId?.slice(0, 8) || "-", 40)
                        : item.label}
                    </TableCell>
                    {selectedLangs.map((lang) => {
                      const text = item.texts?.[lang];
                      const isRtl = LANG_MAP[lang]?.dir === "rtl";
                      return (
                        <TableCell
                          key={lang}
                          dir={isRtl ? "rtl" : undefined}
                          className="max-w-[200px]"
                        >
                          {text ? (
                            <span className="text-sm">{truncate(text, 40)}</span>
                          ) : (
                            <span className="text-muted-foreground/50 text-xs italic">
                              Missing
                            </span>
                          )}
                        </TableCell>
                      );
                    })}
                    <TableCell>
                      <Badge
                        variant={statusVariant(present.length, selectedLangs.length)}
                      >
                        {present.length}/{selectedLangs.length}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Button
                        size="icon-sm"
                        variant="ghost"
                        onClick={() => openEdit(item)}
                        disabled={!item.editable}
                        title={item.editable ? "Edit translations" : "Cannot edit resolved card"}
                      >
                        <Pencil className="size-3.5" />
                      </Button>
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>
      </Card>

      {/* Edit Dialog */}
      <Dialog
        open={editItem !== null}
        onOpenChange={(open) => {
          if (!open) setEditItem(null);
        }}
      >
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Edit Translations</DialogTitle>
          </DialogHeader>
          {editItem && (
            <div className="space-y-4">
              <div className="rounded-lg border border-border bg-muted/50 p-3">
                <p className="text-xs text-muted-foreground uppercase tracking-wider mb-1">
                  {editItem.type === "card"
                    ? "Card Question"
                    : editItem.type === "event_name"
                    ? "Event Name"
                    : "Event Description"}
                </p>
                <p className="text-sm font-medium text-foreground">
                  {editItem.texts?.en || editItem.label}
                </p>
              </div>

              <div className="space-y-3 max-h-[400px] overflow-y-auto">
                {LANGUAGES.map((lang) => {
                  const isRtl = lang.dir === "rtl";
                  const hasValue = !!editTexts[lang.code]?.trim();
                  return (
                    <div key={lang.code} className="space-y-1">
                      <div className="flex items-center justify-between">
                        <Label className="text-xs uppercase tracking-wider text-muted-foreground">
                          {lang.label} ({lang.code.toUpperCase()})
                        </Label>
                        {hasValue && (
                          <Check className="size-3.5 text-green-400" />
                        )}
                      </div>
                      {editItem.type === "event_description" ? (
                        <Textarea
                          dir={isRtl ? "rtl" : "ltr"}
                          value={editTexts[lang.code] || ""}
                          onChange={(e) =>
                            setEditTexts((prev) => ({
                              ...prev,
                              [lang.code]: e.target.value,
                            }))
                          }
                          rows={2}
                          placeholder={`${lang.label} translation...`}
                          className={isRtl ? "text-right" : ""}
                        />
                      ) : (
                        <Input
                          dir={isRtl ? "rtl" : "ltr"}
                          value={editTexts[lang.code] || ""}
                          onChange={(e) =>
                            setEditTexts((prev) => ({
                              ...prev,
                              [lang.code]: e.target.value,
                            }))
                          }
                          placeholder={`${lang.label} translation...`}
                          className={isRtl ? "text-right" : ""}
                        />
                      )}
                    </div>
                  );
                })}
              </div>

              {saveMutation.isError && (
                <Alert variant="destructive">
                  <AlertCircle className="size-4" />
                  <AlertTitle>Error</AlertTitle>
                  <AlertDescription>
                    Failed to save translations. Please try again.
                  </AlertDescription>
                </Alert>
              )}
            </div>
          )}
          <DialogFooter>
            <DialogClose render={<Button variant="outline" />}>Cancel</DialogClose>
            <Button
              onClick={handleSave}
              disabled={saveMutation.isPending}
            >
              {saveMutation.isPending && (
                <Loader2 className="size-4 mr-1 animate-spin" />
              )}
              Save Translations
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
