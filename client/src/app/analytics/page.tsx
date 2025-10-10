import { fetchAnalytics } from "@/lib/api";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";

export default async function AnalyticsPage() {
  const metrics = await fetchAnalytics();

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
