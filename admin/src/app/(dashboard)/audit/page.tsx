"use client";

import { useState, Fragment } from "react";
import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { ChevronRight, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "@/components/ui/table";
import { cn } from "@/lib/utils";

interface AuditLog {
  id: string;
  admin_user: string;
  action: string;
  entity_type: string;
  entity_id: string;
  ip_address: string;
  details: Record<string, unknown>;
  created_at: string;
}

export default function AuditPage() {
  const [filterAdmin, setFilterAdmin] = useState("");
  const [filterAction, setFilterAction] = useState("");
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");
  const [expandedRow, setExpandedRow] = useState<string | null>(null);

  const { data: logs = [], isLoading } = useQuery<AuditLog[]>({
    queryKey: ["admin-audit-logs", filterAdmin, filterAction, dateFrom, dateTo],
    queryFn: async () => {
      const params: Record<string, string> = {};
      if (filterAdmin) params.admin_user = filterAdmin;
      if (filterAction) params.action = filterAction;
      if (dateFrom) params.date_from = dateFrom;
      if (dateTo) params.date_to = dateTo;
      const res = await apiClient.get("/admin/audit-logs", { params });
      return res.data?.data ?? res.data ?? [];
    },
  });

  const hasFilters = filterAdmin || filterAction || dateFrom || dateTo;

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-foreground">Audit Log</h1>
      </div>

      {/* Filters */}
      <Card className="mb-4">
        <CardContent>
          <div className="flex flex-wrap items-end gap-4">
            <div className="space-y-1.5">
              <Label htmlFor="filter-admin" className="text-xs uppercase tracking-wider text-muted-foreground">
                Admin User
              </Label>
              <Input
                id="filter-admin"
                type="text"
                value={filterAdmin}
                onChange={(e) => setFilterAdmin(e.target.value)}
                placeholder="Filter by admin..."
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="filter-action" className="text-xs uppercase tracking-wider text-muted-foreground">
                Action Type
              </Label>
              <Input
                id="filter-action"
                type="text"
                value={filterAction}
                onChange={(e) => setFilterAction(e.target.value)}
                placeholder="Filter by action..."
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="filter-date-from" className="text-xs uppercase tracking-wider text-muted-foreground">
                Date From
              </Label>
              <Input
                id="filter-date-from"
                type="date"
                value={dateFrom}
                onChange={(e) => setDateFrom(e.target.value)}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="filter-date-to" className="text-xs uppercase tracking-wider text-muted-foreground">
                Date To
              </Label>
              <Input
                id="filter-date-to"
                type="date"
                value={dateTo}
                onChange={(e) => setDateTo(e.target.value)}
              />
            </div>
            {hasFilters && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  setFilterAdmin("");
                  setFilterAction("");
                  setDateFrom("");
                  setDateTo("");
                }}
              >
                <X data-icon="inline-start" />
                Clear filters
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Table */}
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-8">
                  {/* Expand */}
                </TableHead>
                <TableHead>Timestamp</TableHead>
                <TableHead>Admin User</TableHead>
                <TableHead>Action</TableHead>
                <TableHead>Entity Type</TableHead>
                <TableHead>Entity ID</TableHead>
                <TableHead>IP Address</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center py-12 text-muted-foreground">
                    Loading audit logs...
                  </TableCell>
                </TableRow>
              ) : logs.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center py-12 text-muted-foreground">
                    No audit logs found.
                  </TableCell>
                </TableRow>
              ) : (
                logs.map((log) => (
                  <Fragment key={log.id}>
                    <TableRow
                      className="cursor-pointer"
                      onClick={() =>
                        setExpandedRow(expandedRow === log.id ? null : log.id)
                      }
                    >
                      <TableCell className="text-muted-foreground">
                        <ChevronRight
                          className={cn(
                            "size-4 transition-transform",
                            expandedRow === log.id && "rotate-90"
                          )}
                        />
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {new Date(log.created_at).toLocaleString()}
                      </TableCell>
                      <TableCell className="text-foreground">
                        {log.admin_user}
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">
                          {log.action}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-muted-foreground capitalize">
                        {log.entity_type}
                      </TableCell>
                      <TableCell className="text-muted-foreground font-mono">
                        {log.entity_id.slice(0, 8)}
                      </TableCell>
                      <TableCell className="text-muted-foreground font-mono">
                        {log.ip_address}
                      </TableCell>
                    </TableRow>
                    {expandedRow === log.id && (
                      <TableRow>
                        <TableCell colSpan={7} className="bg-muted/50 p-4">
                          <div className="mb-2">
                            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                              Full Details
                            </span>
                          </div>
                          <pre className="text-xs text-foreground bg-background border border-border rounded-lg p-4 overflow-x-auto max-h-64">
                            {JSON.stringify(log.details, null, 2)}
                          </pre>
                        </TableCell>
                      </TableRow>
                    )}
                  </Fragment>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
