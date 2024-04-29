import * as React from 'react';
import {Tooltip} from "antd";

type EllipsisProps = {
    fullValue: string
    maxLength?: number
}

export function Ellipsis({fullValue, maxLength}: EllipsisProps) {
    let thisMaxLength = 40
    if (maxLength) {
        thisMaxLength = maxLength
    }

    let displayedValue = fullValue
    let tooltip = null

    if (fullValue.length > thisMaxLength) {
        displayedValue = fullValue.substring(0, thisMaxLength) + "..."
        tooltip = <Tooltip title={fullValue}>
            {displayedValue}
        </Tooltip>
    }

    return <span>
        {tooltip ? tooltip : fullValue}
    </span>
}