// In StudentMessages.tsx
"use client"
import React from 'react'
import { useState } from 'react'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { ChevronDown, ChevronUp } from 'lucide-react'

export default function StudentMessages() {
  const [toggle, setToggle] = useState(false)
  
  return (
    <Card className={`w-full h-full flex flex-col ${
      toggle ? 'absolute bottom-4 h-[500px] right-4 w-[380px] ' : 'min-h-0'
    }`}>
      <CardHeader className='p-2 flex flex-row items-center justify-between'>
        <CardTitle className='font-medium text-[1.25rem]'>Messages</CardTitle>
        <Button 
          variant="ghost" 
          size="sm" 
          onClick={() => setToggle(!toggle)}
          className="ml-auto"
        >
          {toggle ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
        </Button>
      </CardHeader>
      
      <div 
        className={`overflow-hidden transition-all duration-300 ease-in-out ${
          toggle ? '' : 'min-h-0'
        }`}
      >
        <CardContent className={`p-2 ${
          toggle ? 'min-h-[50vh]' : 'min-h-0 hidden'
        }`}>
          <div className="space-y-2">
            <p className="text-sm">You have no new messages.</p>

          </div>
        </CardContent>
      </div>
    </Card>
  )
}