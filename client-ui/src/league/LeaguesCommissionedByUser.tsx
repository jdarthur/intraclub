import * as React from 'react';
import {useGetLeaguesCommissionedByUserIdQuery} from "../redux/api.js";
import {Button, Empty} from "antd";
import {useToken} from "../redux/auth.js";
import {OneLeague, OneLeagueProps} from "./OneLeague";
import {PlusSquareOutlined} from "@ant-design/icons";
import {LeagueForm} from "./LeagueForm";


export function LeaguesCommissionedByUser({UserId}: ByUserId) {

    const token = useToken()
    const {data} = useGetLeaguesCommissionedByUserIdQuery(UserId, {skip: !token})
    console.log(data)

    const leagues = data?.resource.map((league: OneLeagueProps) => {
        return <OneLeague league_id={league.league_id}
                          name={league.name}
                          colors={league.colors}/>
    })

    return <div>
        <div style={{display: "flex", flexWrap: 'wrap', paddingTop: "1em"}}>
            {data?.resource?.length ? leagues : <Empty/>}
        </div>
        <LeagueForm/>
    </div>

}