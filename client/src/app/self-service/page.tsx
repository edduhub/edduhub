"use client";

import { useState } from 'react';
import { useSelfServiceRequests, useCreateSelfServiceRequest, useUpdateSelfServiceRequest, useAllSelfServiceRequests } from '@/lib/api-hooks';
import { useAuth } from '@/lib/auth-context';
import { logger } from '@/lib/logger';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { 
  BookOpen, 
  Calendar, 
  FileText, 
  Send,
  CheckCircle,
  Clock,
  AlertCircle,
  Loader2,
  Users,
  XCircle,
  PlayCircle
} from 'lucide-react';

type RequestStatus = 'pending' | 'approved' | 'rejected' | 'processing';

type Request = {
  id: number;
  type: 'enrollment' | 'schedule' | 'transcript' | 'document';
  title: string;
  description: string;
  status: RequestStatus;
  submittedAt: string;
  respondedAt?: string;
  response?: string;
};

export default function StudentSelfServicePage() {
  const { user } = useAuth();
  const { data: requests = [] } = useSelfServiceRequests();
  const isAdminOrFaculty = user?.role === 'admin' || user?.role === 'faculty' || user?.role === 'super_admin';

  return (
    <div className="min-h-screen bg-muted/10">
      <div className="container mx-auto p-6 space-y-6">
        {/* Header */}
        <div>
          <h1 className="text-3xl font-bold">Student Self-Service</h1>
          <p className="text-muted-foreground mt-1">
            Submit requests for enrollment, schedule changes, and documents
          </p>
        </div>

        {/* Request Tabs */}
        <Tabs defaultValue="enroll" className="space-y-4">
          <TabsList className={`grid w-full grid-cols-${isAdminOrFaculty ? 5 : 4} lg:w-auto`}>
            <TabsTrigger value="enroll">
              <BookOpen className="w-4 h-4 mr-2" />
              Enrollment
            </TabsTrigger>
            <TabsTrigger value="schedule">
              <Calendar className="w-4 h-4 mr-2" />
              Schedule
            </TabsTrigger>
            <TabsTrigger value="documents">
              <FileText className="w-4 h-4 mr-2" />
              Documents
            </TabsTrigger>
            {isAdminOrFaculty && (
              <TabsTrigger value="admin">
                <Users className="w-4 h-4 mr-2" />
                Manage
              </TabsTrigger>
            )}
            <TabsTrigger value="history">
              <Clock className="w-4 h-4 mr-2" />
              History
            </TabsTrigger>
          </TabsList>

          <TabsContent value="enroll" className="space-y-4">
            <EnrollmentRequestForm />
          </TabsContent>

          <TabsContent value="schedule" className="space-y-4">
            <ScheduleChangeRequestForm />
          </TabsContent>

          <TabsContent value="documents" className="space-y-4">
            <DocumentRequestForm />
          </TabsContent>

          {isAdminOrFaculty && (
            <TabsContent value="admin" className="space-y-4">
              <AdminRequestManager />
            </TabsContent>
          )}

          <TabsContent value="history" className="space-y-4">
            <RequestHistory requests={requests} />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}

function EnrollmentRequestForm() {
  const [formData, setFormData] = useState({
    courseCode: '',
    reason: '',
    specialRequests: '',
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitSuccess, setSubmitSuccess] = useState<string | null>(null);
  const createRequest = useCreateSelfServiceRequest();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitError(null);
    setSubmitSuccess(null);
    setIsSubmitting(true);

    try {
      await createRequest.mutateAsync({
        type: 'enrollment',
        title: `Enrollment Request: ${formData.courseCode}`,
        description: `Reason: ${formData.reason}\n\nSpecial Requests: ${formData.specialRequests || 'None'}`,
      });
      
      setFormData({ courseCode: '', reason: '', specialRequests: '' });
      setSubmitSuccess('Enrollment request submitted successfully.');
    } catch (error) {
      logger.error('Failed to submit request:', error as Error);
      setSubmitError('Failed to submit request. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <BookOpen className="w-5 h-5" />
          Course Enrollment Request
        </CardTitle>
        <CardDescription>
          Request enrollment in additional courses or special permissions
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="courseCode">Course Code</Label>
              <Input
                id="courseCode"
                value={formData.courseCode}
                onChange={(e) => setFormData({ ...formData, courseCode: e.target.value })}
                placeholder="e.g., CS401"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="semester">Semester</Label>
              <select
                id="semester"
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                <option>Fall 2024</option>
                <option>Spring 2025</option>
                <option>Summer 2025</option>
              </select>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="reason">Reason for Request</Label>
            <Textarea
              id="reason"
              value={formData.reason}
              onChange={(e) => setFormData({ ...formData, reason: e.target.value })}
              placeholder="Explain why you need to enroll in this course..."
              className="min-h-[120px]"
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="specialRequests">Special Requests (Optional)</Label>
            <Textarea
              id="specialRequests"
              value={formData.specialRequests}
              onChange={(e) => setFormData({ ...formData, specialRequests: e.target.value })}
              placeholder="Any special requirements or notes..."
              className="min-h-[80px]"
            />
          </div>

          <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg">
            <h4 className="font-semibold mb-2 flex items-center gap-2">
              <AlertCircle className="w-4 h-4 text-blue-600" />
              Enrollment Guidelines
            </h4>
            <ul className="space-y-1 text-sm text-muted-foreground">
              <li>• Ensure prerequisites are met before requesting enrollment</li>
              <li>• Enrollment requests are reviewed by academic advisors</li>
              <li>• Processing time: 3-5 business days</li>
              <li>• You will be notified via email when decision is made</li>
            </ul>
          </div>

          {submitError && <p className="text-sm text-destructive">{submitError}</p>}
          {submitSuccess && <p className="text-sm text-green-700">{submitSuccess}</p>}

          <Button type="submit" className="w-full" disabled={isSubmitting}>
            {isSubmitting ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Submitting...
              </>
            ) : (
              <>
                <Send className="w-4 h-4 mr-2" />
                Submit Request
              </>
            )}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

function ScheduleChangeRequestForm() {
  const [formData, setFormData] = useState({
    courseId: '',
    currentSection: '',
    requestedSection: '',
    reason: '',
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitSuccess, setSubmitSuccess] = useState<string | null>(null);
  const createRequest = useCreateSelfServiceRequest();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitError(null);
    setSubmitSuccess(null);
    setIsSubmitting(true);

    try {
      await createRequest.mutateAsync({
        type: 'schedule',
        title: `Schedule Change: ${formData.courseId}`,
        description: `Change from Section ${formData.currentSection} to Section ${formData.requestedSection}\n\nReason: ${formData.reason}`,
      });
      
      setFormData({ courseId: '', currentSection: '', requestedSection: '', reason: '' });
      setSubmitSuccess('Schedule change request submitted successfully.');
    } catch (error) {
      logger.error('Failed to submit request:', error as Error);
      setSubmitError('Failed to submit request. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Calendar className="w-5 h-5" />
          Schedule Change Request
        </CardTitle>
        <CardDescription>
          Request changes to your course schedule or section assignments
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="courseId">Course</Label>
              <select
                id="courseId"
                value={formData.courseId}
                onChange={(e) => setFormData({ ...formData, courseId: e.target.value })}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                required
              >
                <option value="">Select a course</option>
                <option value="CS101">CS101 - Introduction to Programming</option>
                <option value="CS201">CS201 - Data Structures</option>
                <option value="CS301">CS301 - Algorithms</option>
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="currentSection">Current Section</Label>
              <select
                id="currentSection"
                value={formData.currentSection}
                onChange={(e) => setFormData({ ...formData, currentSection: e.target.value })}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                required
              >
                <option value="">Select section</option>
                <option value="A">Section A (Mon/Wed 9-11 AM)</option>
                <option value="B">Section B (Tue/Thu 2-4 PM)</option>
              </select>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="requestedSection">Requested Section</Label>
            <select
              id="requestedSection"
              value={formData.requestedSection}
              onChange={(e) => setFormData({ ...formData, requestedSection: e.target.value })}
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              required
            >
              <option value="">Select section</option>
              <option value="A">Section A (Mon/Wed 9-11 AM)</option>
              <option value="B">Section B (Tue/Thu 2-4 PM)</option>
              <option value="C">Section C (Wed/Fri 10 AM-12 PM)</option>
            </select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="reason">Reason for Change</Label>
            <Textarea
              id="reason"
              value={formData.reason}
              onChange={(e) => setFormData({ ...formData, reason: e.target.value })}
              placeholder="Explain why you need to change sections..."
              className="min-h-[120px]"
              required
            />
          </div>

          <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
            <h4 className="font-semibold mb-2 flex items-center gap-2">
              <AlertCircle className="w-4 h-4 text-yellow-600" />
              Schedule Change Policy
            </h4>
            <ul className="space-y-1 text-sm text-muted-foreground">
              <li>• Schedule changes are subject to availability</li>
              <li>• Must be submitted before add/drop deadline</li>
              <li>• Conflicts with other courses may not be allowed</li>
              <li>• Processing time: 2-3 business days</li>
            </ul>
          </div>

          {submitError && <p className="text-sm text-destructive">{submitError}</p>}
          {submitSuccess && <p className="text-sm text-green-700">{submitSuccess}</p>}

          <Button type="submit" className="w-full" disabled={isSubmitting}>
            {isSubmitting ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Submitting...
              </>
            ) : (
              <>
                <Send className="w-4 h-4 mr-2" />
                Submit Request
              </>
            )}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

function DocumentRequestForm() {
  const [formData, setFormData] = useState({
    documentType: 'transcript' as 'transcript' | 'certificate' | 'id_card' | 'other',
    purpose: '',
    copies: 1,
    deliveryMethod: 'pickup' as 'pickup' | 'email' | 'postal',
    address: '',
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitSuccess, setSubmitSuccess] = useState<string | null>(null);
  const createRequest = useCreateSelfServiceRequest();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitError(null);
    setSubmitSuccess(null);
    setIsSubmitting(true);

    try {
      const requestType = formData.documentType === 'transcript' ? 'transcript' : 'document';
      await createRequest.mutateAsync({
        type: requestType,
        title: `Document Request: ${formData.documentType}`,
        description: `Purpose: ${formData.purpose}\nCopies: ${formData.copies}\nDelivery: ${formData.deliveryMethod}\nAddress: ${formData.address || 'N/A'}`,
        document_type: formData.documentType,
        delivery_method: formData.deliveryMethod,
      });
      
      setFormData({
        documentType: 'transcript',
        purpose: '',
        copies: 1,
        deliveryMethod: 'pickup',
        address: '',
      });
      setSubmitSuccess('Document request submitted successfully.');
    } catch (error) {
      logger.error('Failed to submit request:', error as Error);
      setSubmitError('Failed to submit request. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <FileText className="w-5 h-5" />
          Document Request
        </CardTitle>
        <CardDescription>
          Request official transcripts, grade cards, and certificates
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="documentType">Document Type</Label>
            <select
              id="documentType"
              value={formData.documentType}
              onChange={(e) => setFormData({ ...formData, documentType: e.target.value as 'transcript' | 'certificate' | 'id_card' | 'other' })}
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              required
            >
              <option value="transcript">Official Transcript</option>
              <option value="certificate">Enrollment Certificate</option>
              <option value="id_card">ID Card Replacement</option>
              <option value="other">Other Document</option>
            </select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="purpose">Purpose</Label>
            <Input
              id="purpose"
              value={formData.purpose}
              onChange={(e) => setFormData({ ...formData, purpose: e.target.value })}
              placeholder="e.g., Higher Education, Employment, Visa Application"
              required
            />
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="copies">Number of Copies</Label>
              <Input
                id="copies"
                type="number"
                min="1"
                max="10"
                value={formData.copies}
                onChange={(e) => {
                  const parsed = Number.parseInt(e.target.value, 10);
                  const nextCopies = Number.isFinite(parsed)
                    ? Math.min(10, Math.max(1, parsed))
                    : 1;
                  setFormData({ ...formData, copies: nextCopies });
                }}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="deliveryMethod">Delivery Method</Label>
              <select
                id="deliveryMethod"
                value={formData.deliveryMethod}
                onChange={(e) => setFormData({ ...formData, deliveryMethod: e.target.value as 'pickup' | 'email' | 'postal' })}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                required
              >
                <option value="pickup">Pick up at Registrar Office</option>
                <option value="email">Email (Digital Copy)</option>
                <option value="postal">Mail to Address</option>
              </select>
            </div>
          </div>

          {formData.deliveryMethod === 'postal' && (
            <div className="space-y-2">
              <Label htmlFor="address">Mailing Address</Label>
              <Textarea
                id="address"
                value={formData.address}
                onChange={(e) => setFormData({ ...formData, address: e.target.value })}
                placeholder="Full mailing address..."
                className="min-h-[100px]"
                required
              />
            </div>
          )}

          <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
            <h4 className="font-semibold mb-2 flex items-center gap-2">
              <CheckCircle className="w-4 h-4 text-green-600" />
              Document Processing
            </h4>
            <ul className="space-y-1 text-sm text-muted-foreground">
              <li>• Processing time: 5-7 business days</li>
              <li>• Digital documents: 2-3 business days</li>
              <li>• First transcript: Free (within academic year)</li>
              <li>• Additional copies: $5 per copy</li>
              <li>• Express processing available for additional fee</li>
            </ul>
          </div>

          {submitError && <p className="text-sm text-destructive">{submitError}</p>}
          {submitSuccess && <p className="text-sm text-green-700">{submitSuccess}</p>}

          <Button type="submit" className="w-full" disabled={isSubmitting}>
            {isSubmitting ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Submitting...
              </>
            ) : (
              <>
                <Send className="w-4 h-4 mr-2" />
                Submit Request
              </>
            )}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

function RequestHistory({ requests }: { requests: Request[] }) {
  const statusColors: Record<RequestStatus, string> = {
    pending: 'bg-yellow-100 text-yellow-800 border-yellow-200',
    approved: 'bg-green-100 text-green-800 border-green-200',
    rejected: 'bg-red-100 text-red-800 border-red-200',
    processing: 'bg-blue-100 text-blue-800 border-blue-200',
  };

  const statusIcons: Record<RequestStatus, React.ReactElement> = {
    pending: <Clock className="w-4 h-4" />,
    approved: <CheckCircle className="w-4 h-4" />,
    rejected: <AlertCircle className="w-4 h-4" />,
    processing: <Clock className="w-4 h-4 animate-pulse" />,
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Clock className="w-5 h-5" />
          Request History
        </CardTitle>
        <CardDescription>
          Track the status of your submitted requests
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {requests.length === 0 && (
            <div className="rounded-lg border p-6 text-sm text-muted-foreground">
              No requests submitted yet.
            </div>
          )}
          {requests.map((request) => (
            <div 
              key={request.id} 
              className={`p-4 border rounded-lg ${statusColors[request.status]}`}
            >
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-2">
                    {statusIcons[request.status]}
                    <h4 className="font-semibold">{request.title}</h4>
                    <Badge className="capitalize">{request.type}</Badge>
                  </div>
                  <p className="text-sm mb-2">{request.description}</p>
                  <div className="flex items-center gap-4 text-xs text-muted-foreground">
                    <span>Submitted: {new Date(request.submittedAt).toLocaleString()}</span>
                    {request.respondedAt && (
                      <span>Responded: {new Date(request.respondedAt).toLocaleString()}</span>
                    )}
                  </div>
                  {request.response && (
                    <div className="mt-3 p-3 bg-white/50 rounded">
                      <h5 className="font-medium text-sm mb-1">Response:</h5>
                      <p className="text-sm">{request.response}</p>
                    </div>
                  )}
                </div>
                <Badge className="capitalize">{request.status}</Badge>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

function AdminRequestManager() {
  const { data: allRequests = [], isLoading } = useAllSelfServiceRequests();
  const updateRequest = useUpdateSelfServiceRequest();
  const [selectedRequest, setSelectedRequest] = useState<Request | null>(null);
  const [responseText, setResponseText] = useState('');
  const [filterStatus, setFilterStatus] = useState<string>('all');

  const statusColors: Record<RequestStatus, string> = {
    pending: 'bg-yellow-100 text-yellow-800 border-yellow-200',
    approved: 'bg-green-100 text-green-800 border-green-200',
    rejected: 'bg-red-100 text-red-800 border-red-200',
    processing: 'bg-blue-100 text-blue-800 border-blue-200',
  };

  const filteredRequests = filterStatus === 'all' 
    ? allRequests 
    : allRequests.filter(r => r.status === filterStatus);

  const handleStatusChange = async (requestId: number, newStatus: 'approved' | 'rejected' | 'processing') => {
    try {
      await updateRequest.mutateAsync({
        requestId,
        input: { 
          status: newStatus,
          response: responseText || undefined
        }
      });
      setSelectedRequest(null);
      setResponseText('');
    } catch (error) {
      logger.error('Failed to update request:', error as Error);
    }
  };

  if (isLoading) {
    return (
      <Card>
        <CardContent className="py-10 flex justify-center">
          <Loader2 className="w-6 h-6 animate-spin" />
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Users className="w-5 h-5" />
          Manage Requests
        </CardTitle>
        <CardDescription>
          Review and process student self-service requests
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Filter */}
        <div className="flex gap-2 flex-wrap">
          {['all', 'pending', 'processing', 'approved', 'rejected'].map((status) => (
            <Button
              key={status}
              variant={filterStatus === status ? 'default' : 'outline'}
              size="sm"
              onClick={() => setFilterStatus(status)}
              className="capitalize"
            >
              {status}
            </Button>
          ))}
        </div>

        {/* Requests List */}
        <div className="space-y-4">
          {filteredRequests.length === 0 && (
            <div className="rounded-lg border p-6 text-sm text-muted-foreground">
              No requests found.
            </div>
          )}
          {filteredRequests.map((request) => (
            <div 
              key={request.id} 
              className={`p-4 border rounded-lg ${statusColors[request.status]}`}
            >
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-2">
                    <Badge className="capitalize">{request.type}</Badge>
                    <span className="text-xs text-muted-foreground">
                      ID: {request.id}
                    </span>
                  </div>
                  <h4 className="font-semibold mb-1">{request.title}</h4>
                  <p className="text-sm mb-2">{request.description}</p>
                  <div className="flex items-center gap-4 text-xs text-muted-foreground">
                    <span>Submitted: {new Date(request.submittedAt).toLocaleString()}</span>
                    {request.respondedAt && (
                      <span>Updated: {new Date(request.respondedAt).toLocaleString()}</span>
                    )}
                  </div>
                  {request.response && (
                    <div className="mt-3 p-3 bg-white/50 rounded">
                      <h5 className="font-medium text-sm mb-1">Response:</h5>
                      <p className="text-sm">{request.response}</p>
                    </div>
                  )}
                </div>
                <div className="flex flex-col gap-2 items-end">
                  <Badge className="capitalize">{request.status}</Badge>
                  {request.status !== 'approved' && request.status !== 'rejected' && (
                    <div className="flex gap-1 mt-2">
                      <Button
                        size="sm"
                        variant="outline"
                        className="h-8 bg-blue-50 border-blue-200 hover:bg-blue-100 text-blue-700"
                        onClick={() => setSelectedRequest(request)}
                      >
                        <PlayCircle className="w-3 h-3 mr-1" />
                        Process
                      </Button>
                    </div>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* Process Dialog */}
        {selectedRequest && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
            <Card className="w-full max-w-md mx-4">
              <CardHeader>
                <CardTitle>Process Request</CardTitle>
                <CardDescription>
                  Review and update request #{selectedRequest.id}
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="p-3 bg-muted rounded-lg">
                  <p className="font-medium">{selectedRequest.title}</p>
                  <p className="text-sm text-muted-foreground mt-1">{selectedRequest.description}</p>
                </div>
                
                <div className="space-y-2">
                  <Label htmlFor="response">Response Message (Optional)</Label>
                  <Textarea
                    id="response"
                    value={responseText}
                    onChange={(e) => setResponseText(e.target.value)}
                    placeholder="Add a response note for the student..."
                    className="min-h-[100px]"
                  />
                </div>

                <div className="flex gap-2">
                  <Button
                    className="flex-1 bg-green-600 hover:bg-green-700"
                    onClick={() => handleStatusChange(selectedRequest.id, 'approved')}
                    disabled={updateRequest.isPending}
                  >
                    <CheckCircle className="w-4 h-4 mr-2" />
                    Approve
                  </Button>
                  <Button
                    className="flex-1 bg-blue-600 hover:bg-blue-700"
                    onClick={() => handleStatusChange(selectedRequest.id, 'processing')}
                    disabled={updateRequest.isPending}
                  >
                    <PlayCircle className="w-4 h-4 mr-2" />
                    Process
                  </Button>
                  <Button
                    className="flex-1 bg-red-600 hover:bg-red-700"
                    variant="destructive"
                    onClick={() => handleStatusChange(selectedRequest.id, 'rejected')}
                    disabled={updateRequest.isPending}
                  >
                    <XCircle className="w-4 h-4 mr-2" />
                    Reject
                  </Button>
                </div>

                <Button
                  variant="outline"
                  className="w-full"
                  onClick={() => {
                    setSelectedRequest(null);
                    setResponseText('');
                  }}
                >
                  Cancel
                </Button>
              </CardContent>
            </Card>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
