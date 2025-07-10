import Homepage from "@/components/Homepage";
import Image from "next/image";
import Login from "@/components/Login";
//still need to redirect the button to the pages and need to integrate backend for the getting authentication
export default function Home() {
  return (
    <div className="max-h-screen font-[family-name:var(--font-geist-sans)]">
      
      <Login/>
    </div>
  );
}
