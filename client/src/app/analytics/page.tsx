"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api-client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Loader2 } from "lucide-react";

type AnalyticsMetric = {
  label: string;
  value: number;
  delta?: number;
};

export default function AnalyticsPage() {
  const [metrics, setMetrics] = useState<AnalyticsMetric[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchAnalytics = async () => {
      try {
        setLoading(true);
        const response = await api.get('/api/analytics');
        setMetrics(Array.isArray(response) ? response : []);
      } catch (err) {
        console.error('Failed to fetch analytics:', err);
        setError('Failed to load analytics data');
      } finally {
        setLoading(false);
      }
    };

    fetchAnalytics();
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Analytics</CardTitle>
          <CardDescription>Key indicators driving academic performance.</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-6 md:grid-cols-2">
          {metrics.map((metric) => (
            <div key={metric.label} className="rounded-lg border p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">{metric.label}</p>
                  <p className="text-2xl font-semibold">{metric.value}</p>
                </div>
                {metric.delta !== undefined && (
                  <span className="text-xs text-muted-foreground">Δ {metric.delta >= 0 ? "+" : ""}{metric.delta}</span>
                )}
              </div>
              <Progress value={Math.min(100, Number(metric.value) * 25)} className="mt-3" />
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}
