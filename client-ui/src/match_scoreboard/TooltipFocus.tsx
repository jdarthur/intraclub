import * as React from "react";
import "./scoreboard.css"

type TooltipFocusProps = {
    open: boolean
    zIndex: number
    onClose: () => void
}

export function TooltipFocus({open, zIndex, onClose}: TooltipFocusProps) {

    if (!open) {
        return null
    }

    return <div
        style={{
            width: "100vw",
            height: "100vh",
            position: "fixed",
            top: 0,
            left: 0,
            background: "rgba(0, 0, 0, 0.25)",
            animation: "color 0.5s",
            zIndex: zIndex,
        }}
        onClick={onClose}
    />
}