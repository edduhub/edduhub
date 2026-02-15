"use client";

import { useState, useEffect } from "react";
import Image from "next/image";
import { useAuth } from "@/lib/auth-context";
import { api } from "@/lib/api-client";
import { Card, CardContent, CardHeader, CardTitle, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
    CreditCard,
    AlertTriangle,
    CheckCircle2,
    Download,
    DollarSign,
    PieChart,
    ArrowRight,
    ShieldCheck,
    Zap
} from "lucide-react";
import { format } from "date-fns";
import { logger } from "@/lib/logger";

interface FeeAssignment {
    id: number;
    fee_structure_id: number;
    title: string;
    amount: number;
    due_date: string;
    status: string;
    waiver_amount: number;
    paid_amount: number;
    category: string;
}

interface Payment {
    id: number;
    amount: number;
    payment_date: string;
    payment_method: string;
    payment_status: string;
    transaction_id: string;
    fee_assignment_title: string;
}

declare global {
    interface Window {
        Razorpay: new (options: object) => { open: () => void; on: (event: string, handler: () => void) => void };
    }
}

export default function FeesPage() {
    const { user } = useAuth();
    const [assignments, setAssignments] = useState<FeeAssignment[]>([]);
    const [payments, setPayments] = useState<Payment[]>([]);
    const [loading, setLoading] = useState(true);
    const [payingId, setPayingId] = useState<number | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState<string | null>(null);
    const [downloadingStatement, setDownloadingStatement] = useState(false);
    const [downloadingInvoiceId, setDownloadingInvoiceId] = useState<number | null>(null);

    // SECURITY: Validate Razorpay key is configured
    const razorpayKey = process.env.NEXT_PUBLIC_RAZORPAY_KEY_ID;
    const isPaymentConfigured = !!razorpayKey && !razorpayKey.includes("your_key");

    useEffect(() => {
        const fetchFeeData = async () => {
            try {
                setLoading(true);
                const [assignmentsData, paymentsData] = await Promise.all([
                    api.get<FeeAssignment[]>("/api/fees/my-fees"),
                    api.get<Payment[]>("/api/fees/my-payments")
                ]);
                setAssignments(Array.isArray(assignmentsData) ? assignmentsData : []);
                setPayments(Array.isArray(paymentsData) ? paymentsData : []);
            } catch (err) {
                logger.error("Failed to fetch fee data:", err as Error);
                setError("Failed to load fee information. Please try again.");
            } finally {
                setLoading(false);
            }
        };
        fetchFeeData();
    }, []);

    const initiatePayment = async (assignment: FeeAssignment) => {
        setError(null);
        setSuccess(null);

        // Validate payment configuration
        if (!isPaymentConfigured) {
            setError("Payment system is not configured. Please contact support.");
            return;
        }

        try {
            setPayingId(assignment.id);

            // 1. Create Order on Backend
            const order = await api.post<{ payment_id: number, transaction_id: string }>("/api/fees/payments/online", {
                fee_assignment_id: assignment.id,
                amount: assignment.amount - assignment.paid_amount - assignment.waiver_amount,
                gateway: "razorpay"
            });

            // 2. Load Razorpay Script if not already loaded
            // SECURITY: Using known Razorpay CDN URL with integrity check would be ideal
            // For now, we're trusting the official Razorpay CDN
            if (!window.Razorpay) {
                const script = document.createElement("script");
                script.src = "https://checkout.razorpay.com/v1/checkout.js";
                script.async = true;
                document.body.appendChild(script);

                await new Promise((resolve, reject) => {
                    script.onload = resolve;
                    script.onerror = () => reject(new Error("Failed to load payment script"));
                });
            }

            // 3. Open Razorpay Checkout
            const options = {
                key: razorpayKey,
                amount: (assignment.amount - assignment.paid_amount - assignment.waiver_amount) * 100,
                currency: "INR",
                name: "EduHub College",
                description: `Payment for ${assignment.title}`,
                order_id: order.transaction_id,
                handler: async (response: { razorpay_order_id: string; razorpay_payment_id: string; razorpay_signature: string }) => {
                    try {
                        // 4. Verify Payment on Backend
                        await api.post("/api/fees/payments/verify", {
                            payment_id: order.payment_id,
                            order_id: response.razorpay_order_id,
                            transaction_id: response.razorpay_payment_id,
                            signature: response.razorpay_signature
                        });

                        setSuccess("Payment successful! Refreshing...");
                        // Refresh data after short delay
                        setTimeout(() => window.location.reload(), 1500);
                    } catch (err) {
                        logger.error("Payment verification failed:", err as Error);
                        setError("Payment verification failed. Please contact support if money was deducted.");
                    }
                },
                prefill: {
                    name: `${user?.firstName} ${user?.lastName}`,
                    email: user?.email,
                },
                theme: {
                    color: "#3B82F6",
                },
            };

            const rzp = new window.Razorpay(options);
            rzp.open();
        } catch (err) {
            logger.error("Failed to initiate payment:", err as Error);
            setError("Failed to initiate payment. Please try again.");
        } finally {
            setPayingId(null);
        }
    };

    const pendingTotal = assignments
        .filter((a: FeeAssignment) => a.status !== 'paid')
        .reduce((acc: number, a: FeeAssignment) => acc + (a.amount - a.paid_amount - a.waiver_amount), 0);

    const triggerDownload = (filename: string, content: string, mimeType = "text/plain;charset=utf-8") => {
        const blob = new Blob([content], { type: mimeType });
        const url = window.URL.createObjectURL(blob);
        const link = document.createElement("a");
        link.href = url;
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        link.remove();
        window.URL.revokeObjectURL(url);
    };

    const handleDownloadStatement = () => {
        try {
            setDownloadingStatement(true);
            setError(null);
            setSuccess(null);

            const assignmentLines = assignments.map((assignment) =>
                [
                    assignment.id,
                    `"${assignment.title.replace(/\"/g, '""')}"`,
                    assignment.category,
                    assignment.amount,
                    assignment.paid_amount,
                    assignment.waiver_amount,
                    assignment.status,
                    assignment.due_date,
                ].join(",")
            );

            const paymentLines = payments.map((payment) =>
                [
                    payment.id,
                    `"${(payment.transaction_id || "").replace(/\"/g, '""')}"`,
                    `"${(payment.fee_assignment_title || "Course Fee").replace(/\"/g, '""')}"`,
                    payment.amount,
                    payment.payment_method,
                    payment.payment_status,
                    payment.payment_date,
                ].join(",")
            );

            const statementContent = [
                "Fee Assignments",
                "id,title,category,amount,paid_amount,waiver_amount,status,due_date",
                ...assignmentLines,
                "",
                "Payments",
                "id,transaction_id,fee_title,amount,payment_method,payment_status,payment_date",
                ...paymentLines,
            ].join("\n");

            triggerDownload(
                `fee-statement-${new Date().toISOString().split("T")[0]}.csv`,
                statementContent,
                "text/csv;charset=utf-8"
            );
            setSuccess("Statement downloaded successfully");
        } catch (downloadError) {
            logger.error("Failed to download statement:", downloadError as Error);
            setError("Failed to download statement");
        } finally {
            setDownloadingStatement(false);
        }
    };

    const handleQuickPay = async () => {
        const firstPending = assignments.find(
            (assignment) =>
                assignment.status !== "paid" &&
                assignment.amount - assignment.paid_amount - assignment.waiver_amount > 0
        );

        if (!firstPending) {
            setError("No pending fees available for quick pay");
            return;
        }

        await initiatePayment(firstPending);
    };

    const handleInvoiceDownload = (payment: Payment) => {
        try {
            setDownloadingInvoiceId(payment.id);
            setError(null);

            const invoiceText = [
                "EduHub Fee Payment Invoice",
                "-------------------------",
                `Invoice ID: INV-${payment.id}`,
                `Transaction ID: ${payment.transaction_id || `TXN-${payment.id}`}`,
                `Student: ${user?.firstName || ""} ${user?.lastName || ""}`.trim(),
                `Email: ${user?.email || "N/A"}`,
                `Fee Title: ${payment.fee_assignment_title || "Course Fee"}`,
                `Amount: INR ${payment.amount.toLocaleString()}`,
                `Payment Method: ${payment.payment_method}`,
                `Status: ${payment.payment_status}`,
                `Payment Date: ${new Date(payment.payment_date).toLocaleString()}`,
                "",
                `Generated At: ${new Date().toLocaleString()}`,
            ].join("\n");

            triggerDownload(`invoice-${payment.id}.txt`, invoiceText);
            setSuccess("Invoice downloaded");
        } catch (downloadError) {
            logger.error("Failed to download invoice:", downloadError as Error);
            setError("Failed to download invoice");
        } finally {
            setDownloadingInvoiceId(null);
        }
    };

    return (
        <div className="space-y-8 pb-10">
            {/* Error/Success Messages */}
            {error && (
                <div className="bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded-lg flex items-center gap-2">
                    <AlertTriangle className="h-5 w-5" />
                    <span>{error}</span>
                    <button onClick={() => setError(null)} className="ml-auto text-red-600 hover:text-red-800">×</button>
                </div>
            )}
            {success && (
                <div className="bg-green-50 border border-green-200 text-green-800 px-4 py-3 rounded-lg flex items-center gap-2">
                    <CheckCircle2 className="h-5 w-5" />
                    <span>{success}</span>
                </div>
            )}
            <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
                <div>
                    <h1 className="text-4xl font-extrabold tracking-tight lg:text-5xl bg-gradient-to-r from-blue-600 via-indigo-600 to-purple-600 bg-clip-text text-transparent">
                        Fees & Payments
                    </h1>
                    <p className="text-muted-foreground mt-2 text-lg">
                        Manage your educational expenses and view transaction history securely.
                    </p>
                </div>
                <div className="flex gap-2">
                    <Button variant="outline" className="shadow-sm" onClick={handleDownloadStatement} disabled={downloadingStatement}>
                        <Download className="mr-2 h-4 w-4" /> Download Statement
                    </Button>
                    <Button className="bg-blue-600 hover:bg-blue-700 shadow-lg shadow-blue-500/20" onClick={() => void handleQuickPay()} disabled={!isPaymentConfigured || payingId !== null}>
                        <Zap className="mr-2 h-4 w-4" /> Quick Pay
                    </Button>
                </div>
            </div>

            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                <Card className="relative overflow-hidden border-2 border-blue-100 dark:border-blue-900 shadow-xl shadow-blue-500/5">
                    <div className="absolute top-0 right-0 p-4 opacity-10">
                        <DollarSign className="h-20 w-20 text-blue-600" />
                    </div>
                    <CardHeader>
                        <CardTitle className="text-sm font-medium uppercase tracking-wider text-muted-foreground">Outstanding Balance</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-4xl font-black text-slate-900 dark:text-slate-100 italic">
                            ₹{pendingTotal.toLocaleString()}
                        </div>
                        <div className="flex items-center gap-2 mt-4 text-sm font-medium text-orange-600 bg-orange-50 dark:bg-orange-950/30 px-3 py-1.5 rounded-full w-fit">
                            <AlertTriangle className="h-4 w-4" />
                            Due soon: ₹{assignments[0]?.amount?.toLocaleString() || 0}
                        </div>
                    </CardContent>
                </Card>

                <Card className="relative overflow-hidden border-2 border-green-100 dark:border-green-900 shadow-xl shadow-green-500/5">
                    <div className="absolute top-0 right-0 p-4 opacity-10">
                        < ShieldCheck className="h-20 w-20 text-green-600" />
                    </div>
                    <CardHeader>
                        <CardTitle className="text-sm font-medium uppercase tracking-wider text-muted-foreground">Total Paid</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-4xl font-black text-slate-900 dark:text-slate-100">
                            ₹{assignments.reduce((acc, a) => acc + a.paid_amount, 0).toLocaleString()}
                        </div>
                        <div className="flex items-center gap-2 mt-4 text-sm font-medium text-green-600 bg-green-50 dark:bg-green-950/30 px-3 py-1.5 rounded-full w-fit">
                            <CheckCircle2 className="h-4 w-4" />
                            {payments.length} Transactions Successful
                        </div>
                    </CardContent>
                </Card>

                <Card className="relative overflow-hidden border-2 border-purple-100 dark:border-purple-900 shadow-xl shadow-purple-500/5">
                    <div className="absolute top-0 right-0 p-4 opacity-10">
                        <PieChart className="h-20 w-20 text-purple-600" />
                    </div>
                    <CardHeader>
                        <CardTitle className="text-sm font-medium uppercase tracking-wider text-muted-foreground">Fee Exemptions</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-4xl font-black text-slate-900 dark:text-slate-100">
                            ₹{assignments.reduce((acc, a) => acc + a.waiver_amount, 0).toLocaleString()}
                        </div>
                        <p className="text-sm mt-4 text-muted-foreground">Scholarships and waivers applied</p>
                    </CardContent>
                </Card>
            </div>

            <Card className="border-none shadow-2xl bg-white/50 backdrop-blur-xl dark:bg-slate-950/50 overflow-hidden">
                <Tabs defaultValue="pending" className="w-full">
                    <div className="px-6 pt-6 border-b border-slate-100 dark:border-slate-800">
                        <TabsList className="bg-transparent gap-8 h-auto p-0">
                            <TabsTrigger
                                value="pending"
                                className="data-[state=active]:bg-transparent data-[state=active]:shadow-none data-[state=active]:border-b-2 data-[state=active]:border-blue-600 rounded-none px-0 pb-4 font-bold text-lg"
                            >
                                Pending Dues
                            </TabsTrigger>
                            <TabsTrigger
                                value="history"
                                className="data-[state=active]:bg-transparent data-[state=active]:shadow-none data-[state=active]:border-b-2 data-[state=active]:border-blue-600 rounded-none px-0 pb-4 font-bold text-lg"
                            >
                                Payment History
                            </TabsTrigger>
                        </TabsList>
                    </div>

                    <TabsContent value="pending" className="p-0 animate-in fade-in-50 duration-500">
                        <div className="grid gap-6 p-6">
                            {assignments.filter(a => a.status !== 'paid').map((assignment) => (
                                <Card key={assignment.id} className="group overflow-hidden border-slate-100 dark:border-slate-800 hover:border-blue-200 dark:hover:border-blue-900 transition-all duration-300 shadow-sm hover:shadow-md">
                                    <div className="flex flex-col md:flex-row">
                                        <div className="p-6 flex-1">
                                            <div className="flex items-center gap-3 mb-4">
                                                <div className="p-2 bg-blue-100 dark:bg-blue-900/40 rounded-lg">
                                                    <CreditCard className="h-5 w-5 text-blue-600" />
                                                </div>
                                                <div>
                                                    <h3 className="text-xl font-bold text-slate-900 dark:text-slate-100">{assignment.title}</h3>
                                                    <Badge variant="outline" className="mt-1 grayscale">{assignment.category}</Badge>
                                                </div>
                                            </div>

                                            <div className="grid grid-cols-2 lg:grid-cols-4 gap-6">
                                                <div>
                                                    <p className="text-xs text-muted-foreground uppercase font-black">Total Fee</p>
                                                    <p className="text-lg font-bold">₹{assignment.amount.toLocaleString()}</p>
                                                </div>
                                                <div>
                                                    <p className="text-xs text-muted-foreground uppercase font-black">Paid Already</p>
                                                    <p className="text-lg font-bold text-green-600">₹{assignment.paid_amount.toLocaleString()}</p>
                                                </div>
                                                <div>
                                                    <p className="text-xs text-muted-foreground uppercase font-black">Waiver</p>
                                                    <p className="text-lg font-bold text-orange-600">₹{assignment.waiver_amount.toLocaleString()}</p>
                                                </div>
                                                <div>
                                                    <p className="text-xs text-muted-foreground uppercase font-black">Due Date</p>
                                                    <p className="text-lg font-bold">
                                                        {format(new Date(assignment.due_date), 'MMM dd, yyyy')}
                                                    </p>
                                                </div>
                                            </div>
                                        </div>

                                        <div className="bg-slate-50 dark:bg-slate-900/50 p-8 flex flex-col items-center justify-center border-t md:border-t-0 md:border-l border-slate-100 dark:border-slate-800 min-w-[240px]">
                                            <p className="text-xs text-muted-foreground uppercase font-black mb-1">Payable Balance</p>
                                            <p className="text-3xl font-black mb-6">
                                                ₹{(assignment.amount - assignment.paid_amount - assignment.waiver_amount).toLocaleString()}
                                            </p>
                                            <Button
                                                onClick={() => initiatePayment(assignment)}
                                                disabled={payingId === assignment.id}
                                                className="w-full bg-blue-600 text-white hover:bg-blue-700 h-12 text-lg font-bold shadow-lg shadow-blue-500/20 group"
                                            >
                                                {payingId === assignment.id ? "Processing..." : "Pay Securely"}
                                                <ArrowRight className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform" />
                                            </Button>
                                        </div>
                                    </div>
                                </Card>
                            ))}
                            {assignments.filter(a => a.status !== 'paid').length === 0 && !loading && (
                                <div className="py-20 text-center space-y-4 bg-slate-50 dark:bg-slate-900/30 rounded-2xl mx-6 mb-6">
                                    <div className="inline-flex items-center justify-center p-4 bg-green-100 dark:bg-green-900/30 rounded-full mb-2">
                                        <CheckCircle2 className="h-10 w-10 text-green-600" />
                                    </div>
                                    <h3 className="text-2xl font-bold">You&apos;re all caught up!</h3>
                                    <p className="text-muted-foreground max-w-sm mx-auto">All your fees for the current term have been settled. Great job!</p>
                                </div>
                            )}
                        </div>
                    </TabsContent>

                    <TabsContent value="history" className="p-0 animate-in slide-in-from-bottom-5 duration-500">
                        <div className="p-6">
                            <div className="rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden shadow-sm">
                                <Table>
                                    <TableHeader className="bg-slate-50/50 dark:bg-slate-900/50">
                                        <TableRow>
                                            <TableHead className="font-bold">Transaction ID</TableHead>
                                            <TableHead className="font-bold">Fee Description</TableHead>
                                            <TableHead className="font-bold">Date</TableHead>
                                            <TableHead className="font-bold">Method</TableHead>
                                            <TableHead className="font-bold">Amount</TableHead>
                                            <TableHead className="font-bold">Status</TableHead>
                                            <TableHead className="text-right font-bold">Invoice</TableHead>
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {payments.map((payment) => (
                                            <TableRow key={payment.id} className="hover:bg-slate-50/30 dark:hover:bg-slate-900/30 transition-colors">
                                                <TableCell className="font-mono text-xs text-blue-600 font-bold">{payment.transaction_id || `TXN-${payment.id}`}</TableCell>
                                                <TableCell className="font-medium">{payment.fee_assignment_title || "Course Fee"}</TableCell>
                                                <TableCell className="text-sm">
                                                    {format(new Date(payment.payment_date), 'MMM dd, yyyy')}
                                                </TableCell>
                                                <TableCell className="capitalize text-sm font-medium">
                                                    <div className="flex items-center gap-2">
                                                        <CreditCard className="h-3 w-3 text-slate-400" />
                                                        {payment.payment_method}
                                                    </div>
                                                </TableCell>
                                                <TableCell className="font-bold">₹{payment.amount.toLocaleString()}</TableCell>
                                                <TableCell>
                                                    <Badge
                                                        variant={payment.payment_status === 'completed' ? 'default' : 'secondary'}
                                                        className={`px-3 py-1 rounded-full text-[10px] uppercase font-black tracking-tighter ${payment.payment_status === 'completed' ? 'bg-green-500 text-white border-none' : ''}`}
                                                    >
                                                        {payment.payment_status}
                                                    </Badge>
                                                </TableCell>
                                                <TableCell className="text-right">
                                                    <Button
                                                        variant="ghost"
                                                        size="sm"
                                                        className="h-8 w-8 p-0"
                                                        onClick={() => handleInvoiceDownload(payment)}
                                                        disabled={downloadingInvoiceId === payment.id}
                                                    >
                                                        <Download className="h-4 w-4" />
                                                    </Button>
                                                </TableCell>
                                            </TableRow>
                                        ))}
                                        {payments.length === 0 && !loading && (
                                            <TableRow>
                                                <TableCell colSpan={7} className="h-32 text-center text-muted-foreground italic">
                                                    No payment history found.
                                                </TableCell>
                                            </TableRow>
                                        )}
                                    </TableBody>
                                </Table>
                            </div>
                        </div>
                    </TabsContent>
                </Tabs>

                <CardFooter className="bg-slate-50 dark:bg-slate-950 px-8 py-6 flex flex-col md:flex-row items-center justify-between gap-4 border-t border-slate-100 dark:border-slate-800">
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        <ShieldCheck className="h-4 w-4 text-blue-600" />
                        Your payments are secured with industry-standard 256-bit encryption.
                    </div>
                    <div className="flex items-center gap-4">
                        <Image
                            src="https://upload.wikimedia.org/wikipedia/commons/8/89/Razorpay_logo.svg"
                            alt="Razorpay"
                            width={84}
                            height={16}
                            className="h-4 w-auto grayscale opacity-50 dark:invert"
                            unoptimized
                        />
                        <div className="h-4 w-px bg-slate-200 dark:bg-slate-800 mx-2" />
                        <p className="text-xs text-muted-foreground">© 2024 EduHub. All rights reserved.</p>
                    </div>
                </CardFooter>
            </Card>
        </div>
    );
}
