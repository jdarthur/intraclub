import * as React from 'react';
import {Card, Tag} from "antd";
import {League} from "./LeagueForm";
import {LabeledValue} from "../common/LabeledValue";
import {useGetFacilityByIdQuery, useGetWeekByIdQuery} from "../redux/api.js";


export function OneLeague({name, facility, weeks, start_time}: League) {

    const weeksInLeague = <WeeksInLeague Weeks={weeks}/>

    return <Card size={"small"} title={name}>
        <LabeledValue label={"Start time"} value={start_time} key={"start"} vertical/>
        <LabeledValue label={"Facility"} value={<ShortFacility facilityId={facility}/>} vertical key={"facility"}/>
        {weeksInLeague}
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
    console.log(data)
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