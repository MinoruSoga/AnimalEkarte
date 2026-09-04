import * as React from "react";

import { cn } from "./utils";

interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  ref?: React.Ref<HTMLDivElement>;
}

function Card({ className, ref, ...props }: CardProps) {
  return (
    <div
      ref={ref}
      data-slot="card"
      className={cn(
        "bg-card text-card-foreground flex flex-col gap-6 rounded-lg border",
        className,
      )}
      {...props}
    />
  );
}

function CardHeader({ className, ref, ...props }: CardProps) {
  return (
    <div
      ref={ref}
      data-slot="card-header"
      className={cn(
        "@container/card-header grid auto-rows-min grid-rows-[auto_auto] items-start gap-1.5 px-6 pt-6 has-data-[slot=card-action]:grid-cols-[1fr_auto] [.border-b]:pb-6",
        className,
      )}
      {...props}
    />
  );
}

interface CardTitleProps extends React.HTMLAttributes<HTMLHeadingElement> {
  ref?: React.Ref<HTMLParagraphElement>;
}

function CardTitle({ className, ref, ...props }: CardTitleProps) {
  return (
    <h4 ref={ref} data-slot="card-title" className={cn("leading-none", className)} {...props} />
  );
}

function CardContent({ className, ref, ...props }: CardProps) {
  return (
    <div
      ref={ref}
      data-slot="card-content"
      className={cn("px-6 [&:last-child]:pb-6", className)}
      {...props}
    />
  );
}

export { Card, CardHeader, CardTitle, CardContent };
