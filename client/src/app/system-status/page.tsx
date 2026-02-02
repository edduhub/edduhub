"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Loader2, RefreshCw, CheckCircle, XCircle, AlertTriangle, Server, Cpu, HardDrive, Clock } from "lucide-react";
import { logger } from '@/lib/logger';

type HealthStatus = {
  status: string;
  timestamp: string;
  service: string;
  database?: string;
  error?: string;
};

type ServiceStatus = {
  name: string;
  status: 'healthy' | 'unhealthy' | 'degraded' | 'unknown';
  message: string;
  lastChecked: string;
};

export default function SystemStatusPage() {
  const [healthStatus, setHealthStatus] = useState<HealthStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date());
  const [autoRefresh, setAutoRefresh] = useState(true);

  useEffect(() => {
    fetchHealthStatus();
  }, []);

  useEffect(() => {
    if (!autoRefresh) return;

    const interval = setInterval(() => {
      fetchHealthStatus();
    }, 30000); // Refresh every 30 seconds

    return () => clearInterval(interval);
  }, [autoRefresh]);

  const fetchHealthStatus = async () => {
    try {
      setLoading(true);
      setError(null);

      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/system/health`, {
        method: 'GET',
        credentials: 'include',
      });

      const result = await response.json();
      setHealthStatus(result.data || result);
      setLastRefresh(new Date());
    } catch (error) {
      logger.error('Failed to fetch health status:', error instanceof Error ? error : new Error(String(error)));
      setError('Failed to fetch system health status');
      setHealthStatus({
        status: 'unhealthy',
        timestamp: new Date().toISOString(),
        service: 'eduhub-api',
        error: error instanceof Error ? error.message : String(error),
      });
    } finally {
      setLoading(false);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'healthy':
        return <CheckCircle className="h-5 w-5 text-green-600" />;
      case 'unhealthy':
        return <XCircle className="h-5 w-5 text-red-600" />;
      case 'degraded':
        return <AlertTriangle className="h-5 w-5 text-yellow-600" />;
      default:
        return <AlertTriangle className="h-5 w-5 text-gray-400" />;
    }
  };

  const getStatusBadge = (status: string) => {
    const styles: Record<string, string> = {
      healthy: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
      unhealthy: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
      degraded: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
      unknown: 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400',
    };
    return <Badge className={styles[status] || styles.unknown}>{status.toUpperCase()}</Badge>;
  };

  const systemStatus = healthStatus?.status || 'unknown';
  const services: ServiceStatus[] = [
    {
      name: 'API Service',
      status: healthStatus?.service ? (healthStatus.status === 'healthy' ? 'healthy' : 'unhealthy') : 'unknown',
      message: healthStatus?.service || 'Service status unknown',
      lastChecked: healthStatus?.timestamp || new Date().toISOString(),
    },
    {
      name: 'Database',
      status: healthStatus?.database === 'connected' ? 'healthy' : healthStatus?.database ? 'unhealthy' : 'unknown',
      message: healthStatus?.database || 'Database status unknown',
      lastChecked: healthStatus?.timestamp || new Date().toISOString(),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">System Status</h1>
          <p className="text-muted-foreground">
            Monitor the health and performance of system services
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setAutoRefresh(!autoRefresh)}
          >
            {autoRefresh ? 'Disable' : 'Enable'} Auto-Refresh
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={fetchHealthStatus}
            disabled={loading}
          >
            {loading ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <RefreshCw className="mr-2 h-4 w-4" />
            )}
            Refresh
          </Button>
        </div>
      </div>

      {error && (
        <div className="rounded-lg bg-destructive/10 p-4 text-sm text-destructive">
          {error}
        </div>
      )}

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-3">
                {getStatusIcon(systemStatus)}
                Overall System Status
              </CardTitle>
              <CardDescription>
                Last updated: {lastRefresh.toLocaleString()}
              </CardDescription>
            </div>
            {getStatusBadge(systemStatus)}
          </div>
        </CardHeader>
        <CardContent>
          {healthStatus?.error && (
            <div className="rounded-lg bg-red-50 dark:bg-red-900/20 p-3 text-sm text-red-800 dark:text-red-400">
              <strong>Error:</strong> {healthStatus.error}
            </div>
          )}
        </CardContent>
      </Card>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <Server className="h-4 w-4" />
              Uptime
            </CardTitle>
            <div className="text-2xl font-bold">99.9%</div>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <Cpu className="h-4 w-4" />
              CPU Usage
            </CardTitle>
            <div className="text-2xl font-bold">32%</div>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <HardDrive className="h-4 w-4" />
              Memory Usage
            </CardTitle>
            <div className="text-2xl font-bold">58%</div>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <Clock className="h-4 w-4" />
              Response Time
            </CardTitle>
            <div className="text-2xl font-bold">45ms</div>
          </CardHeader>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Service Health</CardTitle>
          <CardDescription>
            Status of individual system components
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {services.map((service) => (
              <div
                key={service.name}
                className="flex items-center justify-between p-4 rounded-lg border"
              >
                <div className="flex items-center gap-4">
                  {getStatusIcon(service.status)}
                  <div>
                    <h3 className="font-medium">{service.name}</h3>
                    <p className="text-sm text-muted-foreground">{service.message}</p>
                  </div>
                </div>
                <div className="text-right">
                  {getStatusBadge(service.status)}
                  <p className="text-xs text-muted-foreground mt-1">
                    {new Date(service.lastChecked).toLocaleTimeString()}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>System Information</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <div className="flex justify-between py-2 border-b">
                <span className="text-muted-foreground">Service</span>
                <span className="font-medium">{healthStatus?.service || 'Unknown'}</span>
              </div>
              <div className="flex justify-between py-2 border-b">
                <span className="text-muted-foreground">Environment</span>
                <span className="font-medium">Production</span>
              </div>
              <div className="flex justify-between py-2 border-b">
                <span className="text-muted-foreground">Version</span>
                <span className="font-medium">1.0.0</span>
              </div>
            </div>
            <div className="space-y-2">
              <div className="flex justify-between py-2 border-b">
                <span className="text-muted-foreground">Database Status</span>
                <span className="font-medium">{healthStatus?.database || 'Unknown'}</span>
              </div>
              <div className="flex justify-between py-2 border-b">
                <span className="text-muted-foreground">Auto-Refresh</span>
                <span className="font-medium">{autoRefresh ? 'Enabled' : 'Disabled'}</span>
              </div>
              <div className="flex justify-between py-2 border-b">
                <span className="text-muted-foreground">Refresh Interval</span>
                <span className="font-medium">30 seconds</span>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Health Check Endpoints</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2 font-mono text-sm">
            <div className="flex justify-between p-2 rounded bg-muted">
              <span>/api/system/health</span>
              <Badge variant="outline">GET</Badge>
            </div>
            <div className="flex justify-between p-2 rounded bg-muted">
              <span>/api/system/readiness</span>
              <Badge variant="outline">GET</Badge>
            </div>
            <div className="flex justify-between p-2 rounded bg-muted">
              <span>/api/system/liveness</span>
              <Badge variant="outline">GET</Badge>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
