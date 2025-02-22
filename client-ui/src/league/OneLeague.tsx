import * as React from 'react';
import {Card, Space, Tag} from "antd";
import {League, LeagueForm} from "./LeagueForm";
import {LabeledValue} from "../common/LabeledValue";
import {useDeleteLeagueMutation, useGetFacilityByIdQuery, useGetWeekByIdQuery} from "../redux/api.js";
import {DeleteConfirm} from "../common/DeleteConfirm";
import {ColorDisplay, TeamColor} from "./ColorSelect";
import {ArrowRightOutlined, LoginOutlined} from "@ant-design/icons";
import {Link} from "react-router-dom";


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
        <DeleteConfirm deleteFunction={deleteSelf} objectType={"league"}/>
    </Space>

    const title = <Link to={`/league/${league_id}`}>
        <Space>
            {name}
            <ArrowRightOutlined/>
        </Space>
    </Link>


    return <Card size={"small"} title={title} extra={extra}>
        <LabeledValue label={"Start time"} value={start_time} key={"start"} vertical/>
        <LabeledValue label={"Facility"} value={<ShortFacility facilityId={facility}/>} vertical key={"facility"}/>
        <WeeksInLeague Weeks={weeks}/>
        <ColorsInLeague colors={colors}/>
    </Card>
}

type WeeksProps = {
    Weeks: string[]
}

export function WeeksInLeague({Weeks}: WeeksProps) {
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

export function ShortFacility({facilityId}: ShortFacilityProps) {
    const {data} = useGetFacilityByIdQuery(facilityId)
    return <span>
        {data?.resource?.name}
    </span>
}

type ColorsInLeagueProps = Partial<Pick<League, "colors">>

export function ColorsInLeague({colors}: ColorsInLeagueProps) {

    const c = colors?.map((color) => <ColorDisplay name={color.name} hex={color.hex} key={color.hex}/>)
    const value = c?.length ? c : <Tag>None</Tag>

    const wrappedValue = <Space style={{display: "flex", flexWrap: "wrap"}}>
        {value}
    </Space>

    return <LabeledValue label={"Team colors"} value={wrappedValue}/>
}
