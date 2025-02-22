import {Card, Empty} from "antd";
import * as React from "react";
import {useGetTeamsByUserIdQuery} from "../redux/api.js";
import {USER_PAGE_WIDTH} from "./UserPage";

type Team = {
    TeamId: string
    TeamName: string
    IsActive: boolean
}

export function TeamsForUser({UserId}: ByUserId) {

    const {data} = useGetTeamsByUserIdQuery(UserId)

    let content = <Empty/>
    if (data?.resource?.length > 0) {
        content = data?.resource?.map((team: Team) => <div>
            {team.TeamId}
        </div>)
    }

    return <Card title={"Teams"} style={{width: USER_PAGE_WIDTH}}>
        {content}
    </Card>
}