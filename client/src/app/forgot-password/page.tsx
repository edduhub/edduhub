import React from 'react'
import { Button } from "@/components/ui/button"
import {
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import loginimg from "@images/login.jpg"
import Link from 'next/link'
export default function page() {
    return (
        <div className="min-h-[100dvh] flex justify-center overflow-hidden min-w-screen font-[family-name:var(--font-geist-sans)]">
            <div className="w-1/2 min-h-[100dvh] max-lg:hidden" style={{ backgroundImage: `url(${loginimg.src})`, backgroundSize: "cover", backgroundPosition: "center" }}></div>
            <div className="w-1/2 flex flex-col max-lg:min-w-full overflow-hidden justify-center items-center">
                <Card className='w-[350px]'>
                    <CardHeader>
                        <CardTitle>Forgot Password</CardTitle>
                        <CardDescription>
                            Enter your email to reset your password
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <form action="">
                            <div className="grid w-full items-center gap-4">
                                <div className="">
                                    <label htmlFor="email">Email</label>
                                    <Input id="email" type="email" placeholder="Enter your email" />
                                </div>
                            </div>
                        </form>
                    </CardContent>
                    <CardFooter>
                        <div className="flex flex-col gap-4 w-full">
                            <Button className='w-full' variant="default">Reset Password</Button>
                        </div>
                    </CardFooter>
                </Card>
            </div>
        </div>
    )
}
