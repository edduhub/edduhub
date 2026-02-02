"use client";

import { useState, useEffect } from "react";
import { useAuth } from "@/lib/auth-context";
import { api } from "@/lib/api-client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { 
  Plus, 
  Search, 
  BookOpen, 
  Calendar, 
  Clock, 
  MapPin, 
  Users, 
  CheckCircle,
  Trophy,
  Download
} from "lucide-react";
import { format } from "date-fns";
import { logger } from "@/lib/logger";

interface Exam {
  id: number;
  title: string;
  description: string;
  exam_type: string;
  start_time: string;
  end_time: string;
  duration: number;
  total_marks: number;
  passing_marks: number;
  status: string;
  course_name?: string;
  room_number?: string;
}

export default function ExamsPage() {
  const { user } = useAuth();
  const [, setActiveTab] = useState("all-exams");
  const [searchQuery, setSearchQuery] = useState("");
  const [exams, setExams] = useState<Exam[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchExams = async () => {
      try {
        setLoading(true);
        const data = await api.get<Exam[]>("/api/exams");
        setExams(Array.isArray(data) ? data : []);
      } catch (err) {
        logger.error("Failed to fetch exams:", err as Error);
      } finally {
        setLoading(false);
      }
    };
    fetchExams();
  }, []);

  const filteredExams = exams.filter(exam => 
    exam.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
    exam.course_name?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="space-y-6 pb-10">
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="text-4xl font-extrabold tracking-tight lg:text-5xl bg-gradient-to-r from-primary to-blue-600 bg-clip-text text-transparent">
            Exams Portal
          </h1>
          <p className="text-muted-foreground mt-2 text-lg">
            Manage your academic evaluations, schedules, and results in one place.
          </p>
        </div>
        {(user?.role === 'admin' || user?.role === 'faculty') && (
          <Button size="lg" className="shadow-lg hover:shadow-primary/20 transition-all">
            <Plus className="mr-2 h-5 w-5" /> Schedule Exam
          </Button>
        )}
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card className="bg-gradient-to-br from-blue-50 to-white dark:from-slate-900 dark:to-slate-950 border-blue-100 dark:border-blue-900 border-2 shadow-sm">
          <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
            <CardTitle className="text-sm font-medium">Total Exams</CardTitle>
            <BookOpen className="h-4 w-4 text-blue-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{exams.length}</div>
            <p className="text-xs text-muted-foreground mt-1">Scheduled for this semester</p>
          </CardContent>
        </Card>
        <Card className="bg-gradient-to-br from-green-50 to-white dark:from-slate-900 dark:to-slate-950 border-green-100 dark:border-green-900 border-2 shadow-sm">
          <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
            <CardTitle className="text-sm font-medium">Ongoing</CardTitle>
            <Clock className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{exams.filter(e => e.status === 'ongoing').length}</div>
            <p className="text-xs text-muted-foreground mt-1">Currently in progress</p>
          </CardContent>
        </Card>
        <Card className="bg-gradient-to-br from-purple-50 to-white dark:from-slate-900 dark:to-slate-950 border-purple-100 dark:border-purple-900 border-2 shadow-sm">
          <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
            <CardTitle className="text-sm font-medium">Completed</CardTitle>
            <CheckCircle className="h-4 w-4 text-purple-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{exams.filter(e => e.status === 'completed').length}</div>
            <p className="text-xs text-muted-foreground mt-1">Results being processed</p>
          </CardContent>
        </Card>
        <Card className="bg-gradient-to-br from-orange-50 to-white dark:from-slate-900 dark:to-slate-950 border-orange-100 dark:border-orange-900 border-2 shadow-sm">
          <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
            <CardTitle className="text-sm font-medium">Attendance</CardTitle>
            <Users className="h-4 w-4 text-orange-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">94%</div>
            <p className="text-xs text-muted-foreground mt-1">Average participation rate</p>
          </CardContent>
        </Card>
      </div>

      <Card className="border-none shadow-xl bg-white/50 backdrop-blur-md dark:bg-slate-950/50">
        <CardHeader>
          <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
            <div>
              <CardTitle>Management Dashboard</CardTitle>
              <CardDescription>Filter and view exams details</CardDescription>
            </div>
            <div className="relative w-full md:w-72">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Search exams or courses..."
                className="pl-9 bg-white/80 dark:bg-slate-900/80"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="all-exams" onValueChange={setActiveTab}>
            <TabsList className="mb-6 bg-slate-100/80 dark:bg-slate-900/80 p-1">
              <TabsTrigger value="all-exams" className="data-[state=active]:bg-white dark:data-[state=active]:bg-slate-800 data-[state=active]:shadow-sm">
                All Exams
              </TabsTrigger>
              <TabsTrigger value="scheduled" className="data-[state=active]:bg-white dark:data-[state=active]:bg-slate-800 data-[state=active]:shadow-sm">
                Upcoming
              </TabsTrigger>
              <TabsTrigger value="completed" className="data-[state=active]:bg-white dark:data-[state=active]:bg-slate-800 data-[state=active]:shadow-sm">
                Past
              </TabsTrigger>
              {user?.role === 'student' && (
                <TabsTrigger value="results" className="data-[state=active]:bg-white dark:data-[state=active]:bg-slate-800 data-[state=active]:shadow-sm">
                  My Results
                </TabsTrigger>
              )}
            </TabsList>

            <TabsContent value="all-exams" className="space-y-4">
              <div className="rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden shadow-inner">
                <Table>
                  <TableHeader className="bg-slate-50/50 dark:bg-slate-900/50">
                    <TableRow>
                      <TableHead className="font-bold">Exam Title</TableHead>
                      <TableHead className="font-bold">Date & Time</TableHead>
                      <TableHead className="font-bold">Duration</TableHead>
                      <TableHead className="font-bold">Room</TableHead>
                      <TableHead className="font-bold">Status</TableHead>
                      <TableHead className="text-right font-bold">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredExams.map((exam) => (
                      <TableRow key={exam.id} className="hover:bg-slate-50/30 dark:hover:bg-slate-900/30 transition-colors">
                        <TableCell>
                          <div className="flex flex-col">
                            <span className="font-semibold text-slate-900 dark:text-slate-100">{exam.title}</span>
                            <span className="text-xs text-muted-foreground">{exam.course_name}</span>
                          </div>
                        </TableCell>
                        <TableCell className="text-sm font-medium">
                          <div className="flex items-center gap-2">
                            <Calendar className="h-3.5 w-3.5 text-slate-400" />
                            {format(new Date(exam.start_time), 'MMM dd, yyyy')}
                          </div>
                          <div className="flex items-center gap-2 text-xs text-muted-foreground mt-1">
                            <Clock className="h-3.5 w-3.5 text-slate-400" />
                            {format(new Date(exam.start_time), 'hh:mm a')}
                          </div>
                        </TableCell>
                        <TableCell className="text-sm">{exam.duration} mins</TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2 text-sm">
                            <MapPin className="h-3.5 w-3.5 text-slate-400" />
                            {exam.room_number || "TBD"}
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge 
                            variant={
                              exam.status === 'scheduled' ? 'default' : 
                              exam.status === 'ongoing' ? 'secondary' : 
                              exam.status === 'completed' ? 'outline' : 'destructive'
                            }
                            className={`
                              capitalize px-3 py-1 rounded-full text-[10px] tracking-wider font-bold
                              ${exam.status === 'ongoing' ? 'animate-pulse bg-green-500 text-white border-none' : ''}
                            `}
                          >
                            {exam.status}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right">
                          <Button variant="ghost" size="sm" className="hover:bg-primary/10 hover:text-primary transition-all">
                            View
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                    {filteredExams.length === 0 && !loading && (
                      <TableRow>
                        <TableCell colSpan={6} className="h-32 text-center text-muted-foreground italic">
                          No exams found matching your search.
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </div>
            </TabsContent>

            <TabsContent value="results">
              <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                {[1, 2, 3].map((i) => (
                  <Card key={i} className="group hover:scale-[1.02] transition-all duration-300 border-2 border-transparent hover:border-primary/20 cursor-pointer">
                    <CardHeader className="pb-2">
                      <div className="flex items-center justify-between mb-2">
                        <Badge className="bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400 border-none shadow-none">Grade: A</Badge>
                        <Trophy className="h-5 w-5 text-yellow-500" />
                      </div>
                      <CardTitle className="text-lg">Mid-Term Assessment</CardTitle>
                      <CardDescription>Advanced Computer Networks</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                      <div className="flex justify-between items-end border-b pb-4 border-slate-100 dark:border-slate-800">
                        <div>
                          <p className="text-xs text-muted-foreground uppercase tracking-widest font-bold">Marks Obtained</p>
                          <p className="text-3xl font-black text-slate-900 dark:text-slate-100 mt-1">85<span className="text-sm font-medium text-muted-foreground ml-1">/ 100</span></p>
                        </div>
                        <div className="text-right">
                          <p className="text-xs text-muted-foreground uppercase tracking-widest font-bold">Percentile</p>
                          <p className="text-xl font-bold text-slate-700 dark:text-slate-300 mt-1">92.4%</p>
                        </div>
                      </div>
                      <div className="flex items-center justify-between">
                        <div className="flex -space-x-2">
                           <div className="h-8 w-8 rounded-full border-2 border-white dark:border-slate-950 bg-slate-200 flex items-center justify-center text-[10px] font-bold">P</div>
                           <div className="h-8 w-8 rounded-full border-2 border-white dark:border-slate-950 bg-slate-300 flex items-center justify-center text-[10px] font-bold">S</div>
                        </div>
                        <Button variant="outline" size="sm" className="gap-2 group-hover:bg-primary group-hover:text-white transition-colors">
                          <Download className="h-3.5 w-3.5" /> Mark Sheet
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
}
