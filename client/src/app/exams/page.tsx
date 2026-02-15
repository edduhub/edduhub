"use client";

import { useState, useEffect, useCallback, useMemo } from "react";
import { useAuth } from "@/lib/auth-context";
import { api, endpoints } from "@/lib/api-client";
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
  Loader2,
  FileText,
  Check,
  X,
  Eye,
  Pencil,
  Trash2,
  Building,
  DoorOpen,
} from "lucide-react";
import { format } from "date-fns";
import { logger } from "@/lib/logger";

type ExamStatus = "scheduled" | "ongoing" | "completed" | "cancelled";

type ExamType = "midterm" | "final" | "quiz" | "practical";

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
  course_id?: number;
  course_name?: string;
  room_number?: string;
  instructions?: string;
  allowed_materials?: string;
  question_paper_sets?: number;
}

type CourseOption = {
  id: number;
  name?: string;
  title?: string;
};

type ExamResultApi = {
  exam_id?: number;
  examId?: number;
  marks_obtained?: number;
  marksObtained?: number;
  grade?: string;
  percentage?: number;
  result?: string;
};

type StudentResult = {
  examId: number;
  marksObtained?: number;
  grade?: string;
  percentage?: number;
  result?: string;
};

type RevaluationStatus = "pending" | "approved" | "rejected";

interface RevaluationRequest {
  revaluation_id: number;
  exam_id: number;
  student_id: number;
  request_reason: string;
  status: RevaluationStatus;
  request_date: string;
  exam_title?: string;
  student_name?: string;
  marks_obtained?: number;
  total_marks?: number;
}

type CreateExamForm = {
  courseId: string;
  title: string;
  description: string;
  examType: ExamType;
  startTime: string;
  endTime: string;
  duration: string;
  totalMarks: string;
  passingMarks: string;
  instructions: string;
  allowedMaterials: string;
  questionPaperSets: string;
};

interface ExamRoom {
  id: number;
  room_number: string;
  capacity: number;
  building: string;
  floor: string;
  is_available: boolean;
  created_at?: string;
  updated_at?: string;
}

type CreateExamRoomForm = {
  roomNumber: string;
  capacity: string;
  building: string;
  floor: string;
  isAvailable: boolean;
};

const initialCreateForm: CreateExamForm = {
  courseId: "",
  title: "",
  description: "",
  examType: "midterm",
  startTime: "",
  endTime: "",
  duration: "",
  totalMarks: "100",
  passingMarks: "40",
  instructions: "",
  allowedMaterials: "",
  questionPaperSets: "1",
};

const initialExamRoomForm: CreateExamRoomForm = {
  roomNumber: "",
  capacity: "30",
  building: "",
  floor: "1",
  isAvailable: true,
};

function normalizeStatus(exam: Exam): ExamStatus {
  const status = exam.status?.toLowerCase();
  if (status === "scheduled" || status === "ongoing" || status === "completed" || status === "cancelled") {
    return status;
  }

  const now = Date.now();
  const start = new Date(exam.start_time).getTime();
  const end = new Date(exam.end_time).getTime();
  if (Number.isFinite(start) && now < start) {
    return "scheduled";
  }
  if (Number.isFinite(start) && Number.isFinite(end) && now >= start && now <= end) {
    return "ongoing";
  }
  return "completed";
}

function safeFormat(dateValue: string, pattern: string): string {
  const parsed = new Date(dateValue);
  if (Number.isNaN(parsed.getTime())) {
    return "N/A";
  }
  return format(parsed, pattern);
}

export default function ExamsPage() {
  const { user } = useAuth();
  const [searchQuery, setSearchQuery] = useState("");
  const [exams, setExams] = useState<Exam[]>([]);
  const [courses, setCourses] = useState<CourseOption[]>([]);
  const [studentResults, setStudentResults] = useState<Record<number, StudentResult>>({});
  const [loading, setLoading] = useState(true);
  const [loadingResults, setLoadingResults] = useState(false);
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [createForm, setCreateForm] = useState<CreateExamForm>(initialCreateForm);
  const [createError, setCreateError] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [selectedExam, setSelectedExam] = useState<Exam | null>(null);
  const [revaluationRequests, setRevaluationRequests] = useState<RevaluationRequest[]>([]);
  const [loadingRevaluation, setLoadingRevaluation] = useState(false);
  const [isRevaluationDialogOpen, setIsRevaluationDialogOpen] = useState(false);
  const [revaluationForm, setRevaluationForm] = useState({
    examId: "",
    reason: "",
  });
  const [revaluationError, setRevaluationError] = useState<string | null>(null);
  const [isSubmittingRevaluation, setIsSubmittingRevaluation] = useState(false);
  const [selectedRequest, setSelectedRequest] = useState<RevaluationRequest | null>(null);

  // Exam Rooms state
  const [examRooms, setExamRooms] = useState<ExamRoom[]>([]);
  const [loadingRooms, setLoadingRooms] = useState(false);
  const [isCreateRoomDialogOpen, setIsCreateRoomDialogOpen] = useState(false);
  const [isEditRoomDialogOpen, setIsEditRoomDialogOpen] = useState(false);
  const [isAvailabilityDialogOpen, setIsAvailabilityDialogOpen] = useState(false);
  const [roomForm, setRoomForm] = useState<CreateExamRoomForm>(initialExamRoomForm);
  const [roomFormError, setRoomFormError] = useState<string | null>(null);
  const [isSavingRoom, setIsSavingRoom] = useState(false);
  const [selectedRoom, setSelectedRoom] = useState<ExamRoom | null>(null);
  const [roomAvailability, setRoomAvailability] = useState<{available: boolean; message?: string} | null>(null);
  const [checkingAvailability, setCheckingAvailability] = useState(false);

  const canManageExams = user?.role === "admin" || user?.role === "faculty";

  const fetchExams = useCallback(async () => {
    try {
      setLoading(true);
      const data = await api.get<Exam[]>(endpoints.exams.list);
      setExams(Array.isArray(data) ? data : []);
    } catch (err) {
      logger.error("Failed to fetch exams:", err as Error);
      setExams([]);
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchRevaluationRequests = useCallback(async () => {
    try {
      setLoadingRevaluation(true);
      const data = await api.get<RevaluationRequest[]>(endpoints.revaluation.list);
      setRevaluationRequests(Array.isArray(data) ? data : []);
    } catch (err) {
      logger.error("Failed to fetch revaluation requests:", err as Error);
      setRevaluationRequests([]);
    } finally {
      setLoadingRevaluation(false);
    }
  }, []);

  const fetchExamRooms = useCallback(async () => {
    try {
      setLoadingRooms(true);
      const data = await api.get<ExamRoom[]>(endpoints.examRooms.list);
      setExamRooms(Array.isArray(data) ? data : []);
    } catch (err) {
      logger.error("Failed to fetch exam rooms:", err as Error);
      setExamRooms([]);
    } finally {
      setLoadingRooms(false);
    }
  }, []);

  useEffect(() => {
    fetchExams();
  }, [fetchExams]);

  useEffect(() => {
    const shouldLoadRevaluation = user?.role === "student" || canManageExams;
    if (!shouldLoadRevaluation) {
      setRevaluationRequests([]);
      return;
    }
    fetchRevaluationRequests();
  }, [user?.role, canManageExams, fetchRevaluationRequests]);

  useEffect(() => {
    if (!canManageExams) {
      setExamRooms([]);
      return;
    }
    fetchExamRooms();
  }, [canManageExams, fetchExamRooms]);

  useEffect(() => {
    if (!canManageExams) {
      setCourses([]);
      return;
    }

    let mounted = true;
    const fetchCourses = async () => {
      try {
        const data = await api.get<CourseOption[]>(endpoints.courses.list);
        if (!mounted) return;
        setCourses(Array.isArray(data) ? data : []);
      } catch (err) {
        logger.error("Failed to fetch courses for exam scheduling:", err as Error);
        if (mounted) {
          setCourses([]);
        }
      }
    };

    fetchCourses();
    return () => {
      mounted = false;
    };
  }, [canManageExams]);

  useEffect(() => {
    const fetchStudentResults = async () => {
      if (user?.role !== "student") {
        setStudentResults({});
        return;
      }

      const studentId = Number.parseInt(String(user.id ?? ""), 10);
      if (!Number.isFinite(studentId) || studentId <= 0) {
        setStudentResults({});
        return;
      }

      try {
        setLoadingResults(true);
        const data = await api.get<ExamResultApi[]>(endpoints.exams.studentResults(studentId));
        const resultsArray = Array.isArray(data) ? data : [];
        const mapped: Record<number, StudentResult> = {};

        for (const item of resultsArray) {
          const examId = item.exam_id ?? item.examId;
          if (typeof examId !== "number" || examId <= 0) continue;
          mapped[examId] = {
            examId,
            marksObtained: item.marks_obtained ?? item.marksObtained,
            grade: item.grade,
            percentage: item.percentage,
            result: item.result,
          };
        }

        setStudentResults(mapped);
      } catch (err) {
        logger.error("Failed to fetch student exam results:", err as Error);
        setStudentResults({});
      } finally {
        setLoadingResults(false);
      }
    };

    fetchStudentResults();
  }, [user?.id, user?.role]);

  const handleCreateExam = async () => {
    setCreateError(null);

    const courseId = Number.parseInt(createForm.courseId, 10);
    const totalMarks = Number.parseFloat(createForm.totalMarks);
    const passingMarks = Number.parseFloat(createForm.passingMarks);
    const questionPaperSets = Number.parseInt(createForm.questionPaperSets, 10);
    const start = new Date(createForm.startTime);
    const end = new Date(createForm.endTime);

    let duration = Number.parseInt(createForm.duration, 10);
    if (!Number.isFinite(duration) || duration <= 0) {
      duration = Math.round((end.getTime() - start.getTime()) / 60000);
    }

    if (!Number.isFinite(courseId) || courseId <= 0) {
      setCreateError("Please select a course.");
      return;
    }
    if (!createForm.title.trim()) {
      setCreateError("Exam title is required.");
      return;
    }
    if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) {
      setCreateError("Start and end time are required.");
      return;
    }
    if (end <= start) {
      setCreateError("End time must be later than start time.");
      return;
    }
    if (!Number.isFinite(duration) || duration <= 0) {
      setCreateError("Duration must be greater than 0 minutes.");
      return;
    }
    if (!Number.isFinite(totalMarks) || totalMarks <= 0) {
      setCreateError("Total marks must be greater than 0.");
      return;
    }
    if (!Number.isFinite(passingMarks) || passingMarks < 0 || passingMarks > totalMarks) {
      setCreateError("Passing marks must be between 0 and total marks.");
      return;
    }
    if (!Number.isFinite(questionPaperSets) || questionPaperSets <= 0) {
      setCreateError("Question paper sets must be at least 1.");
      return;
    }

    setIsCreating(true);
    try {
      await api.post(endpoints.exams.create, {
        course_id: courseId,
        title: createForm.title.trim(),
        description: createForm.description.trim(),
        exam_type: createForm.examType,
        start_time: start.toISOString(),
        end_time: end.toISOString(),
        duration,
        total_marks: totalMarks,
        passing_marks: passingMarks,
        instructions: createForm.instructions.trim(),
        allowed_materials: createForm.allowedMaterials.trim(),
        question_paper_sets: questionPaperSets,
      });

      setCreateForm(initialCreateForm);
      setIsCreateDialogOpen(false);
      await fetchExams();
    } catch (err) {
      logger.error("Failed to create exam:", err as Error);
      setCreateError("Failed to schedule exam. Please verify details and try again.");
    } finally {
      setIsCreating(false);
    }
  };

  const handleSubmitRevaluation = async () => {
    setRevaluationError(null);

    const examId = Number.parseInt(revaluationForm.examId, 10);
    if (!Number.isFinite(examId) || examId <= 0) {
      setRevaluationError("Please select an exam.");
      return;
    }
    if (!revaluationForm.reason.trim()) {
      setRevaluationError("Please provide a reason for revaluation.");
      return;
    }
    if (revaluationForm.reason.trim().length < 20) {
      setRevaluationError("Reason must be at least 20 characters long.");
      return;
    }

    setIsSubmittingRevaluation(true);
    try {
      await api.post(endpoints.revaluation.create, {
        exam_id: examId,
        request_reason: revaluationForm.reason.trim(),
      });

      setRevaluationForm({ examId: "", reason: "" });
      setIsRevaluationDialogOpen(false);
      await fetchRevaluationRequests();
    } catch (err) {
      logger.error("Failed to submit revaluation request:", err as Error);
      setRevaluationError("Failed to submit revaluation request. Please try again.");
    } finally {
      setIsSubmittingRevaluation(false);
    }
  };

  const handleApproveRevaluation = async (requestId: number) => {
    try {
      await api.patch(endpoints.revaluation.approve(requestId), {});
      await fetchRevaluationRequests();
    } catch (err) {
      logger.error("Failed to approve revaluation request:", err as Error);
    }
  };

  const handleRejectRevaluation = async (requestId: number) => {
    try {
      await api.patch(endpoints.revaluation.reject(requestId), {});
      await fetchRevaluationRequests();
    } catch (err) {
      logger.error("Failed to reject revaluation request:", err as Error);
    }
  };

  const handleCreateRoom = async () => {
    setRoomFormError(null);
    const capacity = Number.parseInt(roomForm.capacity, 10);

    if (!roomForm.roomNumber.trim()) {
      setRoomFormError("Room number is required.");
      return;
    }
    if (!Number.isFinite(capacity) || capacity <= 0) {
      setRoomFormError("Capacity must be a positive number.");
      return;
    }
    if (!roomForm.building.trim()) {
      setRoomFormError("Building is required.");
      return;
    }

    setIsSavingRoom(true);
    try {
      await api.post(endpoints.examRooms.create, {
        room_number: roomForm.roomNumber.trim(),
        capacity,
        building: roomForm.building.trim(),
        floor: roomForm.floor.trim() || "1",
        is_available: roomForm.isAvailable,
      });

      setRoomForm(initialExamRoomForm);
      setIsCreateRoomDialogOpen(false);
      await fetchExamRooms();
    } catch (err) {
      logger.error("Failed to create exam room:", err as Error);
      setRoomFormError("Failed to create exam room. Please try again.");
    } finally {
      setIsSavingRoom(false);
    }
  };

  const handleUpdateRoom = async () => {
    if (!selectedRoom) return;
    setRoomFormError(null);
    const capacity = Number.parseInt(roomForm.capacity, 10);

    if (!roomForm.roomNumber.trim()) {
      setRoomFormError("Room number is required.");
      return;
    }
    if (!Number.isFinite(capacity) || capacity <= 0) {
      setRoomFormError("Capacity must be a positive number.");
      return;
    }
    if (!roomForm.building.trim()) {
      setRoomFormError("Building is required.");
      return;
    }

    setIsSavingRoom(true);
    try {
      await api.put(endpoints.examRooms.update(selectedRoom.id), {
        room_number: roomForm.roomNumber.trim(),
        capacity,
        building: roomForm.building.trim(),
        floor: roomForm.floor.trim() || "1",
        is_available: roomForm.isAvailable,
      });

      setRoomForm(initialExamRoomForm);
      setSelectedRoom(null);
      setIsEditRoomDialogOpen(false);
      await fetchExamRooms();
    } catch (err) {
      logger.error("Failed to update exam room:", err as Error);
      setRoomFormError("Failed to update exam room. Please try again.");
    } finally {
      setIsSavingRoom(false);
    }
  };

  const handleDeleteRoom = async (roomId: number) => {
    if (!window.confirm("Are you sure you want to delete this exam room?")) return;

    try {
      await api.delete(endpoints.examRooms.delete(roomId));
      await fetchExamRooms();
    } catch (err) {
      logger.error("Failed to delete exam room:", err as Error);
    }
  };

  const handleCheckAvailability = async (roomId: number) => {
    setCheckingAvailability(true);
    setRoomAvailability(null);
    setSelectedRoom(examRooms.find(r => r.id === roomId) || null);
    setIsAvailabilityDialogOpen(true);

    try {
      const data = await api.get<{available: boolean; message?: string}>(endpoints.examRooms.availability(roomId));
      setRoomAvailability(data);
    } catch (err) {
      logger.error("Failed to check room availability:", err as Error);
      setRoomAvailability({ available: false, message: "Failed to check availability" });
    } finally {
      setCheckingAvailability(false);
    }
  };

  const openEditRoomDialog = (room: ExamRoom) => {
    setSelectedRoom(room);
    setRoomForm({
      roomNumber: room.room_number,
      capacity: String(room.capacity),
      building: room.building,
      floor: room.floor,
      isAvailable: room.is_available,
    });
    setRoomFormError(null);
    setIsEditRoomDialogOpen(true);
  };

  const allExams = useMemo(() => {
    const query = searchQuery.trim().toLowerCase();
    if (!query) return exams;
    return exams.filter((exam) =>
      exam.title.toLowerCase().includes(query) ||
      (exam.course_name || "").toLowerCase().includes(query)
    );
  }, [exams, searchQuery]);

  const scheduledExams = allExams.filter((exam) => normalizeStatus(exam) === "scheduled");
  const completedExams = allExams.filter((exam) => normalizeStatus(exam) === "completed");

  const getStatusBadge = (status: ExamStatus) => {
    const variant =
      status === "scheduled" ? "default" :
      status === "ongoing" ? "secondary" :
      status === "completed" ? "outline" :
      "destructive";

    return (
      <Badge
        variant={variant}
        className={[
          "capitalize px-3 py-1 rounded-full text-[10px] tracking-wider font-bold",
          status === "ongoing" ? "animate-pulse bg-green-500 text-white border-none" : "",
        ].join(" ")}
      >
        {status}
      </Badge>
    );
  };

  const renderExamTable = (rows: Exam[], emptyText: string) => (
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
          {loading && rows.length === 0 && (
            <TableRow>
              <TableCell colSpan={6} className="h-32 text-center">
                <div className="inline-flex items-center gap-2 text-muted-foreground">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Loading exams...
                </div>
              </TableCell>
            </TableRow>
          )}

          {rows.map((exam) => {
            const status = normalizeStatus(exam);
            return (
              <TableRow key={exam.id} className="hover:bg-slate-50/30 dark:hover:bg-slate-900/30 transition-colors">
                <TableCell>
                  <div className="flex flex-col">
                    <span className="font-semibold text-slate-900 dark:text-slate-100">{exam.title}</span>
                    <span className="text-xs text-muted-foreground">{exam.course_name || `Course ${exam.course_id ?? exam.id}`}</span>
                  </div>
                </TableCell>
                <TableCell className="text-sm font-medium">
                  <div className="flex items-center gap-2">
                    <Calendar className="h-3.5 w-3.5 text-slate-400" />
                    {safeFormat(exam.start_time, "MMM dd, yyyy")}
                  </div>
                  <div className="flex items-center gap-2 text-xs text-muted-foreground mt-1">
                    <Clock className="h-3.5 w-3.5 text-slate-400" />
                    {safeFormat(exam.start_time, "hh:mm a")}
                  </div>
                </TableCell>
                <TableCell className="text-sm">{exam.duration} mins</TableCell>
                <TableCell>
                  <div className="flex items-center gap-2 text-sm">
                    <MapPin className="h-3.5 w-3.5 text-slate-400" />
                    {exam.room_number || "TBD"}
                  </div>
                </TableCell>
                <TableCell>{getStatusBadge(status)}</TableCell>
                <TableCell className="text-right">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="hover:bg-primary/10 hover:text-primary transition-all"
                    onClick={() => setSelectedExam(exam)}
                  >
                    View
                  </Button>
                </TableCell>
              </TableRow>
            );
          })}

          {rows.length === 0 && !loading && (
            <TableRow>
              <TableCell colSpan={6} className="h-32 text-center text-muted-foreground italic">
                {emptyText}
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  );

  return (
    <>
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
          {canManageExams && (
            <Button
              size="lg"
              className="shadow-lg hover:shadow-primary/20 transition-all"
              onClick={() => {
                setCreateError(null);
                setIsCreateDialogOpen(true);
              }}
            >
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
              <div className="text-2xl font-bold">{exams.filter((exam) => normalizeStatus(exam) === "ongoing").length}</div>
              <p className="text-xs text-muted-foreground mt-1">Currently in progress</p>
            </CardContent>
          </Card>
          <Card className="bg-gradient-to-br from-purple-50 to-white dark:from-slate-900 dark:to-slate-950 border-purple-100 dark:border-purple-900 border-2 shadow-sm">
            <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
              <CardTitle className="text-sm font-medium">Completed</CardTitle>
              <CheckCircle className="h-4 w-4 text-purple-600" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{exams.filter((exam) => normalizeStatus(exam) === "completed").length}</div>
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
            <Tabs defaultValue="all-exams">
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
                {user?.role === "student" && (
                  <TabsTrigger value="results" className="data-[state=active]:bg-white dark:data-[state=active]:bg-slate-800 data-[state=active]:shadow-sm">
                    My Results
                  </TabsTrigger>
                )}
                <TabsTrigger value="revaluation" className="data-[state=active]:bg-white dark:data-[state=active]:bg-slate-800 data-[state=active]:shadow-sm">
                  Revaluation Requests
                </TabsTrigger>
                {canManageExams && (
                  <TabsTrigger value="exam-rooms" className="data-[state=active]:bg-white dark:data-[state=active]:bg-slate-800 data-[state=active]:shadow-sm">
                    Exam Rooms
                  </TabsTrigger>
                )}
              </TabsList>

              <TabsContent value="all-exams" className="space-y-4">
                {renderExamTable(allExams, "No exams found matching your search.")}
              </TabsContent>

              <TabsContent value="scheduled" className="space-y-4">
                {renderExamTable(scheduledExams, "No upcoming exams found.")}
              </TabsContent>

              <TabsContent value="completed" className="space-y-4">
                {renderExamTable(completedExams, "No past exams found.")}
              </TabsContent>

              {user?.role === "student" && (
                <TabsContent value="results">
                  {loadingResults ? (
                    <div className="flex items-center justify-center py-12">
                      <Loader2 className="h-6 w-6 animate-spin" />
                    </div>
                  ) : (
                    <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                      {completedExams.map((exam) => {
                        const result = studentResults[exam.id];
                        return (
                          <Card key={exam.id} className="group hover:scale-[1.02] transition-all duration-300 border-2 border-transparent hover:border-primary/20 cursor-pointer">
                            <CardHeader className="pb-2">
                              <div className="flex items-center justify-between mb-2">
                                <Badge className="bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400 border-none shadow-none">
                                  Grade: {result?.grade || "Pending"}
                                </Badge>
                                <Trophy className="h-5 w-5 text-yellow-500" />
                              </div>
                              <CardTitle className="text-lg">{exam.title}</CardTitle>
                              <CardDescription>{exam.course_name || `Course ${exam.course_id ?? exam.id}`}</CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-4">
                              <div className="flex justify-between items-end border-b pb-4 border-slate-100 dark:border-slate-800">
                                <div>
                                  <p className="text-xs text-muted-foreground uppercase tracking-widest font-bold">Marks Obtained</p>
                                  <p className="text-3xl font-black text-slate-900 dark:text-slate-100 mt-1">
                                    {typeof result?.marksObtained === "number" ? result.marksObtained : "--"}
                                    <span className="text-sm font-medium text-muted-foreground ml-1">/ {exam.total_marks}</span>
                                  </p>
                                </div>
                                <div className="text-right">
                                  <p className="text-xs text-muted-foreground uppercase tracking-widest font-bold">Percentage</p>
                                  <p className="text-xl font-bold text-slate-700 dark:text-slate-300 mt-1">
                                    {typeof result?.percentage === "number" ? `${result.percentage.toFixed(1)}%` : "--"}
                                  </p>
                                </div>
                              </div>
                              <div className="text-xs text-muted-foreground">
                                {result?.result ? `Result: ${result.result}` : "Result pending publication"}
                              </div>
                            </CardContent>
                          </Card>
                        );
                      })}
                      {completedExams.length === 0 && (
                        <div className="text-sm text-muted-foreground">No completed exams yet.</div>
                      )}
                    </div>
                  )}
                </TabsContent>
              )}

              <TabsContent value="revaluation">
                <div className="flex justify-between items-center mb-4">
                  <p className="text-sm text-muted-foreground">
                    {user?.role === "student" 
                      ? "Submit revaluation requests for exams you've completed."
                      : "Manage revaluation requests from students."}
                  </p>
                  {user?.role === "student" && (
                    <Button
                      size="sm"
                      onClick={() => {
                        setRevaluationError(null);
                        setRevaluationForm({ examId: "", reason: "" });
                        setIsRevaluationDialogOpen(true);
                      }}
                    >
                      <Plus className="mr-2 h-4 w-4" /> Request Revaluation
                    </Button>
                  )}
                </div>

                {loadingRevaluation ? (
                  <div className="flex items-center justify-center py-12">
                    <Loader2 className="h-6 w-6 animate-spin" />
                  </div>
                ) : revaluationRequests.length === 0 ? (
                  <div className="text-center py-12 text-muted-foreground">
                    <FileText className="h-12 w-12 mx-auto mb-4 opacity-50" />
                    <p>No revaluation requests found.</p>
                  </div>
                ) : (
                  <div className="rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden shadow-inner">
                    <Table>
                      <TableHeader className="bg-slate-50/50 dark:bg-slate-900/50">
                        <TableRow>
                          {canManageExams && <TableHead className="font-bold">Student</TableHead>}
                          <TableHead className="font-bold">Exam</TableHead>
                          <TableHead className="font-bold">Marks</TableHead>
                          <TableHead className="font-bold">Reason</TableHead>
                          <TableHead className="font-bold">Date</TableHead>
                          <TableHead className="font-bold">Status</TableHead>
                          {canManageExams && <TableHead className="text-right font-bold">Actions</TableHead>}
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {revaluationRequests.map((request) => (
                          <TableRow key={request.revaluation_id} className="hover:bg-slate-50/30 dark:hover:bg-slate-900/30 transition-colors">
                            {canManageExams && (
                              <TableCell className="font-medium">
                                {request.student_name || `Student #${request.student_id}`}
                              </TableCell>
                            )}
                            <TableCell>
                              <div className="flex flex-col">
                                <span className="font-semibold">{request.exam_title || `Exam #${request.exam_id}`}</span>
                              </div>
                            </TableCell>
                            <TableCell>
                              {typeof request.marks_obtained === "number" 
                                ? `${request.marks_obtained} / ${request.total_marks || "?"}` 
                                : "--"}
                            </TableCell>
                            <TableCell className="max-w-xs">
                              <p className="truncate text-sm" title={request.request_reason}>
                                {request.request_reason}
                              </p>
                            </TableCell>
                            <TableCell className="text-sm">
                              {safeFormat(request.request_date, "MMM dd, yyyy")}
                            </TableCell>
                            <TableCell>
                              <Badge
                                variant={
                                  request.status === "approved" ? "default" :
                                  request.status === "rejected" ? "destructive" :
                                  "secondary"
                                }
                                className={[
                                  "capitalize px-3 py-1 rounded-full text-[10px] tracking-wider font-bold",
                                  request.status === "pending" ? "bg-yellow-500 text-white border-none animate-pulse" : "",
                                  request.status === "approved" ? "bg-green-500 text-white border-none" : "",
                                  request.status === "rejected" ? "bg-red-500 text-white border-none" : "",
                                ].join(" ")}
                              >
                                {request.status}
                              </Badge>
                            </TableCell>
                            {canManageExams && (
                              <TableCell className="text-right">
                                {request.status === "pending" && (
                                  <div className="flex justify-end gap-2">
                                    <Button
                                      variant="ghost"
                                      size="sm"
                                      className="hover:bg-green-100 hover:text-green-700 dark:hover:bg-green-900/30"
                                      onClick={() => handleApproveRevaluation(request.revaluation_id)}
                                    >
                                      <Check className="h-4 w-4" />
                                    </Button>
                                    <Button
                                      variant="ghost"
                                      size="sm"
                                      className="hover:bg-red-100 hover:text-red-700 dark:hover:bg-red-900/30"
                                      onClick={() => handleRejectRevaluation(request.revaluation_id)}
                                    >
                                      <X className="h-4 w-4" />
                                    </Button>
                                  </div>
                                )}
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => setSelectedRequest(request)}
                                >
                                  <Eye className="h-4 w-4" />
                                </Button>
                              </TableCell>
                            )}
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </div>
                )}
              </TabsContent>

              {canManageExams && (
                <TabsContent value="exam-rooms" className="space-y-4">
                  <div className="flex justify-between items-center mb-4">
                    <p className="text-sm text-muted-foreground">
                      Manage exam rooms for scheduling exams.
                    </p>
                    <Button
                      size="sm"
                      onClick={() => {
                        setRoomFormError(null);
                        setRoomForm(initialExamRoomForm);
                        setIsCreateRoomDialogOpen(true);
                      }}
                    >
                      <Plus className="mr-2 h-4 w-4" /> Add Room
                    </Button>
                  </div>

                  {loadingRooms ? (
                    <div className="flex items-center justify-center py-12">
                      <Loader2 className="h-6 w-6 animate-spin" />
                    </div>
                  ) : examRooms.length === 0 ? (
                    <div className="text-center py-12 text-muted-foreground">
                      <DoorOpen className="h-12 w-12 mx-auto mb-4 opacity-50" />
                      <p>No exam rooms found.</p>
                      <Button
                        variant="ghost"
                        className="mt-2 text-primary hover:text-primary hover:underline"
                        onClick={() => {
                          setRoomFormError(null);
                          setRoomForm(initialExamRoomForm);
                          setIsCreateRoomDialogOpen(true);
                        }}
                      >
                        Add the first room
                      </Button>
                    </div>
                  ) : (
                    <div className="rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden shadow-inner">
                      <Table>
                        <TableHeader className="bg-slate-50/50 dark:bg-slate-900/50">
                          <TableRow>
                            <TableHead className="font-bold">Room</TableHead>
                            <TableHead className="font-bold">Capacity</TableHead>
                            <TableHead className="font-bold">Building</TableHead>
                            <TableHead className="font-bold">Floor</TableHead>
                            <TableHead className="font-bold">Status</TableHead>
                            <TableHead className="text-right font-bold">Actions</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {examRooms.map((room) => (
                            <TableRow key={room.id} className="hover:bg-slate-50/30 dark:hover:bg-slate-900/30 transition-colors">
                              <TableCell>
                                <div className="flex items-center gap-2">
                                  <DoorOpen className="h-4 w-4 text-muted-foreground" />
                                  <span className="font-semibold">{room.room_number}</span>
                                </div>
                              </TableCell>
                              <TableCell>
                                <div className="flex items-center gap-2">
                                  <Users className="h-4 w-4 text-muted-foreground" />
                                  {room.capacity}
                                </div>
                              </TableCell>
                              <TableCell>
                                <div className="flex items-center gap-2">
                                  <Building className="h-4 w-4 text-muted-foreground" />
                                  {room.building}
                                </div>
                              </TableCell>
                              <TableCell>Floor {room.floor}</TableCell>
                              <TableCell>
                                <Badge
                                  variant={room.is_available ? "default" : "secondary"}
                                  className={[
                                    "capitalize px-3 py-1 rounded-full text-[10px] tracking-wider font-bold",
                                    room.is_available ? "bg-green-500 text-white border-none" : "bg-slate-500 text-white border-none",
                                  ].join(" ")}
                                >
                                  {room.is_available ? "Available" : "Unavailable"}
                                </Badge>
                              </TableCell>
                              <TableCell className="text-right">
                                <div className="flex justify-end gap-2">
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    className="hover:bg-blue-100 hover:text-blue-700 dark:hover:bg-blue-900/30"
                                    onClick={() => handleCheckAvailability(room.id)}
                                    title="Check Availability"
                                  >
                                    <Calendar className="h-4 w-4" />
                                  </Button>
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    className="hover:bg-primary/10 hover:text-primary transition-all"
                                    onClick={() => openEditRoomDialog(room)}
                                  >
                                    <Pencil className="h-4 w-4" />
                                  </Button>
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    className="hover:bg-red-100 hover:text-red-700 dark:hover:bg-red-900/30"
                                    onClick={() => handleDeleteRoom(room.id)}
                                  >
                                    <Trash2 className="h-4 w-4" />
                                  </Button>
                                </div>
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </div>
                  )}
                </TabsContent>
              )}
            </Tabs>
          </CardContent>
        </Card>
      </div>

      {canManageExams && isCreateDialogOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 p-4 flex items-center justify-center">
          <Card className="w-full max-w-3xl max-h-[90vh] overflow-y-auto">
            <CardHeader>
              <div className="flex items-center justify-between gap-4">
                <CardTitle>Schedule New Exam</CardTitle>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    setIsCreateDialogOpen(false);
                    setCreateError(null);
                  }}
                >
                  Close
                </Button>
              </div>
              <CardDescription>Create an exam with timings, marks, and instructions.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <label className="text-sm font-medium" htmlFor="examCourse">Course</label>
                  <select
                    id="examCourse"
                    value={createForm.courseId}
                    onChange={(event) => setCreateForm((prev) => ({ ...prev, courseId: event.target.value }))}
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  >
                    <option value="">Select course</option>
                    {courses.map((course) => (
                      <option key={course.id} value={course.id}>
                        {course.name || course.title || `Course ${course.id}`}
                      </option>
                    ))}
                  </select>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium" htmlFor="examType">Exam Type</label>
                  <select
                    id="examType"
                    value={createForm.examType}
                    onChange={(event) => setCreateForm((prev) => ({ ...prev, examType: event.target.value as ExamType }))}
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  >
                    <option value="midterm">Midterm</option>
                    <option value="final">Final</option>
                    <option value="quiz">Quiz</option>
                    <option value="practical">Practical</option>
                  </select>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium" htmlFor="examTitle">Title</label>
                <Input
                  id="examTitle"
                  value={createForm.title}
                  onChange={(event) => setCreateForm((prev) => ({ ...prev, title: event.target.value }))}
                  placeholder="e.g., Data Structures Midterm"
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium" htmlFor="examDescription">Description</label>
                <textarea
                  id="examDescription"
                  value={createForm.description}
                  onChange={(event) => setCreateForm((prev) => ({ ...prev, description: event.target.value }))}
                  className="flex min-h-[90px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  placeholder="Optional exam description"
                />
              </div>

              <div className="grid gap-4 md:grid-cols-3">
                <div className="space-y-2">
                  <label className="text-sm font-medium" htmlFor="examStart">Start Time</label>
                  <Input
                    id="examStart"
                    type="datetime-local"
                    value={createForm.startTime}
                    onChange={(event) => setCreateForm((prev) => ({ ...prev, startTime: event.target.value }))}
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium" htmlFor="examEnd">End Time</label>
                  <Input
                    id="examEnd"
                    type="datetime-local"
                    value={createForm.endTime}
                    onChange={(event) => setCreateForm((prev) => ({ ...prev, endTime: event.target.value }))}
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium" htmlFor="examDuration">Duration (minutes)</label>
                  <Input
                    id="examDuration"
                    type="number"
                    min={1}
                    value={createForm.duration}
                    onChange={(event) => setCreateForm((prev) => ({ ...prev, duration: event.target.value }))}
                    placeholder="Auto-calculated if blank"
                  />
                </div>
              </div>

              <div className="grid gap-4 md:grid-cols-3">
                <div className="space-y-2">
                  <label className="text-sm font-medium" htmlFor="examTotalMarks">Total Marks</label>
                  <Input
                    id="examTotalMarks"
                    type="number"
                    min={1}
                    value={createForm.totalMarks}
                    onChange={(event) => setCreateForm((prev) => ({ ...prev, totalMarks: event.target.value }))}
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium" htmlFor="examPassingMarks">Passing Marks</label>
                  <Input
                    id="examPassingMarks"
                    type="number"
                    min={0}
                    value={createForm.passingMarks}
                    onChange={(event) => setCreateForm((prev) => ({ ...prev, passingMarks: event.target.value }))}
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium" htmlFor="examPaperSets">Question Paper Sets</label>
                  <Input
                    id="examPaperSets"
                    type="number"
                    min={1}
                    value={createForm.questionPaperSets}
                    onChange={(event) => setCreateForm((prev) => ({ ...prev, questionPaperSets: event.target.value }))}
                  />
                </div>
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <label className="text-sm font-medium" htmlFor="examInstructions">Instructions</label>
                  <textarea
                    id="examInstructions"
                    value={createForm.instructions}
                    onChange={(event) => setCreateForm((prev) => ({ ...prev, instructions: event.target.value }))}
                    className="flex min-h-[90px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                    placeholder="Any instructions for students"
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium" htmlFor="examMaterials">Allowed Materials</label>
                  <textarea
                    id="examMaterials"
                    value={createForm.allowedMaterials}
                    onChange={(event) => setCreateForm((prev) => ({ ...prev, allowedMaterials: event.target.value }))}
                    className="flex min-h-[90px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                    placeholder="Calculator, notes, etc."
                  />
                </div>
              </div>

              {createError && <p className="text-sm text-destructive">{createError}</p>}

              <div className="flex justify-end gap-2">
                <Button
                  variant="outline"
                  onClick={() => {
                    setIsCreateDialogOpen(false);
                    setCreateError(null);
                  }}
                  disabled={isCreating}
                >
                  Cancel
                </Button>
                <Button onClick={handleCreateExam} disabled={isCreating}>
                  {isCreating ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Scheduling...
                    </>
                  ) : (
                    <>
                      <Plus className="mr-2 h-4 w-4" />
                      Schedule Exam
                    </>
                  )}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {user?.role === "student" && isRevaluationDialogOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 p-4 flex items-center justify-center">
          <Card className="w-full max-w-lg">
            <CardHeader>
              <div className="flex items-center justify-between gap-4">
                <CardTitle>Request Exam Revaluation</CardTitle>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    setIsRevaluationDialogOpen(false);
                    setRevaluationError(null);
                  }}
                >
                  Close
                </Button>
              </div>
              <CardDescription>
                Submit a request for revaluation (rechecking) of your exam answer script.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium" htmlFor="revalExam">Select Exam</label>
                <select
                  id="revalExam"
                  value={revaluationForm.examId}
                  onChange={(event) => setRevaluationForm((prev) => ({ ...prev, examId: event.target.value }))}
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                >
                  <option value="">Select an exam</option>
                  {completedExams.map((exam) => {
                    const hasResult = studentResults[exam.id]?.marksObtained !== undefined;
                    return (
                      <option key={exam.id} value={exam.id} disabled={!hasResult}>
                        {exam.title} {!hasResult ? "(No result yet)" : ""}
                      </option>
                    );
                  })}
                </select>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium" htmlFor="revalReason">Reason for Revaluation</label>
                <textarea
                  id="revalReason"
                  value={revaluationForm.reason}
                  onChange={(event) => setRevaluationForm((prev) => ({ ...prev, reason: event.target.value }))}
                  className="flex min-h-[120px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  placeholder="Please explain why you believe your exam should be revalued. Provide specific details about any marking discrepancies you've noticed (minimum 20 characters)."
                />
                <p className="text-xs text-muted-foreground">
                  {revaluationForm.reason.length}/20 characters minimum
                </p>
              </div>

              {revaluationError && <p className="text-sm text-destructive">{revaluationError}</p>}

              <div className="flex justify-end gap-2">
                <Button
                  variant="outline"
                  onClick={() => {
                    setIsRevaluationDialogOpen(false);
                    setRevaluationError(null);
                  }}
                  disabled={isSubmittingRevaluation}
                >
                  Cancel
                </Button>
                <Button onClick={handleSubmitRevaluation} disabled={isSubmittingRevaluation}>
                  {isSubmittingRevaluation ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Submitting...
                    </>
                  ) : (
                    <>
                      <FileText className="mr-2 h-4 w-4" />
                      Submit Request
                    </>
                  )}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {selectedRequest && canManageExams && (
        <div className="fixed inset-0 z-50 bg-black/50 p-4 flex items-center justify-center">
          <Card className="w-full max-w-lg">
            <CardHeader>
              <div className="flex items-center justify-between gap-4">
                <CardTitle>Revaluation Request Details</CardTitle>
                <Button variant="ghost" size="sm" onClick={() => setSelectedRequest(null)}>
                  Close
                </Button>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <p className="text-xs uppercase text-muted-foreground">Student</p>
                  <p className="font-medium">{selectedRequest.student_name || `Student #${selectedRequest.student_id}`}</p>
                </div>
                <div>
                  <p className="text-xs uppercase text-muted-foreground">Exam</p>
                  <p className="font-medium">{selectedRequest.exam_title || `Exam #${selectedRequest.exam_id}`}</p>
                </div>
                <div>
                  <p className="text-xs uppercase text-muted-foreground">Marks Obtained</p>
                  <p className="font-medium">
                    {typeof selectedRequest.marks_obtained === "number" 
                      ? `${selectedRequest.marks_obtained} / ${selectedRequest.total_marks || "?"}` 
                      : "N/A"}
                  </p>
                </div>
                <div>
                  <p className="text-xs uppercase text-muted-foreground">Request Date</p>
                  <p className="font-medium">{safeFormat(selectedRequest.request_date, "MMM dd, yyyy")}</p>
                </div>
                <div className="md:col-span-2">
                  <p className="text-xs uppercase text-muted-foreground">Status</p>
                  <div className="mt-1">
                    <Badge
                      variant={
                        selectedRequest.status === "approved" ? "default" :
                        selectedRequest.status === "rejected" ? "destructive" :
                        "secondary"
                      }
                      className="capitalize"
                    >
                      {selectedRequest.status}
                    </Badge>
                  </div>
                </div>
                <div className="md:col-span-2">
                  <p className="text-xs uppercase text-muted-foreground">Reason for Revaluation</p>
                  <p className="text-sm mt-1">{selectedRequest.request_reason}</p>
                </div>
              </div>

              {selectedRequest.status === "pending" && (
                <div className="flex justify-end gap-2 pt-4 border-t">
                  <Button
                    variant="outline"
                    className="hover:bg-red-100 hover:text-red-700 dark:hover:bg-red-900/30"
                    onClick={() => {
                      handleRejectRevaluation(selectedRequest.revaluation_id);
                      setSelectedRequest(null);
                    }}
                  >
                    <X className="mr-2 h-4 w-4" />
                    Reject
                  </Button>
                  <Button
                    className="hover:bg-green-100 hover:text-green-700 dark:hover:bg-green-900/30"
                    onClick={() => {
                      handleApproveRevaluation(selectedRequest.revaluation_id);
                      setSelectedRequest(null);
                    }}
                  >
                    <Check className="mr-2 h-4 w-4" />
                    Approve
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      )}

      {selectedExam && (
        <div className="fixed inset-0 z-50 bg-black/50 p-4 flex items-center justify-center">
          <Card className="w-full max-w-2xl">
            <CardHeader>
              <div className="flex items-center justify-between gap-4">
                <div>
                  <CardTitle>{selectedExam.title}</CardTitle>
                  <CardDescription>{selectedExam.course_name || `Course ${selectedExam.course_id ?? selectedExam.id}`}</CardDescription>
                </div>
                <Button variant="ghost" size="sm" onClick={() => setSelectedExam(null)}>
                  Close
                </Button>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <p className="text-xs uppercase text-muted-foreground">Exam Type</p>
                  <p className="font-medium capitalize">{selectedExam.exam_type}</p>
                </div>
                <div>
                  <p className="text-xs uppercase text-muted-foreground">Status</p>
                  <div className="mt-1">{getStatusBadge(normalizeStatus(selectedExam))}</div>
                </div>
                <div>
                  <p className="text-xs uppercase text-muted-foreground">Date</p>
                  <p className="font-medium">{safeFormat(selectedExam.start_time, "MMM dd, yyyy")}</p>
                </div>
                <div>
                  <p className="text-xs uppercase text-muted-foreground">Time</p>
                  <p className="font-medium">{safeFormat(selectedExam.start_time, "hh:mm a")} - {safeFormat(selectedExam.end_time, "hh:mm a")}</p>
                </div>
                <div>
                  <p className="text-xs uppercase text-muted-foreground">Duration</p>
                  <p className="font-medium">{selectedExam.duration} minutes</p>
                </div>
                <div>
                  <p className="text-xs uppercase text-muted-foreground">Room</p>
                  <p className="font-medium">{selectedExam.room_number || "TBD"}</p>
                </div>
                <div>
                  <p className="text-xs uppercase text-muted-foreground">Total Marks</p>
                  <p className="font-medium">{selectedExam.total_marks}</p>
                </div>
                <div>
                  <p className="text-xs uppercase text-muted-foreground">Passing Marks</p>
                  <p className="font-medium">{selectedExam.passing_marks}</p>
                </div>
              </div>

              {selectedExam.description && (
                <div>
                  <p className="text-xs uppercase text-muted-foreground">Description</p>
                  <p className="text-sm mt-1">{selectedExam.description}</p>
                </div>
              )}

              {selectedExam.instructions && (
                <div>
                  <p className="text-xs uppercase text-muted-foreground">Instructions</p>
                  <p className="text-sm mt-1">{selectedExam.instructions}</p>
                </div>
              )}

              {selectedExam.allowed_materials && (
                <div>
                  <p className="text-xs uppercase text-muted-foreground">Allowed Materials</p>
                  <p className="text-sm mt-1">{selectedExam.allowed_materials}</p>
                </div>
              )}

              {user?.role === "student" && studentResults[selectedExam.id] && (
                <div className="rounded-lg border p-4 bg-muted/20">
                  <p className="text-xs uppercase text-muted-foreground mb-2">Your Result</p>
                  <p className="text-sm">Grade: {studentResults[selectedExam.id].grade || "Pending"}</p>
                  <p className="text-sm">Marks: {typeof studentResults[selectedExam.id].marksObtained === "number" ? studentResults[selectedExam.id].marksObtained : "--"}</p>
                  <p className="text-sm">Percentage: {typeof studentResults[selectedExam.id].percentage === "number" ? `${studentResults[selectedExam.id].percentage?.toFixed(1)}%` : "--"}</p>
                  <p className="text-sm">Result: {studentResults[selectedExam.id].result || "pending"}</p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      )}

      {canManageExams && isCreateRoomDialogOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 p-4 flex items-center justify-center">
          <Card className="w-full max-w-lg">
            <CardHeader>
              <div className="flex items-center justify-between gap-4">
                <CardTitle>Add Exam Room</CardTitle>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    setIsCreateRoomDialogOpen(false);
                    setRoomFormError(null);
                  }}
                >
                  Close
                </Button>
              </div>
              <CardDescription>Create a new exam room for scheduling.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium" htmlFor="roomNumber">Room Number</label>
                <Input
                  id="roomNumber"
                  value={roomForm.roomNumber}
                  onChange={(event) => setRoomForm((prev) => ({ ...prev, roomNumber: event.target.value }))}
                  placeholder="e.g., A-101, LH-1"
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium" htmlFor="roomCapacity">Capacity</label>
                <Input
                  id="roomCapacity"
                  type="number"
                  min={1}
                  value={roomForm.capacity}
                  onChange={(event) => setRoomForm((prev) => ({ ...prev, capacity: event.target.value }))}
                  placeholder="30"
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium" htmlFor="roomBuilding">Building</label>
                <Input
                  id="roomBuilding"
                  value={roomForm.building}
                  onChange={(event) => setRoomForm((prev) => ({ ...prev, building: event.target.value }))}
                  placeholder="e.g., Main Building, Science Block"
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium" htmlFor="roomFloor">Floor</label>
                <Input
                  id="roomFloor"
                  value={roomForm.floor}
                  onChange={(event) => setRoomForm((prev) => ({ ...prev, floor: event.target.value }))}
                  placeholder="1"
                />
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="roomAvailable"
                  checked={roomForm.isAvailable}
                  onChange={(event) => setRoomForm((prev) => ({ ...prev, isAvailable: event.target.checked }))}
                  className="h-4 w-4 rounded border-gray-300"
                />
                <label htmlFor="roomAvailable" className="text-sm font-medium">Room is available</label>
              </div>

              {roomFormError && <p className="text-sm text-destructive">{roomFormError}</p>}

              <div className="flex justify-end gap-2">
                <Button
                  variant="outline"
                  onClick={() => {
                    setIsCreateRoomDialogOpen(false);
                    setRoomFormError(null);
                  }}
                  disabled={isSavingRoom}
                >
                  Cancel
                </Button>
                <Button onClick={handleCreateRoom} disabled={isSavingRoom}>
                  {isSavingRoom ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Creating...
                    </>
                  ) : (
                    <>
                      <Plus className="mr-2 h-4 w-4" />
                      Create Room
                    </>
                  )}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {canManageExams && isEditRoomDialogOpen && selectedRoom && (
        <div className="fixed inset-0 z-50 bg-black/50 p-4 flex items-center justify-center">
          <Card className="w-full max-w-lg">
            <CardHeader>
              <div className="flex items-center justify-between gap-4">
                <CardTitle>Edit Exam Room</CardTitle>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    setIsEditRoomDialogOpen(false);
                    setSelectedRoom(null);
                    setRoomFormError(null);
                  }}
                >
                  Close
                </Button>
              </div>
              <CardDescription>Update exam room details.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium" htmlFor="editRoomNumber">Room Number</label>
                <Input
                  id="editRoomNumber"
                  value={roomForm.roomNumber}
                  onChange={(event) => setRoomForm((prev) => ({ ...prev, roomNumber: event.target.value }))}
                  placeholder="e.g., A-101, LH-1"
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium" htmlFor="editRoomCapacity">Capacity</label>
                <Input
                  id="editRoomCapacity"
                  type="number"
                  min={1}
                  value={roomForm.capacity}
                  onChange={(event) => setRoomForm((prev) => ({ ...prev, capacity: event.target.value }))}
                  placeholder="30"
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium" htmlFor="editRoomBuilding">Building</label>
                <Input
                  id="editRoomBuilding"
                  value={roomForm.building}
                  onChange={(event) => setRoomForm((prev) => ({ ...prev, building: event.target.value }))}
                  placeholder="e.g., Main Building, Science Block"
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium" htmlFor="editRoomFloor">Floor</label>
                <Input
                  id="editRoomFloor"
                  value={roomForm.floor}
                  onChange={(event) => setRoomForm((prev) => ({ ...prev, floor: event.target.value }))}
                  placeholder="1"
                />
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="editRoomAvailable"
                  checked={roomForm.isAvailable}
                  onChange={(event) => setRoomForm((prev) => ({ ...prev, isAvailable: event.target.checked }))}
                  className="h-4 w-4 rounded border-gray-300"
                />
                <label htmlFor="editRoomAvailable" className="text-sm font-medium">Room is available</label>
              </div>

              {roomFormError && <p className="text-sm text-destructive">{roomFormError}</p>}

              <div className="flex justify-end gap-2">
                <Button
                  variant="outline"
                  onClick={() => {
                    setIsEditRoomDialogOpen(false);
                    setSelectedRoom(null);
                    setRoomFormError(null);
                  }}
                  disabled={isSavingRoom}
                >
                  Cancel
                </Button>
                <Button onClick={handleUpdateRoom} disabled={isSavingRoom}>
                  {isSavingRoom ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Saving...
                    </>
                  ) : (
                    <>
                      <Check className="mr-2 h-4 w-4" />
                      Save Changes
                    </>
                  )}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {canManageExams && isAvailabilityDialogOpen && selectedRoom && (
        <div className="fixed inset-0 z-50 bg-black/50 p-4 flex items-center justify-center">
          <Card className="w-full max-w-lg">
            <CardHeader>
              <div className="flex items-center justify-between gap-4">
                <CardTitle>Room Availability</CardTitle>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    setIsAvailabilityDialogOpen(false);
                    setSelectedRoom(null);
                    setRoomAvailability(null);
                  }}
                >
                  Close
                </Button>
              </div>
              <CardDescription>Check availability for {selectedRoom.room_number}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {checkingAvailability ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="h-6 w-6 animate-spin" />
                  <span className="ml-2">Checking availability...</span>
                </div>
              ) : roomAvailability ? (
                <div className="space-y-4">
                  <div className="flex items-center justify-center p-4 rounded-lg bg-muted/30">
                    {roomAvailability.available ? (
                      <div className="text-center">
                        <CheckCircle className="h-12 w-12 mx-auto text-green-500 mb-2" />
                        <p className="text-lg font-semibold text-green-600">Available</p>
                        <p className="text-sm text-muted-foreground">{roomAvailability.message || "Room is available for scheduling"}</p>
                      </div>
                    ) : (
                      <div className="text-center">
                        <X className="h-12 w-12 mx-auto text-red-500 mb-2" />
                        <p className="text-lg font-semibold text-red-600">Not Available</p>
                        <p className="text-sm text-muted-foreground">{roomAvailability.message || "Room is not available"}</p>
                      </div>
                    )}
                  </div>
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <p className="text-xs uppercase text-muted-foreground">Room</p>
                      <p className="font-medium">{selectedRoom.room_number}</p>
                    </div>
                    <div>
                      <p className="text-xs uppercase text-muted-foreground">Capacity</p>
                      <p className="font-medium">{selectedRoom.capacity}</p>
                    </div>
                    <div>
                      <p className="text-xs uppercase text-muted-foreground">Building</p>
                      <p className="font-medium">{selectedRoom.building}</p>
                    </div>
                    <div>
                      <p className="text-xs uppercase text-muted-foreground">Floor</p>
                      <p className="font-medium">Floor {selectedRoom.floor}</p>
                    </div>
                  </div>
                </div>
              ) : (
                <div className="text-center py-8 text-muted-foreground">
                  <p>Unable to fetch availability information.</p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      )}
    </>
  );
}
