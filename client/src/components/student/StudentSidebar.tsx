import React from 'react'
import ProfileCard from './ProfileCard'
import SideNavigationBar from './StudentSideNavigationBar'
import SideSubmissionBar from './StudentSideSubmissionBar'
import { Avatar } from '../ui/avatar'
import { Sheet, SheetContent, SheetTrigger, SheetClose, SheetHeader, SheetTitle } from '../ui/sheet'
import { X } from 'lucide-react'
import { Button } from '../ui/button'

//need to have pass data through the props so keep in mind we look into this
export default function Sidebar({ data }: Readonly<{ data?: JSON }>) {
  
  return (
    <div className='w-full lg:min-h-screen flex flex-col gap-2 p-2 text-white lg:h-screen'>
      {/* Mobile/tablet view with toggle functionality */}
      <div className="lg:hidden">
        <Sheet>
          <SheetTrigger asChild>
            <Avatar className='w-16 h-16 cursor-pointer hover:opacity-80 transition-opacity' />
          </SheetTrigger>
          <SheetHeader className='hidden'><SheetTitle></SheetTitle></SheetHeader>
          <SheetContent side="left" className="max-lg:w-[80vw] max-md:w-[100vw] p-0">
            <h2 className="sr-only">Student Navigation</h2>
            <div className="flex flex-col h-full p-4 bg-background">
              <div className="flex justify-end mb-4">
              </div>
              <div className="mb-4">
                <ProfileCard />
              </div>
              <div className="mb-4">
                <SideNavigationBar />
              </div>
              <div className="flex-grow pb-2">
                <SideSubmissionBar />
              </div>
            </div>
          </SheetContent>
        </Sheet>
      </div>
      
      {/* Desktop view - unchanged */}
      <div className="w-full max-lg:hidden min-h-screen flex flex-col gap-2 p-2 text-white h-screen">
        <div className="">
          <ProfileCard />
        </div>
        <div className="">
          <SideNavigationBar />
        </div>
        <div className="h-full pb-2 mb-2">
          <SideSubmissionBar />
        </div>
      </div>
    </div>
  )
}