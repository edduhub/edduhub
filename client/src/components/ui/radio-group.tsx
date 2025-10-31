"use client";

import React from "react";

type RadioOption = {
  label: string;
  value: string | number;
  description?: string;
  disabled?: boolean;
};

type RadioGroupProps = {
  name: string;
  value: string | number | undefined;
  onChange: (value: string | number) => void;
  options: RadioOption[];
  className?: string;
  direction?: "vertical" | "horizontal";
};

export function RadioGroup({ name, value, onChange, options, className, direction = "vertical" }: RadioGroupProps) {
  return (
    <div className={className} role="radiogroup" aria-labelledby={`${name}-label`}>
      <div className={direction === "horizontal" ? "flex flex-wrap gap-3" : "space-y-2"}>
        {options.map((opt) => {
          const id = `${name}-${String(opt.value)}`;
          return (
            <label key={id} htmlFor={id} className="flex cursor-pointer items-start gap-2">
              <input
                id={id}
                type="radio"
                name={name}
                className="mt-0.5 h-4 w-4 cursor-pointer accent-primary"
                checked={String(value) === String(opt.value)}
                onChange={() => onChange(opt.value)}
                value={String(opt.value)}
                disabled={opt.disabled}
              />
              <div className="space-y-1">
                <span className="text-sm font-medium leading-none">{opt.label}</span>
                {opt.description && (
                  <p className="text-xs text-muted-foreground">{opt.description}</p>
                )}
              </div>
            </label>
          );
        })}
      </div>
    </div>
  );
}

export default RadioGroup;

"use client"

import * as React from "react"
import * as RadioGroupPrimitive from "@radix-ui/react-radio-group"
import { Circle } from "lucide-react"

import { cn } from "@/lib/utils"

const RadioGroup = React.forwardRef<
  React.ElementRef<typeof RadioGroupPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof RadioGroupPrimitive.Root>
>(({ className, ...props }, ref) => {
  return (
    <RadioGroupPrimitive.Root
      className={cn("grid gap-2", className)}
      {...props}
      ref={ref}
    />
  )
})
RadioGroup.displayName = RadioGroupPrimitive.Root.displayName

const RadioGroupItem = React.forwardRef<
  React.ElementRef<typeof RadioGroupPrimitive.Item>,
  React.ComponentPropsWithoutRef<typeof RadioGroupPrimitive.Item>
>(({ className, ...props }, ref) => {
  return (
    <RadioGroupPrimitive.Item
      ref={ref}
      className={cn(
        "aspect-square h-4 w-4 rounded-full border border-primary text-primary ring-offset-background focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
        className
      )}
      {...props}
    >
      <RadioGroupPrimitive.Indicator className="flex items-center justify-center">
        <Circle className="h-2.5 w-2.5 fill-current text-current" />
      </RadioGroupPrimitive.Indicator>
    </RadioGroupPrimitive.Item>
  )
})
RadioGroupItem.displayName = RadioGroupPrimitive.Item.displayName

export { RadioGroup, RadioGroupItem }
