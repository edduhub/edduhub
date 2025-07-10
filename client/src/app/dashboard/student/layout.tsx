import Sidebar from "@/components/student/StudentSidebar";
import Navbar from "../../../components/student/navbar";
import { ThemeProvider } from "@/components/theme/ThemeProvider";
import RSidebar from '../../../components/student/StudentRightSidebar';
export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  /// here for data is a missing attribute the reason it's occuring is cause the details that are needed to be there weren't given and
  /// we have to wait for integration of backend  
  return (
    <html lang="en" suppressHydrationWarning>
      <body suppressHydrationWarning >
        <ThemeProvider
                    attribute="class"
                    defaultTheme="system"
                    enableSystem
                    disableTransitionOnChange
                  >
        <div className="flex max-h-screen overflow-hidden p-2">
          <div className=" lg:w-[18vw] max-lg:hidden flex-shrink-0">
            <Sidebar />
          </div>
          <div className="flex flex-col flex-grow overflow-hidden">
            <div className="flex-none max-lg:flex lg:mt-3 px-1">
              <div className="lg:hidden">
                <Sidebar />
              </div>
              <Navbar />
            </div>
            <div className="flex flex-grow overflow-hidden">
              <div className="flex-grow overflow-auto">
                {children}
              </div>
              <div className="max-lg:hidden w-[20vw]">
                <RSidebar />
              </div>
            </div>
          </div>
        </div>
        </ThemeProvider>
      </body>
    </html>
  );
}