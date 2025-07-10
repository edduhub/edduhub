import React from 'react'
import Link from 'next/link'
import {
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
} from "@/components/ui/card"
export default function SideNavigationBar() {
  return (
    <Card className='w-full'>
        <div className='flex p-4 font-medium text-[1.25rem] flex-col gap-2'>
          <Link href="/">Dashboard</Link>
          <Link href="/">Profile</Link>
          <Link href="/dashboard/student/calendar">Calendar</Link>
          <Link href="/">Timetable</Link>
          <Link href="/">Hostel</Link>
        </div>
    </Card>
  )
}
