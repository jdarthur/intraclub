import {Card} from "antd";
import React from "react";
import {LabeledValue} from "../common/LabeledValue";
import {DeleteConfirm} from "../common/DeleteConfirm";
import {useDeleteSkillInfoMutation} from "../redux/api";

export type SkillInfoRecord = {
    league_type: string
    most_recent_year: number
    captain: string
    level: string
    line: string
    id?: string
}

export type SkillInfoBody = SkillInfoRecord & {
    user_id: string
}

export function SkillInfo({...props}: SkillInfoRecord) {

    const [deleteSkillInfo] = useDeleteSkillInfoMutation()

    const onDelete = () => {
        console.log(`Delete skill_info ${props.id}`)
        deleteSkillInfo(props.id)
    }

    const extra = <DeleteConfirm deleteFunction={onDelete} objectType={"Experience info"}/>

    return <Card size={"small"} title={`${props.league_type} (${props.most_recent_year})`}
                 style={{width: 250, margin: "0.5em"}} extra={extra}>
        <LabeledValue label={"Captain"} value={props.captain}/>
        <LabeledValue label={"Level"} value={props.level}/>
        <LabeledValue label={"Line"} value={props.line}/>
    </Card>
}