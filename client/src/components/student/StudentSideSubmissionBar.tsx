import { Card, CardHeader, CardTitle } from '@/components/ui/card'
import React from 'react'

export default function SideSubmissionBar() {
  return (
    <Card className='w-full h-full'>
      <CardHeader className=''>
        <CardTitle className='font-medium text-[1.25rem]'>Submissions</CardTitle>
      </CardHeader>
    </Card>
  )
}
