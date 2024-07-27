import {InputNumber, Popover} from "antd";
import * as React from "react";

type OneSetScoreProps = {
    value: number
    setValue: (v: number) => void
    max: number
    onSave: () => void
    readOnly?: boolean
}

export function OneSetScore({value, setValue, max, onSave, readOnly}: OneSetScoreProps) {

    const content = <InputNumber size={"large"}
                                 value={value} onChange={(v) => setValue(v)}
                                 min={0} max={max}
                                 style={{width: "100%", fontSize: "1.4em"}}/>

    const onOpenChange = (open: boolean) => {
        if (!open) {
            console.log("close the set score popover")
            onSave()
        }
    }

    const v = readOnly ? value : <span style={{cursor: "pointer"}}>{value}</span>

    const displayValue = <div style={{
        width: "100%",
        display: "flex",
        justifyContent: "center",
        padding: "0.5em",
        cursor: readOnly ? "auto" : "pointer",
    }}>
        {v}
    </div>

    if (readOnly) {
        return displayValue
    }

    const title = <span style={{fontSize: "1.4em"}}>Games won</span>

    return <Popover title={title} content={content} trigger={"click"} overlayStyle={{width: 200}}
                    onOpenChange={onOpenChange}>
        {displayValue}
    </Popover>

}