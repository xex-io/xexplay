"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { Send, AlertCircle, CheckCircle2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  DialogClose,
} from "@/components/ui/dialog";
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "@/components/ui/table";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { Separator } from "@/components/ui/separator";

interface NotificationHistory {
  id: string;
  title: string;
  body: string;
  target: string;
  sent_at: string;
  sent_by: string;
}

const TARGET_OPTIONS = [
  { value: "all", label: "All Users" },
  { value: "active", label: "Active Users (last 7 days)" },
  { value: "new", label: "New Users (last 24h)" },
  { value: "dormant", label: "Dormant Users (inactive 30+ days)" },
];

export default function NotificationsPage() {
  const queryClient = useQueryClient();
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [target, setTarget] = useState("all");
  const [dialogOpen, setDialogOpen] = useState(false);

  const { data: history = [], isLoading } = useQuery<NotificationHistory[]>({
    queryKey: ["admin-notifications"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/notifications");
      return res.data?.data ?? res.data ?? [];
    },
  });

  const sendMutation = useMutation({
    mutationFn: async (data: { title: string; body: string; target: string }) => {
      return apiClient.post("/admin/notifications/send", data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-notifications"] });
      setTitle("");
      setBody("");
      setTarget("all");
      setDialogOpen(false);
    },
  });

  function handleSend() {
    if (!title.trim() || !body.trim()) return;
    setDialogOpen(true);
  }

  function confirmSend() {
    sendMutation.mutate({ title, body, target });
  }

  const targetLabel = TARGET_OPTIONS.find((o) => o.value === target)?.label ?? target;

  return (
    <div>
      <h1 className="text-2xl font-bold text-foreground mb-6">Notifications</h1>

      {/* Compose Form */}
      <Card className="mb-8">
        <CardHeader>
          <CardTitle>Compose Notification</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="notif-title">Title</Label>
              <Input
                id="notif-title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Notification title"
              />
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="notif-body">Body</Label>
              <Textarea
                id="notif-body"
                value={body}
                onChange={(e) => setBody(e.target.value)}
                placeholder="Notification body"
                rows={4}
              />
            </div>

            <div className="space-y-1.5">
              <Label>Target</Label>
              <Select value={target} onValueChange={(v) => setTarget(v ?? "")}>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Select target" />
                </SelectTrigger>
                <SelectContent>
                  {TARGET_OPTIONS.map((opt) => (
                    <SelectItem key={opt.value} value={opt.value}>
                      {opt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="flex justify-end">
              <Button
                onClick={handleSend}
                disabled={!title.trim() || !body.trim() || sendMutation.isPending}
              >
                <Send className="size-4 mr-1.5" />
                Send Notification
              </Button>
            </div>

            {sendMutation.isError && (
              <Alert variant="destructive">
                <AlertCircle className="size-4" />
                <AlertTitle>Error</AlertTitle>
                <AlertDescription>
                  Failed to send notification. Please try again.
                </AlertDescription>
              </Alert>
            )}
            {sendMutation.isSuccess && (
              <Alert>
                <CheckCircle2 className="size-4" />
                <AlertTitle>Success</AlertTitle>
                <AlertDescription>
                  Notification sent successfully.
                </AlertDescription>
              </Alert>
            )}
          </div>
        </CardContent>
      </Card>

      {/* History Table */}
      <Card>
        <CardHeader>
          <CardTitle>History</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Title</TableHead>
                <TableHead>Body</TableHead>
                <TableHead>Target</TableHead>
                <TableHead>Sent By</TableHead>
                <TableHead>Sent At</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-center py-12 text-muted-foreground">
                    Loading history...
                  </TableCell>
                </TableRow>
              ) : history.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-center py-12 text-muted-foreground">
                    No notifications sent yet.
                  </TableCell>
                </TableRow>
              ) : (
                history.map((n) => (
                  <TableRow key={n.id}>
                    <TableCell className="font-medium">{n.title}</TableCell>
                    <TableCell className="max-w-xs truncate text-muted-foreground">
                      {n.body}
                    </TableCell>
                    <TableCell>
                      <Badge variant="secondary" className="capitalize">
                        {n.target}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {n.sent_by}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {new Date(n.sent_at).toLocaleString()}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Confirmation Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Confirm Send</DialogTitle>
            <DialogDescription>
              Are you sure you want to send this notification? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <div className="rounded-lg border border-border bg-muted/50 p-4 space-y-2">
            <p className="text-sm text-foreground">
              <span className="font-medium">Title:</span> {title}
            </p>
            <Separator />
            <p className="text-sm text-foreground">
              <span className="font-medium">Target:</span>{" "}
              <span className="capitalize">{targetLabel}</span>
            </p>
          </div>
          <DialogFooter>
            <DialogClose render={<Button variant="outline" />}>
              Cancel
            </DialogClose>
            <Button
              onClick={confirmSend}
              disabled={sendMutation.isPending}
            >
              {sendMutation.isPending ? "Sending..." : "Confirm Send"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
