import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function cx(...inputs: ClassValue[]) {
  return clsx(inputs)
}

export function focusRing(
  className: string,
  focusRingClassName: string,
  focusRingOpacity: string = "50",
) {
  return cx(
    className,
    `focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-white focus-visible:ring-${focusRingClassName}/${focusRingOpacity}`,
  )
}

export function focusInput(className: string) {
  return focusRing(className, "blue", "30")
}

export function hasErrorInput(className: string) {
  return focusRing(className, "red", "30")
}

