"use client";

import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { 
  TrendingUp, 
  TrendingDown, 
  BarChart3, 
  PieChart, 
  LineChart,
  BookOpen,
  Clock,
  Target,
  AlertTriangle,
  CheckCircle,
  Download,
  Calendar
} from 'lucide-react';
import type { 
  PerformanceMetrics, 
  AttendanceTrend, 
  LearningAnalytics,
  PredictiveInsight,
  PerformanceTrend,
  Course,
  Grade
} from '@/lib/types';

export default function StudentAnalyticsPage() {
  const [performanceData, setPerformanceData] = useState<PerformanceMetrics | null>(null);
  const [attendanceTrends, setAttendanceTrends] = useState<AttendanceTrend[]>([]);
  const [learningAnalytics, setLearningAnalytics] = useState<LearningAnalytics | null>(null);
  const [predictiveInsights, setPredictiveInsights] = useState<PredictiveInsight[]>([]);
  const [performanceTrends, setPerformanceTrends] = useState<PerformanceTrend[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    fetchAnalyticsData();
  }, []);

  const fetchAnalyticsData = async () => {
    try {
      // Mock data - replace with API calls
      setPerformanceData({
        studentId:1,
        courseId: 101,
        averageScore: 82.5,
        highestScore: 95,
        lowestScore: 70,
        totalAssessments: 12,
        trend: 'improving',
        percentile: 75,
        gradeDistribution: [
          { grade: 'A', count: 5 },
          { grade: 'B', count: 4 },
          { grade: 'C', count: 2 },
          { grade: 'D', count: 1 },
        ],
      });

      setAttendanceTrends([
        {
          period: 'January',
          presentCount: 20,
          absentCount: 2,
          lateCount: 1,
          excusedCount: 0,
          attendanceRate: 87,
        },
        {
          period: 'February',
          presentCount: 18,
          absentCount: 3,
          lateCount: 1,
          excusedCount: 1,
          attendanceRate: 78,
        },
        {
          period: 'March',
          presentCount: 22,
          absentCount: 1,
          lateCount: 0,
          excusedCount: 0,
          attendanceRate: 96,
        },
      ]);

      setLearningAnalytics({
        period: 'Current Semester',
        engagementRate: 85,
        completionRate: 78,
        averageTimeSpent: 45,
        mostAccessedMaterials: [
          { materialId: 1, title: 'Introduction to Algorithms', accessCount: 156 },
          { materialId: 2, title: 'Data Structures', accessCount: 132 },
        ],
        leastAccessedMaterials: [
          { materialId: 3, title: 'Advanced Topics', accessCount: 23 },
        ],
        peakActivityHours: [
          { hour: 10, activityCount: 45 },
          { hour: 14, activityCount: 67 },
          { hour: 19, activityCount: 52 },
        ],
      });

      setPredictiveInsights([
        {
          studentId: 1,
          studentName: 'John Doe',
          riskLevel: 'medium',
          factors: [
            'Declining attendance in last 2 weeks',
            'Late assignment submissions',
          ],
          recommendations: [
            'Focus on attending classes regularly',
            'Start assignments early',
            'Attend office hours for help',
          ],
          confidenceScore: 0.85,
        },
      ]);

      setPerformanceTrends([
        {
          date: '2024-01-15',
          score: 78,
          percentile: 68,
        },
        {
          date: '2024-02-15',
          score: 82,
          percentile: 72,
        },
        {
          date: '2024-03-15',
          score: 85,
          percentile: 75,
        },
      ]);
    } catch (error) {
      console.error('Failed to fetch analytics:', error);
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
          <p className="mt-4 text-muted-foreground">Loading analytics...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-muted/10">
      <div className="container mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">Student Analytics</h1>
            <p className="text-muted-foreground mt-1">
              Track your academic performance and learning progress
            </p>
          </div>
          <Button variant="outline">
            <Download className="w-4 h-4 mr-2" />
            Export Report
          </Button>
        </div>

        {/* Predictive Insights */}
        {predictiveInsights.length > 0 && (
          <div className="space-y-4">
            <h2 className="text-xl font-semibold">Insights & Recommendations</h2>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {predictiveInsights.map((insight) => (
                <PredictiveInsightCard key={insight.studentId} insight={insight} />
              ))}
            </div>
          </div>
        )}

        {/* Performance Metrics */}
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          <PerformanceMetricCard
            title="Average Score"
            value={performanceData?.averageScore || 0}
            suffix="%"
            trend={performanceData?.trend}
          />
          <PerformanceMetricCard
            title="Attendance Rate"
            value={learningAnalytics?.engagementRate || 0}
            suffix="%"
            trend="stable"
          />
          <PerformanceMetricCard
            title="Completion Rate"
            value={learningAnalytics?.completionRate || 0}
            suffix="%"
            trend="improving"
          />
          <PerformanceMetricCard
            title="Assessments"
            value={performanceData?.totalAssessments || 0}
            trend={null}
          />
        </div>

        {/* Detailed Analytics Tabs */}
        <Tabs defaultValue="performance" className="space-y-4">
          <TabsList className="grid w-full grid-cols-4 lg:w-auto">
            <TabsTrigger value="performance">Performance</TabsTrigger>
            <TabsTrigger value="attendance">Attendance</TabsTrigger>
            <TabsTrigger value="learning">Learning</TabsTrigger>
            <TabsTrigger value="progress">Progress</TabsTrigger>
          </TabsList>

          <TabsContent value="performance" className="space-y-4">
            <PerformanceSection performanceData={performanceData} performanceTrends={performanceTrends} />
          </TabsContent>

          <TabsContent value="attendance" className="space-y-4">
            <AttendanceSection attendanceTrends={attendanceTrends} />
          </TabsContent>

          <TabsContent value="learning" className="space-y-4">
            <LearningSection learningAnalytics={learningAnalytics} />
          </TabsContent>

          <TabsContent value="progress" className="space-y-4">
            <ProgressSection />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}

function PredictiveInsightCard({ insight }: { insight: PredictiveInsight }) {
  const riskColors = {
    low: 'bg-green-100 text-green-800 border-green-200',
    medium: 'bg-yellow-100 text-yellow-800 border-yellow-200',
    high: 'bg-red-100 text-red-800 border-red-200',
  };

  const riskIcons = {
    low: <CheckCircle className="w-5 h-5 text-green-600" />,
    medium: <AlertTriangle className="w-5 h-5 text-yellow-600" />,
    high: <AlertTriangle className="w-5 h-5 text-red-600" />,
  };

  return (
    <Card className={`border-2 ${riskColors[insight.riskLevel]}`}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            {riskIcons[insight.riskLevel]}
            {insight.riskLevel.charAt(0).toUpperCase() + insight.riskLevel.slice(1)} Risk
          </CardTitle>
          <Badge variant="outline">
            {Math.round(insight.confidenceScore * 100)}% confidence
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <h4 className="font-semibold mb-2">Risk Factors:</h4>
          <ul className="space-y-1">
            {insight.factors.map((factor, i) => (
              <li key={i} className="text-sm text-muted-foreground flex items-start gap-2">
                <span className="text-red-500">•</span>
                {factor}
              </li>
            ))}
          </ul>
        </div>
        <div>
          <h4 className="font-semibold mb-2">Recommendations:</h4>
          <ul className="space-y-1">
            {insight.recommendations.map((rec, i) => (
              <li key={i} className="text-sm text-muted-foreground flex items-start gap-2">
                <span className="text-green-500">✓</span>
                {rec}
              </li>
            ))}
          </ul>
        </div>
      </CardContent>
    </Card>
  );
}

function PerformanceMetricCard({ 
  title, 
  value, 
  suffix, 
  trend 
}: { 
  title: string
  value: number
  suffix?: string
  trend: 'improving' | 'declining' | 'stable' | null
}) {
  const trendColors = {
    improving: 'text-green-600',
    declining: 'text-red-600',
    stable: 'text-gray-600',
  };

  const trendIcons = {
    improving: <TrendingUp className="w-5 h-5" />,
    declining: <TrendingDown className="w-5 h-5" />,
    stable: null,
  };

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium">
          {title}
        </CardTitle>
        <BarChart3 className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="flex items-center justify-between">
          <div className="text-2xl font-bold">
            {value}{suffix}
          </div>
          {trend && (
            <div className={`flex items-center gap-1 ${trendColors[trend]}`}>
              {trendIcons[trend]}
              <span className="text-sm">{trend}</span>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

function PerformanceSection({ 
  performanceData, 
  performanceTrends 
}: { 
  performanceData: PerformanceMetrics | null
  performanceTrends: PerformanceTrend[]
}) {
  if (!performanceData) {
    return <div className="text-center py-8 text-muted-foreground">No performance data available</div>;
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <BarChart3 className="w-5 h-5" />
            Performance Overview
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="grid gap-4 md:grid-cols-3">
            <div className="p-4 bg-blue-50 rounded-lg">
              <div className="text-sm text-muted-foreground">Average Score</div>
              <div className="text-2xl font-bold">{performanceData.averageScore}%</div>
            </div>
            <div className="p-4 bg-green-50 rounded-lg">
              <div className="text-sm text-muted-foreground">Highest Score</div>
              <div className="text-2xl font-bold">{performanceData.highestScore}%</div>
            </div>
            <div className="p-4 bg-red-50 rounded-lg">
              <div className="text-sm text-muted-foreground">Lowest Score</div>
              <div className="text-2xl font-bold">{performanceData.lowestScore}%</div>
            </div>
          </div>

          <div>
            <h4 className="font-semibold mb-3">Grade Distribution</h4>
            <div className="flex gap-2 items-end h-32">
              {performanceData.gradeDistribution?.map((item) => (
                <div 
                  key={item.grade} 
                  className="flex-1 bg-primary rounded-t-md relative group"
                  style={{ 
                    height: `${(item.count / Math.max(...performanceData.gradeDistribution.map(g => g.count))) * 100}%` 
                  }}
                >
                  <div className="absolute -top-6 left-1/2 -translate-x-1/2 opacity-0 group-hover:opacity-100 transition-opacity">
                    <Badge>{item.count}</Badge>
                  </div>
                </div>
              ))}
            </div>
            <div className="flex gap-2 mt-2">
              {performanceData.gradeDistribution?.map((item) => (
                <div key={item.grade} className="flex-1 text-center text-sm">
                  {item.grade}
                </div>
              ))}
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <LineChart className="w-5 h-5" />
            Performance Trends
          </CardTitle>
          <CardDescription>Your performance over time</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {performanceTrends.map((trend) => (
              <div key={trend.date} className="flex items-center justify-between p-3 border rounded-lg">
                <div>
                  <div className="font-medium">{new Date(trend.date).toLocaleDateString()}</div>
                  <div className="text-sm text-muted-foreground">
                    Percentile: {trend.percentile}th
                  </div>
                </div>
                <div className="text-2xl font-bold">{trend.score}%</div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function AttendanceSection({ attendanceTrends }: { attendanceTrends: AttendanceTrend[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Calendar className="w-5 h-5" />
          Attendance Trends
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {attendanceTrends.map((trend) => (
            <div key={trend.period} className="space-y-2">
              <div className="flex items-center justify-between">
                <span className="font-semibold">{trend.period}</span>
                <Badge 
                  variant={trend.attendanceRate >= 80 ? 'default' : 'destructive'}
                >
                  {trend.attendanceRate}% attendance
                </Badge>
              </div>
              <div className="flex gap-2 text-sm">
                <div className="flex items-center gap-1">
                  <CheckCircle className="w-4 h-4 text-green-600" />
                  Present: {trend.presentCount}
                </div>
                <div className="flex items-center gap-1">
                  <AlertTriangle className="w-4 h-4 text-red-600" />
                  Absent: {trend.absentCount}
                </div>
                <div className="flex items-center gap-1">
                  <Clock className="w-4 h-4 text-yellow-600" />
                  Late: {trend.lateCount}
                </div>
                {trend.excusedCount > 0 && (
                  <div className="flex items-center gap-1">
                    Excused: {trend.excusedCount}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

function LearningSection({ learningAnalytics }: { learningAnalytics: LearningAnalytics | null }) {
  if (!learningAnalytics) {
    return <div className="text-center py-8 text-muted-foreground">No learning analytics available</div>;
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Target className="w-5 h-5" />
            Learning Engagement
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3">
            <div className="p-4 bg-blue-50 rounded-lg">
              <div className="text-sm text-muted-foreground">Engagement Rate</div>
              <div className="text-2xl font-bold">{learningAnalytics.engagementRate}%</div>
            </div>
            <div className="p-4 bg-green-50 rounded-lg">
              <div className="text-sm text-muted-foreground">Completion Rate</div>
              <div className="text-2xl font-bold">{learningAnalytics.completionRate}%</div>
            </div>
            <div className="p-4 bg-purple-50 rounded-lg">
              <div className="text-sm text-muted-foreground">Avg. Time Spent</div>
              <div className="text-2xl font-bold">{learningAnalytics.averageTimeSpent}m</div>
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Most Accessed Materials</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {learningAnalytics.mostAccessedMaterials?.map((material) => (
                <div key={material.materialId} className="flex justify-between items-center p-2 bg-muted/50 rounded">
                  <span className="text-sm">{material.title}</span>
                  <Badge>{material.accessCount} views</Badge>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Peak Activity Hours</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {learningAnalytics.peakActivityHours?.map((hour) => (
                <div key={hour.hour} className="flex justify-between items-center p-2 bg-muted/50 rounded">
                  <span className="text-sm">{hour.hour}:00</span>
                  <Badge>{hour.activityCount} activities</Badge>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function ProgressSection() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <BookOpen className="w-5 h-5" />
          Course Progress
        </CardTitle>
        <CardDescription>Your progress in enrolled courses</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          {[
            { name: 'Computer Science 101', progress: 75, grade: 'A-' },
            { name: 'Mathematics 201', progress: 60, grade: 'B+' },
            { name: 'Physics 102', progress: 85, grade: 'A' },
          ].map((course) => (
            <div key={course.name} className="space-y-2">
              <div className="flex justify-between items-center">
                <span className="font-medium">{course.name}</span>
                <Badge variant="outline">{course.grade}</Badge>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div 
                  className="bg-primary h-2 rounded-full transition-all"
                  style={{ width: `${course.progress}%` }}
                />
              </div>
              <div className="text-sm text-muted-foreground text-right">
                {course.progress}% complete
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
