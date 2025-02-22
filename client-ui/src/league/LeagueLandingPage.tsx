import {NavigationBreadcrumb} from "../navigation/NavigationBreadcrumb";
import * as React from "react";
import {Link, useParams} from "react-router-dom";
import {useGetLeagueQuery} from "../redux/api.js";
import {League} from "./LeagueForm";
import {ColorsInLeague, ShortFacility, WeeksInLeague} from "./OneLeague";
import {Card, Space, Tabs} from "antd";
import {LabeledValue} from "../common/LabeledValue";
import {TeamColor} from "./ColorSelect";

const {TabPane} = Tabs

export function LeagueLandingPage() {
    const {id} = useParams()
    const {data} = useGetLeagueQuery(id)

    const league: League = {
        colors: data?.resource?.colors,
        commissioner: data?.resource?.commissioner,
        facility: data?.resource?.facility,
        league_id: id,
        name: data?.resource?.name,
        start_time: data?.resource?.start_time,
        weeks: data?.resource?.weeks,
    }

    return <div>
        <NavigationBreadcrumb items={[<Link to={"/leagues"}>Leagues</Link>, league.name]}/>
        <div style={{width: 500}}>
            <LeagueInfo facility={league.facility} start_time={league.start_time} weeks={league.weeks}/>
            <div style={{height: "1em"}}/>
            <TeamInfo colors={league.colors} league_id={league.league_id}/>
        </div>
    </div>
}

type LeagueInfoProps = Partial<Pick<League, "start_time" | "facility" | "weeks">>;

function LeagueInfo({start_time, facility, weeks}: LeagueInfoProps) {
    return <Card title={"League info"} style={{width: 400}}>
        <div style={{fontSize: "1.2em"}}>
            <LabeledValue vertical label={"Start time"} value={start_time}/>
            <LabeledValue vertical label={"Facility"} value={<ShortFacility facilityId={facility}/>}/>
            <WeeksInLeague Weeks={weeks}/>
        </div>
    </Card>
}

type TeamInfoProps = Partial<Pick<League, "colors" | "league_id">>

function TeamInfo({colors, league_id}: TeamInfoProps) {
    return <Card title={"Team info"} style={{width: 400}}>
        <div style={{fontSize: "1.2em"}}>
            <TeamCaptains colors={colors}/>
        </div>
    </Card>
}

function TeamCaptains({colors, league_id}: TeamInfoProps) {
    const panes = colors?.map((c) => {
        const name = <Space>
            <SmallColor hex={c.hex}/>
            {c.name}
        </Space>
        return <TabPane tab={name} key={c.hex}>
            <LabeledValue label={"Captain"} value={"not set"}/>
        </TabPane>
    })

    return <Tabs>
        {panes}
    </Tabs>
}

type smallColorProps = Partial<Pick<TeamColor, "hex">>

function SmallColor({hex}: smallColorProps) {
    return <div style={{
        width: "1em",
        height: "1em",
        borderRadius: 5,
        backgroundColor: `#${hex}`,
        border: "1px solid #d9d9d9"
    }}/>
}