import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Users, Calendar, Swords, CreditCard } from "lucide-react";

export default function DashboardPage() {
  const stats = [
    { label: "Total Users", value: "--", icon: Users },
    { label: "Active Events", value: "--", icon: Calendar },
    { label: "Open Matches", value: "--", icon: Swords },
    { label: "Cards Issued", value: "--", icon: CreditCard },
  ];

  return (
    <div>
      <h1 className="text-2xl font-bold text-foreground mb-6">
        XEX Play Admin
      </h1>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
        {stats.map((stat) => (
          <Card key={stat.label}>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {stat.label}
                </CardTitle>
                <stat.icon className="size-4 text-muted-foreground" />
              </div>
            </CardHeader>
            <CardContent>
              <p className="text-3xl font-semibold text-foreground">
                {stat.value}
              </p>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
