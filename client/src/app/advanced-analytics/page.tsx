"use client";

import { useState, useEffect } from "react";
import { api } from "@/lib/api-client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Label } from "@/components/ui/label";
import { Progress } from "@/components/ui/progress";
import { Loader2, TrendingUp, TrendingDown, Users, Target, Award, BookOpen, Brain, AlertTriangle } from "lucide-react";

type StudentProgression = {
  student_id: number;
  student_name: string;
  overall_gpa: number;
  courses_completed: number;
  courses_in_progress: number;
  attendance_rate: number;
  performance_trend: 'improving' | 'declining' | 'stable';
  predicted_gpa: number;
  at_risk: boolean;
  strengths: string[];
  weaknesses: string[];
};

type CourseEngagement = {
  course_id: number;
  course_name: string;
  total_students: number;
  active_students: number;
  engagement_rate: number;
  average_grade: number;
  completion_rate: number;
  attendance_rate: number;
  assignment_submission_rate: number;
  most_engaged_students: string[];
  least_engaged_students: string[];
};

type PredictiveInsight = {
  type: 'at_risk' | 'high_performer' | 'improvement_needed';
  student_id: number;
  student_name: string;
  confidence: number;
  factors: string[];
  recommendations: string[];
};

type LearningAnalytics = {
  total_students: number;
  average_gpa: number;
  retention_rate: number;
  graduation_rate: number;
  top_performing_courses: Array<{ name: string; avg_grade: number }>;
  struggling_areas: Array<{ name: string; failure_rate: number }>;
};

export default function AdvancedAnalyticsPage() {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Student Progression
  const [students, setStudents] = useState<any[]>([]);
  const [selectedStudent, setSelectedStudent] = useState<string>("");
  const [studentProgression, setStudentProgression] = useState<StudentProgression | null>(null);
  const [loadingProgression, setLoadingProgression] = useState(false);

  // Course Engagement
  const [courses, setCourses] = useState<any[]>([]);
  const [selectedCourse, setSelectedCourse] = useState<string>("");
  const [courseEngagement, setCourseEngagement] = useState<CourseEngagement | null>(null);
  const [loadingEngagement, setLoadingEngagement] = useState(false);

  // Predictive Insights
  const [insights, setInsights] = useState<PredictiveInsight[]>([]);
  const [loadingInsights, setLoadingInsights] = useState(false);

  // Learning Analytics
  const [learningAnalytics, setLearningAnalytics] = useState<LearningAnalytics | null>(null);

  useEffect(() => {
    fetchInitialData();
  }, []);

  const fetchInitialData = async () => {
    try {
      setLoading(true);
      await Promise.all([
        fetchStudents(),
        fetchCourses(),
        fetchLearningAnalytics(),
      ]);
    } catch (err) {
      console.error('Failed to fetch initial data:', err);
      setError('Failed to load analytics data');
    } finally {
      setLoading(false);
    }
  };

  const fetchStudents = async () => {
    try {
      const response = await api.get('/api/students');
      setStudents(Array.isArray(response) ? response : []);
    } catch (err) {
      console.error('Failed to fetch students:', err);
    }
  };

  const fetchCourses = async () => {
    try {
      const response = await api.get('/api/courses');
      setCourses(Array.isArray(response) ? response : []);
    } catch (err) {
      console.error('Failed to fetch courses:', err);
    }
  };

  const fetchStudentProgression = async (studentId: string) => {
    try {
      setLoadingProgression(true);
      const response = await api.get(`/api/analytics/advanced/students/${studentId}/progression`);
      setStudentProgression(response);
    } catch (err) {
      console.error('Failed to fetch student progression:', err);
      setError('Failed to load student progression');
    } finally {
      setLoadingProgression(false);
    }
  };

  const fetchCourseEngagement = async (courseId: string) => {
    try {
      setLoadingEngagement(true);
      const response = await api.get(`/api/analytics/advanced/courses/${courseId}/engagement`);
      setCourseEngagement(response);
    } catch (err) {
      console.error('Failed to fetch course engagement:', err);
      setError('Failed to load course engagement');
    } finally {
      setLoadingEngagement(false);
    }
  };

  const fetchPredictiveInsights = async () => {
    try {
      setLoadingInsights(true);
      const response = await api.get('/api/analytics/advanced/predictive-insights');
      setInsights(Array.isArray(response) ? response : []);
    } catch (err) {
      console.error('Failed to fetch insights:', err);
      setError('Failed to load predictive insights');
    } finally {
      setLoadingInsights(false);
    }
  };

  const fetchLearningAnalytics = async () => {
    try {
      const response = await api.get('/api/analytics/advanced/learning-analytics');
      setLearningAnalytics(response);
    } catch (err) {
      console.error('Failed to fetch learning analytics:', err);
    }
  };

  const handleStudentChange = (value: string) => {
    setSelectedStudent(value);
    if (value) {
      fetchStudentProgression(value);
    }
  };

  const handleCourseChange = (value: string) => {
    setSelectedCourse(value);
    if (value) {
      fetchCourseEngagement(value);
    }
  };

  const getTrendIcon = (trend: string) => {
    if (trend === 'improving') return <TrendingUp className="h-5 w-5 text-green-600" />;
    if (trend === 'declining') return <TrendingDown className="h-5 w-5 text-red-600" />;
    return <div className="h-5 w-5" />;
  };

  const getInsightBadge = (type: string) => {
    const styles: Record<string, string> = {
      at_risk: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
      high_performer: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
      improvement_needed: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
    };
    return <Badge className={styles[type]}>{type.replace('_', ' ').toUpperCase()}</Badge>;
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Advanced Analytics</h1>
        <p className="text-muted-foreground">
          Deep insights into student progression and course engagement
        </p>
      </div>

      {error && (
        <div className="rounded-lg bg-destructive/10 p-4 text-sm text-destructive">
          {error}
        </div>
      )}

      {loading ? (
        <div className="flex items-center justify-center py-16">
          <Loader2 className="h-6 w-6 animate-spin" />
        </div>
      ) : (
        <>
          {learningAnalytics && (
            <div className="grid gap-4 md:grid-cols-4">
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                    <Users className="h-4 w-4" />
                    Total Students
                  </CardTitle>
                  <div className="text-2xl font-bold">{learningAnalytics.total_students}</div>
                </CardHeader>
              </Card>
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                    <Target className="h-4 w-4" />
                    Average GPA
                  </CardTitle>
                  <div className="text-2xl font-bold">{learningAnalytics.average_gpa.toFixed(2)}</div>
                </CardHeader>
              </Card>
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                    <Award className="h-4 w-4" />
                    Retention Rate
                  </CardTitle>
                  <div className="text-2xl font-bold">{learningAnalytics.retention_rate.toFixed(1)}%</div>
                </CardHeader>
              </Card>
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                    <BookOpen className="h-4 w-4" />
                    Graduation Rate
                  </CardTitle>
                  <div className="text-2xl font-bold">{learningAnalytics.graduation_rate.toFixed(1)}%</div>
                </CardHeader>
              </Card>
            </div>
          )}

          <Tabs defaultValue="progression" className="space-y-4">
            <TabsList>
              <TabsTrigger value="progression" className="flex items-center gap-2">
                <Brain className="h-4 w-4" />
                Student Progression
              </TabsTrigger>
              <TabsTrigger value="engagement" className="flex items-center gap-2">
                <BookOpen className="h-4 w-4" />
                Course Engagement
              </TabsTrigger>
              <TabsTrigger value="insights" className="flex items-center gap-2" onClick={() => fetchPredictiveInsights()}>
                <AlertTriangle className="h-4 w-4" />
                Predictive Insights
              </TabsTrigger>
            </TabsList>

            <TabsContent value="progression" className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle>Student Progression Analysis</CardTitle>
                  <CardDescription>
                    Detailed analysis of individual student performance and progression
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="student-select">Select Student</Label>
                    <Select value={selectedStudent} onValueChange={handleStudentChange}>
                      <SelectTrigger id="student-select">
                        <SelectValue placeholder="Choose a student" />
                      </SelectTrigger>
                      <SelectContent>
                        {students.map((student) => (
                          <SelectItem key={student.id} value={student.id.toString()}>
                            {student.name || `${student.first_name} ${student.last_name}`} - {student.roll_no || student.rollNo}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  {loadingProgression ? (
                    <div className="flex items-center justify-center py-8">
                      <Loader2 className="h-6 w-6 animate-spin" />
                    </div>
                  ) : studentProgression ? (
                    <div className="space-y-6">
                      <div className="grid gap-4 md:grid-cols-3">
                        <Card>
                          <CardHeader className="pb-3">
                            <CardTitle className="text-sm font-medium text-muted-foreground">
                              Overall GPA
                            </CardTitle>
                            <div className="text-2xl font-bold">{studentProgression.overall_gpa.toFixed(2)}</div>
                          </CardHeader>
                        </Card>
                        <Card>
                          <CardHeader className="pb-3">
                            <CardTitle className="text-sm font-medium text-muted-foreground">
                              Attendance Rate
                            </CardTitle>
                            <div className="text-2xl font-bold">{studentProgression.attendance_rate.toFixed(1)}%</div>
                          </CardHeader>
                        </Card>
                        <Card>
                          <CardHeader className="pb-3">
                            <CardTitle className="text-sm font-medium text-muted-foreground">
                              Predicted GPA
                            </CardTitle>
                            <div className="text-2xl font-bold">{studentProgression.predicted_gpa.toFixed(2)}</div>
                          </CardHeader>
                        </Card>
                      </div>

                      <div className="grid gap-4 md:grid-cols-2">
                        <Card>
                          <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                              Performance Trend
                              {getTrendIcon(studentProgression.performance_trend)}
                            </CardTitle>
                          </CardHeader>
                          <CardContent>
                            <Badge className={
                              studentProgression.performance_trend === 'improving' ? 'bg-green-100 text-green-800' :
                              studentProgression.performance_trend === 'declining' ? 'bg-red-100 text-red-800' :
                              'bg-gray-100 text-gray-800'
                            }>
                              {studentProgression.performance_trend.toUpperCase()}
                            </Badge>
                          </CardContent>
                        </Card>

                        <Card>
                          <CardHeader>
                            <CardTitle>Courses</CardTitle>
                          </CardHeader>
                          <CardContent className="space-y-2">
                            <div className="flex justify-between">
                              <span className="text-muted-foreground">Completed:</span>
                              <span className="font-semibold">{studentProgression.courses_completed}</span>
                            </div>
                            <div className="flex justify-between">
                              <span className="text-muted-foreground">In Progress:</span>
                              <span className="font-semibold">{studentProgression.courses_in_progress}</span>
                            </div>
                          </CardContent>
                        </Card>
                      </div>

                      {studentProgression.at_risk && (
                        <Card className="border-red-200 bg-red-50 dark:bg-red-900/10">
                          <CardHeader>
                            <CardTitle className="flex items-center gap-2 text-red-800 dark:text-red-400">
                              <AlertTriangle className="h-5 w-5" />
                              At Risk Student
                            </CardTitle>
                          </CardHeader>
                          <CardContent>
                            <p className="text-sm text-red-700 dark:text-red-300">
                              This student requires immediate attention and support.
                            </p>
                          </CardContent>
                        </Card>
                      )}

                      <div className="grid gap-4 md:grid-cols-2">
                        <Card>
                          <CardHeader>
                            <CardTitle className="text-green-800 dark:text-green-400">Strengths</CardTitle>
                          </CardHeader>
                          <CardContent>
                            <ul className="list-disc list-inside space-y-1">
                              {studentProgression.strengths?.map((strength, idx) => (
                                <li key={idx} className="text-sm">{strength}</li>
                              ))}
                            </ul>
                          </CardContent>
                        </Card>

                        <Card>
                          <CardHeader>
                            <CardTitle className="text-yellow-800 dark:text-yellow-400">Weaknesses</CardTitle>
                          </CardHeader>
                          <CardContent>
                            <ul className="list-disc list-inside space-y-1">
                              {studentProgression.weaknesses?.map((weakness, idx) => (
                                <li key={idx} className="text-sm">{weakness}</li>
                              ))}
                            </ul>
                          </CardContent>
                        </Card>
                      </div>
                    </div>
                  ) : (
                    <div className="text-center py-8 text-muted-foreground">
                      Select a student to view progression analysis
                    </div>
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="engagement" className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle>Course Engagement Analysis</CardTitle>
                  <CardDescription>
                    Analyze student engagement and participation in courses
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="course-select">Select Course</Label>
                    <Select value={selectedCourse} onValueChange={handleCourseChange}>
                      <SelectTrigger id="course-select">
                        <SelectValue placeholder="Choose a course" />
                      </SelectTrigger>
                      <SelectContent>
                        {courses.map((course) => (
                          <SelectItem key={course.id} value={course.id.toString()}>
                            {course.code} - {course.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  {loadingEngagement ? (
                    <div className="flex items-center justify-center py-8">
                      <Loader2 className="h-6 w-6 animate-spin" />
                    </div>
                  ) : courseEngagement ? (
                    <div className="space-y-6">
                      <div className="grid gap-4 md:grid-cols-3">
                        <Card>
                          <CardHeader className="pb-3">
                            <CardTitle className="text-sm font-medium text-muted-foreground">
                              Total Students
                            </CardTitle>
                            <div className="text-2xl font-bold">{courseEngagement.total_students}</div>
                          </CardHeader>
                        </Card>
                        <Card>
                          <CardHeader className="pb-3">
                            <CardTitle className="text-sm font-medium text-muted-foreground">
                              Average Grade
                            </CardTitle>
                            <div className="text-2xl font-bold">{courseEngagement.average_grade.toFixed(1)}%</div>
                          </CardHeader>
                        </Card>
                        <Card>
                          <CardHeader className="pb-3">
                            <CardTitle className="text-sm font-medium text-muted-foreground">
                              Completion Rate
                            </CardTitle>
                            <div className="text-2xl font-bold">{courseEngagement.completion_rate.toFixed(1)}%</div>
                          </CardHeader>
                        </Card>
                      </div>

                      <Card>
                        <CardHeader>
                          <CardTitle>Engagement Metrics</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-4">
                          <div className="space-y-2">
                            <div className="flex justify-between text-sm">
                              <span>Engagement Rate</span>
                              <span className="font-medium">{courseEngagement.engagement_rate.toFixed(1)}%</span>
                            </div>
                            <Progress value={courseEngagement.engagement_rate} />
                          </div>
                          <div className="space-y-2">
                            <div className="flex justify-between text-sm">
                              <span>Attendance Rate</span>
                              <span className="font-medium">{courseEngagement.attendance_rate.toFixed(1)}%</span>
                            </div>
                            <Progress value={courseEngagement.attendance_rate} />
                          </div>
                          <div className="space-y-2">
                            <div className="flex justify-between text-sm">
                              <span>Assignment Submission Rate</span>
                              <span className="font-medium">{courseEngagement.assignment_submission_rate.toFixed(1)}%</span>
                            </div>
                            <Progress value={courseEngagement.assignment_submission_rate} />
                          </div>
                        </CardContent>
                      </Card>

                      <div className="grid gap-4 md:grid-cols-2">
                        <Card>
                          <CardHeader>
                            <CardTitle className="text-green-800 dark:text-green-400">
                              Most Engaged Students
                            </CardTitle>
                          </CardHeader>
                          <CardContent>
                            <ul className="space-y-2">
                              {courseEngagement.most_engaged_students?.map((student, idx) => (
                                <li key={idx} className="text-sm flex items-center gap-2">
                                  <Badge variant="outline">{idx + 1}</Badge>
                                  {student}
                                </li>
                              ))}
                            </ul>
                          </CardContent>
                        </Card>

                        <Card>
                          <CardHeader>
                            <CardTitle className="text-red-800 dark:text-red-400">
                              Needs Attention
                            </CardTitle>
                          </CardHeader>
                          <CardContent>
                            <ul className="space-y-2">
                              {courseEngagement.least_engaged_students?.map((student, idx) => (
                                <li key={idx} className="text-sm flex items-center gap-2">
                                  <Badge variant="outline">{idx + 1}</Badge>
                                  {student}
                                </li>
                              ))}
                            </ul>
                          </CardContent>
                        </Card>
                      </div>
                    </div>
                  ) : (
                    <div className="text-center py-8 text-muted-foreground">
                      Select a course to view engagement analysis
                    </div>
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="insights" className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle>Predictive Insights</CardTitle>
                  <CardDescription>
                    AI-powered predictions and recommendations
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  {loadingInsights ? (
                    <div className="flex items-center justify-center py-8">
                      <Loader2 className="h-6 w-6 animate-spin" />
                    </div>
                  ) : insights.length === 0 ? (
                    <div className="text-center py-8 text-muted-foreground">
                      Click the tab to load predictive insights
                    </div>
                  ) : (
                    <div className="space-y-4">
                      {insights.map((insight, idx) => (
                        <Card key={idx} className="border-l-4" style={{
                          borderLeftColor:
                            insight.type === 'at_risk' ? '#ef4444' :
                            insight.type === 'high_performer' ? '#22c55e' :
                            '#eab308'
                        }}>
                          <CardHeader>
                            <div className="flex items-center justify-between">
                              <CardTitle className="text-lg">{insight.student_name}</CardTitle>
                              {getInsightBadge(insight.type)}
                            </div>
                            <CardDescription>
                              Confidence: {(insight.confidence * 100).toFixed(0)}%
                            </CardDescription>
                          </CardHeader>
                          <CardContent className="space-y-3">
                            <div>
                              <h4 className="font-medium mb-2">Factors:</h4>
                              <ul className="list-disc list-inside space-y-1 text-sm text-muted-foreground">
                                {insight.factors?.map((factor, i) => (
                                  <li key={i}>{factor}</li>
                                ))}
                              </ul>
                            </div>
                            <div>
                              <h4 className="font-medium mb-2">Recommendations:</h4>
                              <ul className="list-disc list-inside space-y-1 text-sm text-muted-foreground">
                                {insight.recommendations?.map((rec, i) => (
                                  <li key={i}>{rec}</li>
                                ))}
                              </ul>
                            </div>
                          </CardContent>
                        </Card>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </>
      )}
    </div>
  );
}
