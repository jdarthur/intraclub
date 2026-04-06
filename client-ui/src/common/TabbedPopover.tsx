import {Popover, Radio} from "antd";
import React from "react";
import {InfoCircleOutlined} from "@ant-design/icons";

export type TabbedPopoverTab = {
    title: string,
    content: React.ReactNode
}

type TabbedPopoverProps = {
    style?: any,
    title?: any,
    Tabs: TabbedPopoverTab[]
}

export function TabbedPopover({style, Tabs, title}: TabbedPopoverProps) {

    const [selected, setSelected] = React.useState<string>(Tabs[0].title)

    const body = Tabs?.map((t) => {
        return <div>
            {selected === t.title ? t.content : null}
        </div>
    })

    const radioButtons = Tabs?.map((t) => {
        return <Radio.Button value={t.title} onClick={() => setSelected(t.title)}>{t.title}</Radio.Button>
    })

    const content = <div style={style}>
        <Radio.Group buttonStyle={"solid"} value={selected}
                     onChange={(e) => setSelected(e.target.value)}>
            {radioButtons}
        </Radio.Group>
        <div style={{marginTop: "1em"}} />
        {body}
    </div>

    return <Popover title={title} content={content} placement={"right"}>
        <InfoCircleOutlined/>
    </Popover>
}

