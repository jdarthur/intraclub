import * as React from "react"

type WinLossIconProps = {
    win: boolean
    sizeEm: number
}

export function WinLossIcon({win, sizeEm}: WinLossIconProps) {

    let char = "✕"
    if (win) {
        char = "✔"
    }

    return <div style={{
        background: "rgba(0, 0, 0, 0.3)",
        width: `${sizeEm}em`,
        height: `${sizeEm}em`,
        borderRadius: `${sizeEm / 2}em`,
        textAlign: "center",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        color: "white",
    }}>
        <span style={{fontSize: win ? "0.7em" : '0.5em'}}>
            {char}
        </span>
    </div>
}