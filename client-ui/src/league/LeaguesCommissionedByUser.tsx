import * as React from 'react';
import {useGetLeaguesCommissionedByUserIdQuery} from "../redux/api.js";
import {Empty} from "antd";
import {useToken} from "../redux/auth.js";
import {OneLeague} from "./OneLeague";
import {League, LeagueForm} from "./LeagueForm";


export function LeaguesCommissionedByUser({UserId}: ByUserId) {

    const token = useToken()
    const {data} = useGetLeaguesCommissionedByUserIdQuery(UserId, {skip: !token})

    const leagues = data?.resource?.map((league: League) => {
        return <OneLeague league_id={league.league_id}
                          name={league.name}
                          colors={league.colors}
                          facility={league.facility}
                          weeks={league.weeks}
                          commissioner={league.commissioner}
                          start_time={league.start_time}
                          key={league.league_id}
        />
    })

    return <div>
        <div style={{display: "flex", flexWrap: 'wrap', paddingTop: "1em"}}>
            {data?.resource?.length ? leagues : <Empty/>}
        </div>
        <LeagueForm/>
    </div>

}