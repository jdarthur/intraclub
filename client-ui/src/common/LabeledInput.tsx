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
    onEnter?: () => void
}

function LabeledInput({label, value, setValue, style, placeholder, disabled, onEnter}: LabeledInputArgs) {
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

    const onHitEnter = () => {
        if (onEnter) {
            onEnter()
        }
    }

    return (
        <div style={s}>
            <span style={{marginRight: "1em", fontWeight: "bold", fontSize: "0.9em", marginBottom: "0.2em"}}>
                {label}
            </span>
            <Input value={value} onChange={(e) => setValue(e.target.value)} placeholder={actualPlaceholder}
                   disabled={disabled} onPressEnter={onHitEnter}/>
        </div>

    )
}

export default LabeledInput
