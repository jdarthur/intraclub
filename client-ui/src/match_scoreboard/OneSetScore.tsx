import {Popover} from "antd";
import * as React from "react";
import {CloseOutlined} from "@ant-design/icons";
import {useState} from "react";
import {TooltipFocus} from "./TooltipFocus";

type SetScoreEditorProps = {
    value: number
    setValue: (v: number) => void
    max: number
}

type OneSetScoreProps = SetScoreEditorProps & {
    onSave: () => void
    readOnly?: boolean
}

function SetScoreEditor({value, setValue, max}: SetScoreEditorProps) {

    const onClick = () => {
        if (value == max) {
            setValue(0)
        } else {
            setValue(value + 1)
        }
    }

    return <span style={{fontSize: "2em", margin: "1em", cursor: "pointer"}} onClick={onClick}>
        {value}
    </span>
}

export function OneSetScore({value, setValue, max, onSave, readOnly}: OneSetScoreProps) {

    const content = <SetScoreEditor
        value={value} setValue={setValue} max={max}
    />

    const [open, setOpen] = useState<boolean>(false)

    const onClose = () => {
        console.log("Update set score: ", value)
        setOpen(false)
        onSave()
    }

    const v = readOnly ? value : <span style={{cursor: "pointer"}}>{value}</span>

    const displayValue = <div
        style={{
            width: "100%",
            display: "flex",
            justifyContent: "center",
            cursor: readOnly ? "auto" : "pointer",
        }}
        onClick={() => setOpen(true)}
    >
        {v}
    </div>

    if (readOnly) {
        return displayValue
    }

    const title = <span
        style={{fontSize: "1.4em", display: "flex", alignItems: "center", justifyContent: "space-between"}}>
        Games won
        <CloseOutlined onClick={onClose} style={{color: "rgba(0, 0, 0, 0.4)"}}/>
    </span>



    return <div>
        <TooltipFocus open={open} zIndex={3} onClose={onClose} />
        <Popover title={title} content={content} trigger={"click"} overlayStyle={{width: 200}} open={open} zIndex={4}>
            {displayValue}
        </Popover>
    </div>


}