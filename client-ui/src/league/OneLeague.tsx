import * as React from 'react';
import {Card, Empty} from "antd";
import {TeamColorProps} from "../team/TeamColor";


export type OneLeagueProps = {
    league_id: string,
    name: string,
    colors: TeamColorProps[],
}

export function OneLeague({name, colors}: OneLeagueProps) {
    return <Card size={"small"} title={name}>
        <Empty/>
    </Card>
}