"use client";

import { Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";
import { Button } from "@/components/ui/button";

export function ThemeToggle() {
  const { theme, setTheme } = useTheme();

  return (
    <Button
      variant="ghost"
      size="icon"
      className="rounded-full h-10 w-10 bg-black dark:bg-white hover:bg-gray-800 dark:hover:bg-gray-100"
      onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
    >
      <Sun className="h-8 w-8 rotate-0 scale-0 transition-all dark:rotate-0 dark:scale-100 text-white dark:text-black" />
      <Moon className="absolute h-8 w-8 rotate-0 scale-100 transition-all dark:rotate-90 dark:scale-0 text-white dark:text-black" />
      <span className="sr-only">Toggle theme</span>
    </Button>
  );
}