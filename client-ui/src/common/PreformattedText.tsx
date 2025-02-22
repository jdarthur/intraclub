import * as React from "react";
import {Space} from "antd";

type PreformattedTextProps = {
    text: string;
    maxHeight?: number;
    lineNumbers?: boolean
}

const lineNumberStyle = {
    color: "#8c8c8c",
    // borderRight: "1px solid #8c8c8c",
    marginRight: "1em",
    padding: "0.5em",
    marginLeft: "-1em",
    alignSelf: "flex-end"
}

export function PreformattedText({text, maxHeight, lineNumbers}: PreformattedTextProps) {

    const lines = text.split('\n')

    const content = lines.map((line, i) =>
        <div>
            {/*{lineNumbers ? <span style={lineNumberStyle}>{i}</span> : null}*/}
            {line}
        </div>
    )

    const lineNumbersContent = lines.map((line, i) =>
        <div>
            {lineNumbers ? <span style={lineNumberStyle}>{i}</span> : null}
        </div>
    )

    const style = {
        fontFamily: "monospace",
        background: "#bfbfbf",
        padding: "0.5em 1.5em",
        borderRadius: 5,
        display: "flex",
        overflow: "clip"
    }

    if (maxHeight) {
        style["maxHeight"] = maxHeight
    }

    return <pre style={style}>
        <div style={{display: "flex", alignItems: "flex-end", flexDirection: "column"}}>{lineNumbersContent}</div>
        <div>{content}</div>
    </pre>
}