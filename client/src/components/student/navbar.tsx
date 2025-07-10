import { Card } from '@/components/ui/card'
import React from 'react'
import Link from 'next/link'
import { Button } from '../ui/button'
import Image from 'next/image'
import { ThemeToggle } from '../theme/ThemeToggle'
import downWhite from "@assets/downWhite.svg"
export default function Navbar() {
  return (
    <div className="w-full flex justify-between max-lg:justify-end gap-4 pb-2">
      <Card className='w-full max-lg:hidden sticky rounded-full lg:py-0 flex items-center justify-between px-4'>
        <div className='flex px-4 font-normal text-[1.25rem] gap-8'>
          <Link href="/">Fees</Link>
          <Link href="/">Exams</Link>
          <Link href="/">Placement</Link>
          <Link href="/">Documents</Link>
          <Link href="/">Raise Ticket</Link>
        </div>
        <div className="w-10 h-10 bg-black rounded-full">
          <Image src={downWhite} alt="down" className='w-10 h-10 rounded-full' />
        </div>
      </Card>
      <div className="flex items-center gap-2">
        {/* if we change height of these circles then the height of the navbar changes */}
        <ThemeToggle />
        <div className="h-12 bg-black w-12 rounded-full"></div>
        <div className="h-12 bg-black w-12 rounded-full"></div>
      </div>
    </div>
  )
}
