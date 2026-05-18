"use client";

import { Dialog as DialogPrimitive } from "@base-ui/react/dialog";
import { cn } from "@leros/ui/lib/utils";
import { XIcon } from "lucide-react";
import type * as React from "react";
import { Button } from "./button";

function Dialog({ ...props }: DialogPrimitive.Root.Props) {
	return <DialogPrimitive.Root data-slot="dialog" {...props} />;
}

function DialogTrigger({ ...props }: DialogPrimitive.Trigger.Props) {
	return <DialogPrimitive.Trigger data-slot="dialog-trigger" {...props} />;
}

function DialogClose({ ...props }: DialogPrimitive.Close.Props) {
	return <DialogPrimitive.Close data-slot="dialog-close" {...props} />;
}

function DialogContent({
	className,
	children,
	showCloseButton = true,
	...props
}: DialogPrimitive.Popup.Props & {
	showCloseButton?: boolean;
}) {
	return (
		<DialogPrimitive.Portal>
			<DialogPrimitive.Backdrop
				style={{
					position: "fixed",
					inset: 0,
					zIndex: 50,
					backgroundColor: "rgba(0, 0, 0, 0.4)",
				}}
				className="data-[starting-style]:opacity-0 data-[ending-style]:opacity-0 transition-all duration-200"
			/>
			<DialogPrimitive.Viewport
				style={{
					position: "fixed",
					inset: 0,
					zIndex: 50,
					display: "flex",
					alignItems: "center",
					justifyContent: "center",
					padding: "1rem",
				}}
			>
				<DialogPrimitive.Popup
					data-slot="dialog-content"
					className={cn(
						"relative w-full max-w-sm rounded-xl border bg-background p-6 shadow-lg",
						"data-[starting-style]:opacity-0 data-[ending-style]:opacity-0 data-[starting-style]:scale-95 data-[ending-style]:scale-95 transition-all duration-200",
						className,
					)}
					{...props}
				>
					{children}
					{showCloseButton && (
						<DialogPrimitive.Close
							data-slot="dialog-close"
							render={
								<Button
									variant="ghost"
									size="icon-sm"
									className="absolute top-4 right-4 h-6 w-6 opacity-70 hover:opacity-100"
								>
									<XIcon className="h-4 w-4" />
								</Button>
							}
						/>
					)}
				</DialogPrimitive.Popup>
			</DialogPrimitive.Viewport>
		</DialogPrimitive.Portal>
	);
}

function DialogHeader({ className, ...props }: React.ComponentProps<"div">) {
	return (
		<div data-slot="dialog-header" className={cn("flex flex-col gap-2", className)} {...props} />
	);
}

function DialogFooter({ className, ...props }: React.ComponentProps<"div">) {
	return (
		<div
			data-slot="dialog-footer"
			className={cn("flex flex-col-reverse gap-2 sm:flex-row sm:justify-end", className)}
			{...props}
		/>
	);
}

function DialogTitle({ className, ...props }: DialogPrimitive.Title.Props) {
	return (
		<DialogPrimitive.Title
			data-slot="dialog-title"
			className={cn("text-lg font-semibold leading-none tracking-tight", className)}
			{...props}
		/>
	);
}

function DialogDescription({ className, ...props }: DialogPrimitive.Description.Props) {
	return (
		<DialogPrimitive.Description
			data-slot="dialog-description"
			className={cn("text-sm text-muted-foreground", className)}
			{...props}
		/>
	);
}

function DialogViewport({ ...props }: DialogPrimitive.Viewport.Props) {
	return <DialogPrimitive.Viewport data-slot="dialog-viewport" {...props} />;
}

function DialogPortal({ ...props }: DialogPrimitive.Portal.Props) {
	return <DialogPrimitive.Portal data-slot="dialog-portal" {...props} />;
}

export {
	Dialog,
	DialogClose,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogPortal,
	DialogTitle,
	DialogTrigger,
	DialogViewport,
};
