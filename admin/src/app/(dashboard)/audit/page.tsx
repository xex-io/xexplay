"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";

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

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-100">Audit Log</h1>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap items-end gap-4 mb-4">
        <div>
          <label className="block text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">
            Admin User
          </label>
          <input
            type="text"
            value={filterAdmin}
            onChange={(e) => setFilterAdmin(e.target.value)}
            placeholder="Filter by admin..."
            className="bg-gray-900 border border-gray-700 text-gray-200 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">
            Action Type
          </label>
          <input
            type="text"
            value={filterAction}
            onChange={(e) => setFilterAction(e.target.value)}
            placeholder="Filter by action..."
            className="bg-gray-900 border border-gray-700 text-gray-200 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
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
        {(filterAdmin || filterAction || dateFrom || dateTo) && (
          <button
            onClick={() => {
              setFilterAdmin("");
              setFilterAction("");
              setDateFrom("");
              setDateTo("");
            }}
            className="text-gray-400 hover:text-gray-200 text-sm font-medium transition-colors pb-1"
          >
            Clear filters
          </button>
        )}
      </div>

      {/* Table */}
      <div className="bg-gray-800 shadow rounded-lg border border-gray-700 overflow-hidden">
        <table className="min-w-full divide-y divide-gray-700">
          <thead className="bg-gray-900">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider w-8">
                {/* Expand */}
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Timestamp
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Admin User
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Action
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Entity Type
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                Entity ID
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                IP Address
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-700">
            {isLoading ? (
              <tr>
                <td colSpan={7} className="px-6 py-12 text-center text-sm text-gray-400">
                  Loading audit logs...
                </td>
              </tr>
            ) : logs.length === 0 ? (
              <tr>
                <td colSpan={7} className="px-6 py-12 text-center text-sm text-gray-400">
                  No audit logs found.
                </td>
              </tr>
            ) : (
              logs.map((log) => (
                <>
                  <tr
                    key={log.id}
                    className="hover:bg-gray-750 transition-colors cursor-pointer"
                    onClick={() =>
                      setExpandedRow(expandedRow === log.id ? null : log.id)
                    }
                  >
                    <td className="px-6 py-4 text-sm text-gray-400">
                      <svg
                        className={`w-4 h-4 transition-transform ${
                          expandedRow === log.id ? "rotate-90" : ""
                        }`}
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        strokeWidth={2}
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          d="M9 5l7 7-7 7"
                        />
                      </svg>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-300">
                      {new Date(log.created_at).toLocaleString()}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-200">
                      {log.admin_user}
                    </td>
                    <td className="px-6 py-4 text-sm">
                      <span className="inline-flex px-2.5 py-0.5 rounded-full text-xs font-semibold bg-blue-500/20 text-blue-400 border border-blue-500/30">
                        {log.action}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-300 capitalize">
                      {log.entity_type}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-300 font-mono">
                      {log.entity_id.slice(0, 8)}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-400 font-mono">
                      {log.ip_address}
                    </td>
                  </tr>
                  {expandedRow === log.id && (
                    <tr key={`${log.id}-details`}>
                      <td colSpan={7} className="px-6 py-4 bg-gray-900">
                        <div className="mb-2">
                          <span className="text-xs font-medium text-gray-400 uppercase tracking-wider">
                            Full Details
                          </span>
                        </div>
                        <pre className="text-xs text-gray-300 bg-gray-950 border border-gray-700 rounded-lg p-4 overflow-x-auto max-h-64">
                          {JSON.stringify(log.details, null, 2)}
                        </pre>
                      </td>
                    </tr>
                  )}
                </>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
