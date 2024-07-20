import * as React from "react"
import {isMatchWon} from "./SetScores";
import {MatchupProps} from "./Matchup";
import {
    CheckCircleFilled,
    CloseCircleFilled,
} from "@ant-design/icons";
import {Space} from "antd";

type WonLostTopLineProps = {
    Matchups: MatchupProps[]
    Home?: boolean
}

export function WonLostTopLine({Matchups, Home}: WonLostTopLineProps) {

    let wonCount = 0
    let lostCount = 0

    for (let i = 0; i < Matchups?.length; i++) {
        const m = Matchups[i]
        const us = Home ? m.Result.Us : m.Result.Them
        const them = Home ? m.Result.Them : m.Result.Us

        if (isMatchWon(us, them)) {
            wonCount += 1
        } else if (isMatchWon(them, us)) {
            lostCount += 1
        }
    }

    return <Space>
        <Space>
            {wonCount}
            <CheckCircleFilled style={{}}/>
        </Space>
        <Space>
            {lostCount}
            <CloseCircleFilled/>
        </Space>

    </Space>


}