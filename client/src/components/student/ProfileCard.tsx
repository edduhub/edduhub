import React from 'react'
import { Avatar } from '@/components/ui/avatar'
import {
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
} from "@/components/ui/card"
export default function ProfileCard() {
    return (
        <Card className='w-full '>
            <CardHeader className=''>
                <div className="flex items-center  gap-2">
                <div className="">
                    <Avatar className='w-16 h-16' />
                </div>
                <div className="flex flex-col">
                    <CardTitle className=' font-medium text-[1.25rem]'>Rakesh nayak</CardTitle>
                    <CardDescription className=''>BTECH-CSE</CardDescription>
                    <CardDescription className=''>SEM-X</CardDescription>
                </div>
                </div>
            </CardHeader>
        </Card>
    )
}
