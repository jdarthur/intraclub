import * as React from 'react';
import {useGetLeaguesByUserIdQuery} from "../redux/api.js";
import {Empty} from "antd";
import {useToken} from "../redux/auth.js";
import {OneLeague} from "./OneLeague";
import {League} from "./LeagueForm";

export function LeaguesForUser({UserId}: ByUserId) {

    const token = useToken()
    const {data} = useGetLeaguesByUserIdQuery(UserId, {skip: !token})

    const leagues = data?.resource.map((league: League) => {
        return <OneLeague league_id={league.league_id}
                          name={league.name}
                          colors={league.colors}
                          commissioner={league.commissioner}
                          facility={league.facility}
                          start_time={league.start_time}
                          weeks={league.weeks}/>
    })

    return <div style={{display: "flex", flexWrap: 'wrap', paddingTop: "1em"}}>
        {data?.resource?.length ? leagues : <Empty/>}
    </div>

}