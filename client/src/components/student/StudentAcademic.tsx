import React from 'react'
import { Card, CardHeader, CardTitle } from '@/components/ui/card'
export default function StudentAcademic() {
  return (
    <div className='w-full flex flex-col gap-2  min-h-full'>
        <Card className="p-4">
            <CardTitle className='font-medium flex items-center justify-center text-[1.25rem]'>
                75% Attendance
            </CardTitle>
        </Card>
        <Card className="p-4">
            <CardTitle className='font-medium flex items-center justify-center text-[1.25rem]'>
                $0 Dues
            </CardTitle>
        </Card>
        <Card className="p-4">
            <CardTitle className='font-medium flex items-center justify-center text-[1.25rem]'>
                8.5 CGPA
            </CardTitle>
        </Card>
    </div>
  )
}
