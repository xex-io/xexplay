"use client";

import { useState, useEffect } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "@/components/ui/table";
import { Save, Trash2, Eye, EyeOff, KeyRound } from "lucide-react";

interface SettingView {
  key: string;
  value: string;
  description: string;
  is_secret: boolean;
  has_value: boolean;
  updated_at: string;
}

export default function SettingsPage() {
  const queryClient = useQueryClient();
  const [editingKey, setEditingKey] = useState<string | null>(null);
  const [editValue, setEditValue] = useState("");
  const [showSecrets, setShowSecrets] = useState<Record<string, boolean>>({});
  const [mounted, setMounted] = useState(false);

  useEffect(() => { setMounted(true); }, []);

  const { data: settings, isLoading } = useQuery<SettingView[]>({
    queryKey: ["admin-settings"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/settings");
      const d = res.data?.data ?? res.data;
      return Array.isArray(d) ? d : [];
    },
    enabled: mounted,
  });

  const updateMutation = useMutation({
    mutationFn: async ({ key, value }: { key: string; value: string }) => {
      return apiClient.put(`/admin/settings/${key}`, { value });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-settings"] });
      setEditingKey(null);
      setEditValue("");
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (key: string) => {
      return apiClient.delete(`/admin/settings/${key}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-settings"] });
    },
  });

  function startEditing(setting: SettingView) {
    setEditingKey(setting.key);
    setEditValue(setting.is_secret ? "" : setting.value);
  }

  function toggleShowSecret(key: string) {
    setShowSecrets((prev) => ({ ...prev, [key]: !prev[key] }));
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Settings</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Manage API keys and automation configuration. Changes take effect on
            next server restart or job trigger.
          </p>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-sm uppercase tracking-wider text-muted-foreground">
            <KeyRound className="size-4" />
            Configuration
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Key</TableHead>
                <TableHead>Value</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Updated</TableHead>
                <TableHead className="w-[140px]">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell
                    colSpan={5}
                    className="h-24 text-center text-muted-foreground"
                  >
                    Loading settings...
                  </TableCell>
                </TableRow>
              ) : !settings || settings.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={5}
                    className="h-24 text-center text-muted-foreground"
                  >
                    No settings found.
                  </TableCell>
                </TableRow>
              ) : (
                settings.map((setting) => (
                  <TableRow key={setting.key}>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <code className="text-sm font-mono">
                          {setting.key}
                        </code>
                        {setting.is_secret && (
                          <Badge variant="outline" className="text-xs">
                            Secret
                          </Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      {editingKey === setting.key ? (
                        <div className="flex items-center gap-2">
                          <Input
                            type={setting.is_secret && !showSecrets[setting.key] ? "password" : "text"}
                            value={editValue}
                            onChange={(e) => setEditValue(e.target.value)}
                            placeholder={
                              setting.is_secret
                                ? "Enter new value..."
                                : "Enter value..."
                            }
                            className="h-8 w-64 font-mono text-sm"
                            autoFocus
                          />
                          {setting.is_secret && (
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() => toggleShowSecret(setting.key)}
                              className="h-8 w-8 p-0"
                            >
                              {showSecrets[setting.key] ? (
                                <EyeOff className="size-4" />
                              ) : (
                                <Eye className="size-4" />
                              )}
                            </Button>
                          )}
                          <Button
                            size="sm"
                            onClick={() =>
                              updateMutation.mutate({
                                key: setting.key,
                                value: editValue,
                              })
                            }
                            disabled={updateMutation.isPending}
                            className="h-8"
                          >
                            <Save className="size-4 mr-1" />
                            Save
                          </Button>
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => {
                              setEditingKey(null);
                              setEditValue("");
                            }}
                            className="h-8"
                          >
                            Cancel
                          </Button>
                        </div>
                      ) : (
                        <div className="flex items-center gap-2">
                          {setting.is_secret ? (
                            <span className="text-sm text-muted-foreground font-mono">
                              {setting.has_value ? setting.value : "(not set)"}
                            </span>
                          ) : (
                            <code className="text-sm font-mono">
                              {setting.value || "(not set)"}
                            </code>
                          )}
                        </div>
                      )}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground max-w-[200px]">
                      {setting.description}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {setting.updated_at
                        ? new Date(setting.updated_at).toISOString().split("T")[0]
                        : "-"}
                    </TableCell>
                    <TableCell>
                      {editingKey !== setting.key && (
                        <div className="flex items-center gap-1">
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => startEditing(setting)}
                            className="h-8"
                          >
                            Edit
                          </Button>
                          {setting.has_value && (
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() =>
                                deleteMutation.mutate(setting.key)
                              }
                              disabled={deleteMutation.isPending}
                              className="h-8 text-destructive hover:text-destructive"
                            >
                              <Trash2 className="size-4" />
                            </Button>
                          )}
                        </div>
                      )}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
