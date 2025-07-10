import Link from 'next/link'
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


export default function Login() {
  return (
    <div className='min-w-screen min-h-screen flex justify-center items-center font-[family-name:var(--font-geist-sans)]'>
      <div className="w-1/2 min-h-screen max-lg:hidden" style={{ backgroundImage: `url(${loginimg.src})`, backgroundSize: "cover", backgroundPosition: "center" }}></div>
      <div className="w-1/2 flex flex-col max-lg:min-w-full min-h-screen justify-center items-center">
        <h1 className="text-[2rem] md:text-[2.5rem] font-bold text-center mb-6 text-black dark:text-white">Welcome to Edduhub</h1>
        <Card className='w-[350px]'>
          <CardHeader>
            <CardTitle>Login</CardTitle>
            <CardDescription>
              Enter your credentials to access your account
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form action="">
              <div className="grid w-full items-center gap-4">
                <div className="">
                  <label htmlFor="email">Email</label>
                  <Input id="email" type="email" placeholder="Enter your email" />
                </div>
                <div className="">
                  <label htmlFor="password">Password</label>
                  <Input id="password" type="password" placeholder="Enter your password" />
                </div>
              </div>
            </form>
          </CardContent>
          <CardFooter>
            <div className="flex flex-col gap-4 w-full">
              <Button className='w-full' variant="default">Login</Button>
              <Link href="/forgot-password" className='w-full text-center'>Forgotten your password?</Link>
            </div>
          </CardFooter>
        </Card>
      </div>
    </div>
  )
}
