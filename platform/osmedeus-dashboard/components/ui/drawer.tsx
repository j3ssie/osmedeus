"use client";

import * as React from "react";
import {
  Sheet,
  SheetTrigger,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";

type DrawerDirection = "top" | "right" | "bottom" | "left";

const DrawerDirectionContext = React.createContext<DrawerDirection>("bottom");

function Drawer({
  direction = "bottom",
  ...props
}: React.ComponentProps<typeof Sheet> & { direction?: DrawerDirection }) {
  return (
    <DrawerDirectionContext.Provider value={direction}>
      <Sheet {...props} />
    </DrawerDirectionContext.Provider>
  );
}

function DrawerTrigger({
  ...props
}: React.ComponentProps<typeof SheetTrigger>) {
  return <SheetTrigger {...props} />;
}

function DrawerClose({
  ...props
}: React.ComponentProps<typeof SheetClose>) {
  return <SheetClose {...props} />;
}

function DrawerContent({
  ...props
}: React.ComponentProps<typeof SheetContent>) {
  const direction = React.useContext(DrawerDirectionContext);
  return <SheetContent side={direction} {...props} />;
}

function DrawerHeader({
  ...props
}: React.ComponentProps<typeof SheetHeader>) {
  return <SheetHeader {...props} />;
}

function DrawerFooter({
  ...props
}: React.ComponentProps<typeof SheetFooter>) {
  return <SheetFooter {...props} />;
}

function DrawerTitle({
  ...props
}: React.ComponentProps<typeof SheetTitle>) {
  return <SheetTitle {...props} />;
}

function DrawerDescription({
  ...props
}: React.ComponentProps<typeof SheetDescription>) {
  return <SheetDescription {...props} />;
}

export {
  Drawer,
  DrawerTrigger,
  DrawerClose,
  DrawerContent,
  DrawerHeader,
  DrawerFooter,
  DrawerTitle,
  DrawerDescription,
};
