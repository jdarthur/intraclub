import * as React from "react"
import {Input, Popover, Space} from "antd";
import {useUpdateNameForLineMutation} from "../redux/api.js";
import {useSearchParams} from "react-router-dom";
import {useEffect} from "react";


export type PlayerProps = {
    line: number
    name: string
}

export type stringEditorDisplayType = {
    value: string
    setValue: (v: string) => void
    onSave: () => void
    readOnly: boolean
}

function NameDisplay({value, setValue, onSave, readOnly}: stringEditorDisplayType) {
    const onOpenChange = (open: boolean) => {
        if (!open) {
            console.log("close the set player name popover")
            onSave()
        }
    }

    let nameValue = value ? value : <span style={{color: "#bfbfbf"}}>Name not set</span>
    const name = <span style={{cursor: readOnly ? "auto" : "pointer", fontSize: "1.5em"}}>
        {nameValue}
    </span>

    if (readOnly) {
        return name
    }

    const onChange = (event: any) => {
        setValue(event.target.value)
    }

    const content = <Input value={value}
                           onChange={onChange}
                           ref={el => {
                               setTimeout(() => el?.focus(), 0);
                           }}
    />

    return <Popover title={"Set player name"}
                    content={content}
                    onOpenChange={onOpenChange}
                    trigger={"click"}>
        {name}
    </Popover>
}

function LineNumber({
                        line
                    }
                        :
                        PlayerProps
) {
    return <div style={{
        color: "rgba(0, 0, 0, 0.5)",
        border: "1px solid rgba(0, 0, 0, 0.5)",
        width: 30,
        height: 30,
        borderRadius: 15,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        fontSize: "1.4em",
        fontWeight: "bold"
    }}>
        {line}
    </div>
}

type PlayerDisplayProps = {
    matchup_line: string
    player1: boolean
    player_line: number
    home: boolean
    initialName: string
}

export function Player({matchup_line, player1, player_line, home, initialName}: PlayerDisplayProps) {
    const [name, setName] = React.useState<string>(initialName)
    const [updateName] = useUpdateNameForLineMutation()

    const [searchParams] = useSearchParams()
    const key = searchParams.get('key')

    useEffect(() => {
        setName(initialName)

    }, [initialName])

    const onSave = () => {
        const body = {
            player1: player1,
            matchup_line: matchup_line,
            name: name,
            home: home,
            key: key
        }

        console.log("Save player name: ", body)
        updateName(body)
    }

    return <Space>
        <LineNumber line={player_line} name={""}/>
        <NameDisplay value={name} setValue={setName} onSave={onSave} readOnly={!key}/>
    </Space>
}