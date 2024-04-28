import * as React from 'react';
import {useGetLeaguesByUserIdQuery} from "../redux/api";
import {Empty} from "antd";
import {useToken} from "../redux/auth";
import {OneLeague, OneLeagueProps} from "./OneLeague";

export function LeaguesForUser({UserId}: ByUserId) {

    const token = useToken()
    const {data} = useGetLeaguesByUserIdQuery(UserId, {skip: !token})
    console.log(data)

    const leagues = data?.resource.map((league: OneLeagueProps) => {
        return <OneLeague league_id={league.league_id}
                          name={league.name}
                          colors={league.colors}/>
    })

    return <div style={{display: "flex", flexWrap: 'wrap', paddingTop: "1em"}}>
        {data?.resource?.length ? leagues : <Empty/>}
    </div>

}