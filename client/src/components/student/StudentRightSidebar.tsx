
import React from 'react'
import StudentAcademic from './StudentAcademic'
import StudentAnnouncement from './StudentAnnouncement'
import StudentMessages from './StudentMessages'

export default function RSidebar() {
  return (
    <div className='w-full flex flex-col h-full max-h-screen px-2'>
      {/* Academic section with auto height */}
      <div className="flex-none mb-2">
        <StudentAcademic />
      </div>
      
      {/* Announcements section that fills available space */}
      <div className="flex-grow mb-2">
        <StudentAnnouncement />
      </div>
      
      {/* Messages section with fixed height */}
      <div className="flex-none">
        <StudentMessages />
      </div>
    </div>
  )
}