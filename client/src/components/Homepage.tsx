import Link from 'next/link'
import React from 'react'

export default function Homepage() {
  return (
    <div className='min-w-screen min-h-screen p-16 max-md:p-8 gap-16 font-[family-name:var(--font-geist-sans)]'>
        <div className="min-w-screen flex flex-col max-md:text-center items-center justify-center">
            <h1 className="text-[3rem] max-md:text-[2rem] font-semibold">Transform Education with Our Smart ERP Solution</h1>
            <Link href="/" className='p-2 my-4 text-[#000] bg-[#fff] hover:text-[#fff] hover:bg-[#000] rounded-lg hover:border-spacing-1 hover:border border border-spacing-1 border-black hover:border-white'>Request Demo</Link>
        </div>
    </div>
  )
}
