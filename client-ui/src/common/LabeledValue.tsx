import * as React from 'react';

type LabeledValueArgs = {
    label: string
    value: React.ReactNode
    style?: object
    vertical?: boolean
}

export function LabeledValue({label, value, vertical}: LabeledValueArgs) {
    return <span style={{
        display: "flex",
        justifyContent: "space-between",
        flexDirection: vertical ? "column" : "row",
        paddingBottom: "0.5em"
    }}>
        <span style={{fontSize: '0.9em'}}>{label}:</span>
        <span style={{padding: "0 1em", fontSize: "1.1em"}}>{value}</span>
    </span>
}