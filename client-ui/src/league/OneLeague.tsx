import * as React from 'react';
import {Card, Space, Tag} from "antd";
import {League, LeagueForm} from "./LeagueForm";
import {LabeledValue} from "../common/LabeledValue";
import {useDeleteLeagueMutation, useGetFacilityByIdQuery, useGetWeekByIdQuery} from "../redux/api.js";
import {DeleteConfirm} from "../common/DeleteConfirm";
import {ColorDisplay, TeamColor} from "./ColorSelect";


export function OneLeague({league_id, name, facility, weeks, start_time, commissioner, colors}: League) {

    const [deleteLeague] = useDeleteLeagueMutation()

    const deleteSelf = () => {
        deleteLeague(league_id).then((res: { error: any, data: any }) => {
            if (res.error) {
                console.log(res.error)
            } else if (res.data) {
                console.log(res.error)
            }
        })
    }

    const l: League = {
        commissioner, facility, name, start_time, weeks, league_id, colors
    }

    const editForm = <LeagueForm LeagueId={league_id} InitialState={l} Update/>

    const extra = <Space>
        {editForm}
        <DeleteConfirm deleteFunction={deleteSelf} objectType={"facility"}/>
    </Space>

    return <Card size={"small"} title={name} extra={extra}>
        <LabeledValue label={"Start time"} value={start_time} key={"start"} vertical/>
        <LabeledValue label={"Facility"} value={<ShortFacility facilityId={facility}/>} vertical key={"facility"}/>
        <WeeksInLeague Weeks={weeks}/>
        <ColorsInLeague Colors={colors}/>
    </Card>
}

type WeeksProps = {
    Weeks: string[]
}

function WeeksInLeague({Weeks}: WeeksProps) {
    const weeks = Weeks?.map((w) => <Week weekId={w} key={w}/>)
    return <LabeledValue label={"Weeks"} value={weeks} vertical/>
}

type WeekProps = {
    weekId: string
}

function Week({weekId}: WeekProps) {
    const {data} = useGetWeekByIdQuery(weekId)
    return <Tag>
        {data?.resource?.date}
    </Tag>
}


type ShortFacilityProps = {
    facilityId: string
}

function ShortFacility({facilityId}: ShortFacilityProps) {
    const {data} = useGetFacilityByIdQuery(facilityId)
    return <span>
        {data?.resource?.name}
    </span>
}

type ColorsInLeagueProps = {
    Colors: TeamColor[]
}

function ColorsInLeague({Colors}: ColorsInLeagueProps) {

    const colors = Colors?.map((c) => <ColorDisplay name={c.name} hex={c.hex} key={c.hex}/>)
    const value = colors?.length ? colors : <Tag>None</Tag>

    const wrappedValue = <Space style={{display: "flex", flexWrap: "wrap"}}>
        {value}
    </Space>

    return <LabeledValue label={"Team colors"} value={wrappedValue}/>
}
