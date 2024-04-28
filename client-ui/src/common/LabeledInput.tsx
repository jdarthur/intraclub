import {useState} from 'react'
import * as React from 'react'
import {Input} from "antd";

type LabeledInputArgs = {
    label: string,
    value: string,
    setValue: (arg0: string) => any,
    style?: {},
    placeholder?: string,
    disabled?: boolean
}

function LabeledInput({label, value, setValue, style, placeholder, disabled}: LabeledInputArgs) {
    const s = {display: "flex", flexDirection: 'column' as 'column', alignItems: "flex-start"}

    if (style) {
        for (const [key, value] of Object.entries(style)) {
            s[key] = value
        }
    }


    let actualPlaceholder = placeholder
    if (!placeholder) {
        actualPlaceholder = label
    }

    return (
        <div style={s}>
            <span style={{marginRight: "1em", fontWeight: "bold", fontSize: "0.9em", marginBottom: "0.2em"}}>
                {label}
            </span>
            <Input value={value} onChange={(e) => setValue(e.target.value)} placeholder={actualPlaceholder}
                   disabled={disabled}/>
        </div>

    )
}

export default LabeledInput
