import React from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export default function StudentAnnouncement() {
  return (
    <Card className='w-full h-full flex flex-col'>
      <CardHeader className='pb-2'>
        <CardTitle className='font-medium text-[1.25rem]'>Announcements</CardTitle>
      </CardHeader>
      <CardContent className='flex-grow'>
        {/* Content will go here, but the card will expand to fill space even when empty */}
      </CardContent>
    </Card>
  )
}