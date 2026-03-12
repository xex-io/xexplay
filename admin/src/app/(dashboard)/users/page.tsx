"use client";

import { useState, useMemo } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { asArray } from "@/lib/loc-str";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "@/components/ui/table";
import { Search, ChevronLeft, ChevronRight, Shield } from "lucide-react";

interface User {
  id: string;
  email: string;
  display_name: string;
  xex_user_id: string;
  role: string;
  total_points: number;
  is_active: boolean;
  created_at: string;
}

interface PaginatedResponse {
  data: User[];
  meta?: {
    page: number;
    per_page: number;
    total: number;
  };
}

const PER_PAGE = 20;

function UserRow({
  user,
  toggleMutation,
  showAdminToggle,
}: {
  user: User;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  toggleMutation: any;
  showAdminToggle?: boolean;
}) {
  return (
    <TableRow>
      <TableCell>{user.email || "-"}</TableCell>
      <TableCell>{user.display_name || "-"}</TableCell>
      <TableCell className="font-mono text-muted-foreground">
        {user.xex_user_id?.slice(0, 8) || "-"}
      </TableCell>
      <TableCell className="font-mono">
        {(user.total_points ?? 0).toLocaleString()}
      </TableCell>
      <TableCell>
        <Switch
          checked={user.is_active}
          onCheckedChange={(checked: boolean) =>
            toggleMutation.mutate({
              id: user.id,
              data: { is_active: checked },
            })
          }
          disabled={toggleMutation.isPending}
        />
      </TableCell>
      {showAdminToggle && (
        <TableCell>
          <Switch
            checked={user.role === "admin"}
            onCheckedChange={(checked: boolean) =>
              toggleMutation.mutate({
                id: user.id,
                data: { role: checked ? "admin" : "user" },
              })
            }
            disabled={toggleMutation.isPending}
          />
        </TableCell>
      )}
      <TableCell className="text-muted-foreground">
        {new Date(user.created_at).toLocaleDateString()}
      </TableCell>
    </TableRow>
  );
}

export default function UsersPage() {
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [page, setPage] = useState(1);
  const [debounceTimer, setDebounceTimer] = useState<ReturnType<typeof setTimeout> | null>(null);
  const queryClient = useQueryClient();

  function handleSearchChange(value: string) {
    setSearch(value);
    if (debounceTimer) clearTimeout(debounceTimer);
    const timer = setTimeout(() => {
      setDebouncedSearch(value);
      setPage(1);
    }, 400);
    setDebounceTimer(timer);
  }

  const isSearching = debouncedSearch.length > 0;

  const {
    data: listResponse,
    isLoading: listLoading,
  } = useQuery<PaginatedResponse>({
    queryKey: ["admin-users", page],
    queryFn: async () => {
      const res = await apiClient.get("/admin/users", {
        params: { page, per_page: PER_PAGE },
      });
      return res.data;
    },
    enabled: !isSearching,
  });

  const {
    data: searchResponse,
    isLoading: searchLoading,
  } = useQuery<User[]>({
    queryKey: ["admin-users-search", debouncedSearch],
    queryFn: async () => {
      const res = await apiClient.get("/admin/users/search", {
        params: { q: debouncedSearch },
      });
      return asArray<User>(res);
    },
    enabled: isSearching,
  });

  const toggleMutation = useMutation({
    mutationFn: async ({ id, data }: { id: string; data: Record<string, unknown> }) => {
      return apiClient.put(`/admin/users/${id}`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-users"] });
      queryClient.invalidateQueries({ queryKey: ["admin-users-search"] });
    },
  });

  const allUsers = isSearching
    ? (Array.isArray(searchResponse) ? searchResponse : [])
    : (Array.isArray(listResponse?.data) ? listResponse.data : []);
  const total = listResponse?.meta?.total ?? 0;
  const totalPages = Math.ceil(total / PER_PAGE);
  const isLoading = isSearching ? searchLoading : listLoading;

  const admins = useMemo(() => allUsers.filter((u) => u.role === "admin"), [allUsers]);
  const regularUsers = useMemo(() => allUsers.filter((u) => u.role !== "admin"), [allUsers]);

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-foreground">Users</h1>
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
          <Input
            type="text"
            placeholder="Search users..."
            value={search}
            onChange={(e) => handleSearchChange(e.target.value)}
            className="pl-8 w-64"
          />
        </div>
      </div>

      {/* Admins Section */}
      {admins.length > 0 && (
        <Card className="mb-6">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-sm uppercase tracking-wider text-muted-foreground">
              <Shield className="size-4" />
              Administrators ({admins.length})
            </CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Email</TableHead>
                  <TableHead>Display Name</TableHead>
                  <TableHead>XEX User ID</TableHead>
                  <TableHead>Points</TableHead>
                  <TableHead>Active</TableHead>
                  <TableHead>Admin</TableHead>
                  <TableHead>Joined</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {admins.map((user) => (
                  <UserRow
                    key={user.id}
                    user={user}
                    toggleMutation={toggleMutation}
                    showAdminToggle
                  />
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      {/* Regular Users Section */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm uppercase tracking-wider text-muted-foreground">
            Users {!isLoading && `(${isSearching ? regularUsers.length : total})`}
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Email</TableHead>
                <TableHead>Display Name</TableHead>
                <TableHead>XEX User ID</TableHead>
                <TableHead>Points</TableHead>
                <TableHead>Active</TableHead>
                <TableHead>Joined</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell
                    colSpan={6}
                    className="h-24 text-center text-muted-foreground"
                  >
                    {isSearching ? "Searching..." : "Loading users..."}
                  </TableCell>
                </TableRow>
              ) : regularUsers.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={6}
                    className="h-24 text-center text-muted-foreground"
                  >
                    {isSearching
                      ? "No users match your search."
                      : "No users found."}
                  </TableCell>
                </TableRow>
              ) : (
                regularUsers.map((user) => (
                  <UserRow
                    key={user.id}
                    user={user}
                    toggleMutation={toggleMutation}
                  />
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Pagination controls - only shown for list view */}
      {!isSearching && totalPages > 1 && (
        <div className="flex items-center justify-between mt-4">
          <p className="text-sm text-muted-foreground">
            Page {page} of {totalPages} ({total} users)
          </p>
          <div className="flex gap-2">
            <Button
              size="sm"
              variant="outline"
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
            >
              <ChevronLeft className="size-4" data-icon="inline-start" />
              Previous
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
            >
              Next
              <ChevronRight className="size-4" data-icon="inline-end" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
